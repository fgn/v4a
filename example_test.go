package v4a_test

import (
	"fmt"

	"github.com/fgn/v4a"
)

func ExampleApply() {
	input := "def greet():\n    print(\"Hi\")\n"
	diff := "@@ def greet():\n-    print(\"Hi\")\n+    print(\"Hello\")"
	output, err := v4a.Apply(input, diff, "update")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print(output)
	// Output:
	// def greet():
	//     print("Hello")
}

func ExampleApply_strict() {
	// Whitespace differences are not forgiven.
	_, err := v4a.Apply("  indented\n", "-indented\n+changed", "update")
	fmt.Println(err)
	// Output: target must match exactly once
}

func ExampleApplyBatch() {
	input := "alpha\nbeta\ngamma\n"
	output, err := v4a.ApplyBatch(input, []string{"-alpha\n+ALPHA", "-gamma\n+GAMMA"})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print(output)
	// Output:
	// ALPHA
	// beta
	// GAMMA
}

func ExampleApplyBatch_overlap() {
	// Both diffs target the original input, so they cannot edit the same line.
	_, err := v4a.ApplyBatch("alpha\n", []string{"-alpha\n+one", "-alpha\n+two"})
	fmt.Println(err)
	// Output: overlapping edits
}
