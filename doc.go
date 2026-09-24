// Package v4a applies V4A update diffs strictly.
//
// V4A is the context-anchored patch format that OpenAI models emit through
// the apply_patch tool. An update diff looks like this:
//
//	@@ def greet():
//	-    print("Hi")
//	+    print("Hello")
//
// Lines starting with a space are context, "-" lines are removed, and "+"
// lines are added. "@@ text" moves the search past the one input line equal
// to text; a bare "@@" starts a new section. "*** End of File" pins the
// section before it to the end of the input.
//
// Unlike the reference implementations, this package never guesses. Anchors,
// context, and removed lines must match the input byte for byte and exactly
// once: there is no whitespace trimming or Unicode normalization, and an
// insertion without context needs an anchor or "*** End of File". A diff that
// does not apply returns an error and no partial result, so the caller can
// send the error back to the model and ask for a new patch.
//
// Only update diffs are supported. The "*** Begin Patch" envelope and the
// Add File, Delete File, and Move to headers are rejected; the caller handles
// those operations.
package v4a
