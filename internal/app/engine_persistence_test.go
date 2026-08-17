package app

import "testing"

func TestHistoryCompactionInterval(
	t *testing.T,
) {

	tests := []struct {
		name     string
		limit    int
		expected int
	}{
		{
			name:     "small history",
			limit:    50,
			expected: 10,
		},
		{
			name:     "default history",
			limit:    1000,
			expected: 100,
		},
		{
			name:     "large history",
			limit:    5000,
			expected: 100,
		},
		{
			name:     "unlimited history",
			limit:    0,
			expected: 100,
		},
	}

	for _, test := range tests {

		t.Run(
			test.name,
			func(t *testing.T) {

				got := historyCompactionInterval(
					test.limit,
				)

				if got != test.expected {

					t.Fatalf(
						"expected interval %d, got %d",
						test.expected,
						got,
					)
				}
			},
		)
	}
}
