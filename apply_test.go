package v4a

import "testing"

func TestStrictUpdates(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, input, diff, want string
		failed                  bool
	}{
		{"exact", "a\nb\n", "@@\n-a\n+c", "c\nb\n", false},
		{"ambiguous", "same\nsame\n", "@@\n-same\n+new", "", true},
		{"missing anchor", "a", "@@ missing\n-a\n+b", "", true},
		{"repeated anchor", "h\na\nh\nb", "@@ h\n-a\n+c", "", true},
		{"anchor narrows", "a\nheader\na", "@@ header\n-a\n+b", "a\nheader\nb", false},
		{"trailing space mismatch", "a \n", "-a\n+b", "", true},
		{"leading space mismatch", " a\n", "-a\n+b", "", true},
		{"unicode exact", "café é 😀\n", "-café é 😀\n+café é 🙂", "café é 🙂\n", false},
		{"unicode normalization mismatch", "é", "-é\n+b", "", true},
		{"CRLF", "a\r\nb\r\n", "-a\n+c", "c\r\nb\r\n", false},
		{"no final newline", "a", "-a\n+b\n", "b", false},
		{"EOF", "a\n", " a\n+b\n*** End of File", "a\nb\n", false},
		{"EOF no fallback", "a\nb\n", "-a\n+c\n*** End of File", "", true},
		{"EOF append", "a\n", "+b\n*** End of File", "a\nb\n", false},
		{"unanchored insertion", "a", "+b", "", true},
		{"empty input", "", "+a", "a", false},
		{"empty input has no line to delete", "", "@@\n-\n+a", "", true},
		{"empty input has no context line", "", "@@\n \n+a", "", true},
		{"empty input has no empty anchor", "", "@@ \n+a", "", true},
		{"empty input EOF insertion", "", "@@\n+a\n*** End of File", "a", false},
		{"existing blank line replacement", "\n", "@@\n-\n+a\n*** End of File", "a\n", false},
		{"mixed newlines", "a\r\nb\n", "-a\n+c", "", true},
		{"header", "a", "*** Update File: x", "", true},
		{"marker suffix", "a", "*** End of File garbage", "", true},
		{"after EOF", "a", "-a\n+b\n*** End of File\n+c", "", true},
		{"invalid hunk", "a", "@@invalid\n-a\n+b", "", true},
		{"empty hunk", "a", "@@", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := Apply(tc.input, tc.diff)
			if (err != nil) != tc.failed || got != tc.want {
				t.Fatalf("got %q, %v; want %q, failure=%v", got, err, tc.want, tc.failed)
			}
		})
	}
}

func TestFrozenBatch(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		diffs  []string
		want   string
		failed bool
	}{
		{"independent", []string{"-a\n+x", "-b\n+y"}, "x\ny\n", false},
		{"overlap", []string{"-a\n+x", "-a\n+y"}, "", true},
		{"sequential dependency", []string{"-a\n+x", "-x\n+y"}, "", true},
		{"late invalid", []string{"-a\n+x", "-missing\n+y"}, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := ApplyBatch("a\nb\n", tc.diffs)
			if (err != nil) != tc.failed || got != tc.want {
				t.Fatalf("got %q, %v", got, err)
			}
		})
	}
}
