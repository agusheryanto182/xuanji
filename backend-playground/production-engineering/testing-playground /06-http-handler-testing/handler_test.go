package httphandlertesting

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeUserService struct {
	user User
	err  error
}

func (f *fakeUserService) GetUser(id int) (User, error) {
	return f.user, f.err
}

func TestHandler_GetUser(t *testing.T) {
	service := &fakeUserService{
		user: User{
			ID:    1,
			Name:  "Agus",
			Email: "agus@example.com",
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users?id=1",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.GetUser(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusOK)
	}

	var got User

	err := json.NewDecoder(rec.Body).Decode(&got)
	if err != nil {
		t.Fatal(err)
	}

	if got.ID != 1 {
		t.Errorf("got ID %d, want 1", got.ID)
	}

	if got.Name != "Agus" {
		t.Errorf("got name %s, want Agus", got.Name)
	}
}

func TestHandler_GetUser_NotFound(t *testing.T) {
	service := &fakeUserService{
		err: errors.New("user not found"),
	}

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users?id=999",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.GetUser(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf(
			"got status %d, want %d",
			rec.Code,
			http.StatusNotFound,
		)
	}
}
