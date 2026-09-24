package v4a

import (
	"strings"
	"testing"
)

func FuzzApply(f *testing.F) {
	for _, seed := range [][2]string{
		{"a\nb\n", "@@\n-a\n+c"},
		{"a\nheader\na", "@@ header\n-a\n+b"},
		{"a\n", " a\n+b\n*** End of File"},
		{"a\nb\nc\n", "@@\n a\n-b\n+B\n@@\n c\n+d\n*** End of File"},
		{"", "@@\n+a\n*** End of File"},
		{"\n", "@@\n-\n+a\n*** End of File"},
		{"a\r\nb\r\n", "-a\n+c"},
	} {
		f.Add(seed[0], seed[1])
	}
	f.Fuzz(func(t *testing.T, input, diff string) {
		output, err := Apply(input, diff, "update")
		if err != nil && output != "" {
			t.Fatalf("error %v returned output %q", err, output)
		}
		if strings.Contains(input+diff, "\r") || !strings.Contains(input, "\n") {
			return
		}
		// CRLF input must behave exactly like LF input and keep its newlines.
		crlf, crlfErr := Apply(strings.ReplaceAll(input, "\n", "\r\n"), diff, "update")
		if (err == nil) != (crlfErr == nil) || crlf != strings.ReplaceAll(output, "\n", "\r\n") {
			t.Fatalf("LF gave %q, %v; CRLF gave %q, %v", output, err, crlf, crlfErr)
		}
	})
}
