package basicUnitTest

import "testing"

func TestAdd(t *testing.T) {
	got := add(2, 3)

	if got != 5 {
		t.Errorf("got %d, want %d", got, 5)
	}
}
