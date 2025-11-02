package matcher

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		wantErr  bool
	}{
		{
			name:     "valid patterns",
			patterns: []string{"hello", "world.*", "[0-9]+"},
			wantErr:  false,
		},
		{
			name:     "invalid pattern",
			patterns: []string{"[invalid"},
			wantErr:  true,
		},
		{
			name:     "empty patterns",
			patterns: []string{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.patterns)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMatcher_Match(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		text     string
		want     bool
	}{
		{
			name:     "exact match",
			patterns: []string{"hello"},
			text:     "hello",
			want:     true,
		},
		{
			name:     "regex match",
			patterns: []string{"hello.*"},
			text:     "hello world",
			want:     true,
		},
		{
			name:     "no match",
			patterns: []string{"hello"},
			text:     "world",
			want:     false,
		},
		{
			name:     "multiple patterns one match",
			patterns: []string{"hello", "world"},
			text:     "world",
			want:     true,
		},
		{
			name:     "numeric pattern",
			patterns: []string{"[0-9]+"},
			text:     "12345",
			want:     true,
		},
		{
			name:     "case sensitive",
			patterns: []string{"hello"},
			text:     "Hello",
			want:     true,
		},
		{
			name:     "text with special chars",
			patterns: []string{"helloworld"},
			text:     "Hello~World!",
			want:     true,
		},
		{
			name:     "text with apostrophe",
			patterns: []string{"dont"},
			text:     "Don't",
			want:     true,
		},
		{
			name:     "text with tilde",
			patterns: []string{"cafe"},
			text:     "Café~",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := New(tt.patterns)
			if err != nil {
				t.Fatalf("Failed to create matcher: %v", err)
			}
			if got := m.Match(tt.text); got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatcher_FindMatches(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		text     string
		want     int
	}{
		{
			name:     "no matches",
			patterns: []string{"hello", "world"},
			text:     "goodbye",
			want:     0,
		},
		{
			name:     "one match",
			patterns: []string{"hello", "world"},
			text:     "hello",
			want:     1,
		},
		{
			name:     "multiple matches",
			patterns: []string{"h.*o", "hel.*"},
			text:     "hello",
			want:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := New(tt.patterns)
			if err != nil {
				t.Fatalf("Failed to create matcher: %v", err)
			}
			got := m.FindMatches(tt.text)
			if len(got) != tt.want {
				t.Errorf("FindMatches() returned %d matches, want %d", len(got), tt.want)
			}
		})
	}
}
