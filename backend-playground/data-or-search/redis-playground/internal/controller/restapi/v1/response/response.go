package response

const (
	StatusSuccess = "success"

	StatusError = "error"
)

type Response[T any] struct {
	Status string `json:"status"` // "success" | "error"
	Data   T      `json:"data,omitempty"`
	Meta   *Meta  `json:"meta,omitempty"`
	Error  *Error `json:"error,omitempty"`
}

type Meta struct {
	Total  int `json:"total,omitempty"`
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}
