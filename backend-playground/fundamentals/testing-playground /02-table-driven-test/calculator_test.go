package tabledriventest

import "testing"

func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{
			name: "1 + 1 = 2",
			a:    1,
			b:    1,
			want: 2,
		},
		{
			name: "2 + 2 = 4",
			a:    2,
			b:    2,
			want: 4,
		},
		{
			name: "3 + 3 = 6",
			a:    3,
			b:    3,
			want: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := add(tt.a, tt.b)

			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}
