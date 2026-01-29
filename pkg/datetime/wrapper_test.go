package datetime

import (
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}

	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:  "utc",
			input: "2023-01-02T01:02:03Z",
			want:  time.Date(2023, 1, 2, 8, 2, 3, 0, loc),
		},
		{
			name:  "offset",
			input: "2023-01-02T08:00:00+07:00",
			want:  time.Date(2023, 1, 2, 8, 0, 0, 0, loc),
		},
		{
			name:  "fractional",
			input: "2023-01-02T01:02:03.123Z",
			want:  time.Date(2023, 1, 2, 8, 2, 3, 123000000, loc),
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "nonsense",
			input:   "not-a-time",
			wantErr: true,
		},
		{
			name:    "badmonth",
			input:   "2023-13-01T00:00:00Z",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseTime(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				if !got.IsZero() {
					t.Fatalf("expected zero time, got %v", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Location().String() != "Asia/Jakarta" {
				t.Fatalf("expected location Asia/Jakarta, got %s", got.Location().String())
			}
			if !got.Equal(tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
