package main

import (
	"errors"
	"fmt"
	"testing"
)

func TestWrappedErrorPreservesIdentity(t *testing.T) {
	err := fmt.Errorf("get user: %w", ErrUserNotFound)

	if !errors.Is(err, ErrUserNotFound) {
		t.Fatal("expected errors.Is to recognize ErrUserNotFound")
	}
}

func TestAsFindsTypedError(t *testing.T) {
	err := fmt.Errorf(
		"request rejected: %w",
		&ValidationError{
			Field:  "id",
			Reason: "must be greater than zero",
		},
	)

	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatal("expected errors.As to find ValidationError")
	}

	if validationErr.Field != "id" {
		t.Fatalf("unexpected field: %q", validationErr.Field)
	}
}

func TestPercentVBreaksIdentity(t *testing.T) {
	err := fmt.Errorf("get user: %v", ErrUserNotFound)

	if errors.Is(err, ErrUserNotFound) {
		t.Fatal("expected %v formatting not to preserve the wrapped identity")
	}
}
