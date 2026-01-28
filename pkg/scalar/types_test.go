package scalar

import (
	"reflect"
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	tests := []struct {
		name      string
		input     Options
		expected  Options
	}{
		{
			name:  "defaults",
			input: Options{},
			expected: Options{
				CDN:    DefaultCDN,
				Layout: LayoutModern,
			},
		},
		{
			name: "cdn-only",
			input: Options{
				CDN: "https://example.com/cdn",
			},
			expected: Options{
				CDN:    "https://example.com/cdn",
				Layout: LayoutModern,
			},
		},
		{
			name: "layout-only",
			input: Options{
				Layout: LayoutClassic,
			},
			expected: Options{
				CDN:    DefaultCDN,
				Layout: LayoutClassic,
			},
		},
		{
			name: "keep",
			input: Options{
				CDN:           "https://example.com/cdn",
				Layout:        LayoutClassic,
				Theme:         ThemeMars,
				SpecURL:       "spec.json",
				ShowSidebar:   true,
				HiddenClients: []string{"go", "ruby"},
				CustomOptions: CustomOptions{
					PageTitle: "Docs",
				},
			},
			expected: Options{
				CDN:           "https://example.com/cdn",
				Layout:        LayoutClassic,
				Theme:         ThemeMars,
				SpecURL:       "spec.json",
				ShowSidebar:   true,
				HiddenClients: []string{"go", "ruby"},
				CustomOptions: CustomOptions{
					PageTitle: "Docs",
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			inputCopy := tc.input
			got := DefaultOptions(tc.input)
			if got == nil {
				t.Fatalf("expected non-nil result")
			}
			if &tc.input == got {
				t.Fatalf("expected returned pointer to differ from input address")
			}
			if !reflect.DeepEqual(*got, tc.expected) {
				t.Fatalf("unexpected result: got %+v, want %+v", *got, tc.expected)
			}
			if !reflect.DeepEqual(tc.input, inputCopy) {
				t.Fatalf("input mutated: got %+v, want %+v", tc.input, inputCopy)
			}
		})
	}
}
