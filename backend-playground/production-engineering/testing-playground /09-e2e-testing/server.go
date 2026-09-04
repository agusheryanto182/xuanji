package e2etesting

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(user User) (User, error) {
	var created User

	err := r.db.QueryRow(
		`INSERT INTO users (name, email)
		 VALUES ($1, $2)
		 RETURNING id, name, email`,
		user.Name,
		user.Email,
	).Scan(
		&created.ID,
		&created.Name,
		&created.Email,
	)

	return created, err
}

func (r *Repository) GetByID(id int) (User, error) {
	var user User

	err := r.db.QueryRow(
		`SELECT id, name, email
		 FROM users
		 WHERE id = $1`,
		id,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
	)

	return user, err
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(user User) (User, error) {
	return s.repo.Create(user)
}

func (s *Service) GetUser(id int) (User, error) {
	return s.repo.GetByID(id)
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	created, err := h.service.CreateUser(user)
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(created)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	idString := strings.TrimPrefix(r.URL.Path, "/users/")

	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetUser(id)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(user)
}

func NewServer(db *sql.DB) http.Handler {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /users", handler.CreateUser)
	mux.HandleFunc("GET /users/{id}", handler.GetUser)

	return mux
}
