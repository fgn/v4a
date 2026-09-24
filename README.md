# v4a

[![Go Reference](https://pkg.go.dev/badge/github.com/fgn/v4a.svg)](https://pkg.go.dev/github.com/fgn/v4a)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A strict Go applier for V4A update diffs: the context-anchored patch format
OpenAI models emit through the `apply_patch` tool. Give it a file's text and
the `diff` of an `update_file` operation; get the patched text or an error
that explains why the patch does not apply. No dependencies outside the
standard library and no filesystem access.

The reference implementations forgive a lot: they take the first match,
ignore whitespace differences, and in some cases normalize Unicode
punctuation. This package fails closed instead. Every anchor, context line,
and removed line must match byte for byte and exactly once, so a patch either
lands where the model meant it or is rejected and can be sent back to the
model for another attempt.

## Install

```sh
go get github.com/fgn/v4a
```

## Usage

```go
input := "def greet():\n    print(\"Hi\")\n"
diff := "@@ def greet():\n-    print(\"Hi\")\n+    print(\"Hello\")"

output, err := v4a.Apply(input, diff, "update")
if err != nil {
    // Return err to the model as the apply_patch_call_output.
}
```

`ApplyBatch` applies several diffs for the same file as one edit. Each diff
targets the original input, not the output of the diffs before it, and the
batch fails as a whole when any diff fails or two diffs overlap:

```go
output, err := v4a.ApplyBatch(input, []string{diffA, diffB})
```

## Format

```text
@@ class Greeter:
@@     def greet(self):
         name = "world"
-        print("Hi")
+        print("Hello, " + name)
@@
         return None
+# done
*** End of File
```

- A line starting with a space is context, `-` removes a line, and `+` adds
  one. An empty line is an empty context line.
- `@@ text` moves the search past the one input line equal to `text`,
  including its indentation. Several anchors in a row are applied in order. A
  bare `@@` starts a new section.
- `*** End of File` pins the section before it to the end of the input and
  must be the last line of the diff.

## Rules

- Only update diffs. `Apply` rejects any mode other than `"update"`, and any
  `***` line other than `*** End of File`, including the `*** Begin Patch`
  envelope and the Add File, Delete File, and Move to headers. Creating,
  deleting, and moving files is left to the caller.
- An anchor must equal exactly one line in the whole input, and that line
  must come after the previous section. Anchors do not disambiguate each
  other: a method name that appears in two classes cannot be an anchor.
- A section's context and removed lines must match exactly once in the rest of
  the input. Nothing is trimmed or normalized.
- An insertion with no context needs an anchor, `*** End of File`, or an empty
  input. A bare `@@` followed only by `+` lines is rejected, because there is
  no agreed place to put them.
- Input with LF or CRLF newlines keeps them, and a missing final newline stays
  missing. Mixed newlines and bare CR are rejected.
- A diff with no changes, such as an empty or context-only diff, succeeds and
  returns the input unchanged. If every call must change something, compare
  the output with the input.

## Differences from other implementations

| Behavior | This package | OpenAI Agents SDK `applyDiff` | Codex `apply_patch` |
| --- | --- | --- | --- |
| Ambiguous context | Error | First match | First match |
| Whitespace differences | Error | Ignores trailing, then surrounding whitespace | Ignores trailing, then surrounding whitespace |
| Unicode dashes, quotes, and non-breaking spaces | Error | Exact | Normalized |
| `*** End of File` context not at the end | Error | Searched anywhere | Error |
| `@@` followed only by `+` lines | Error | Inserted at the cursor (top of the file for the first section) | Appended at the end |
| Final newline | Preserved | Preserved | Always added |
| Empty update diff | No change | No change | Error |

## Development

The [Taskfile](Taskfile.yml) installs its own tools:

```sh
task          # format, lint, and test
task test     # tests with the race detector
task fuzz     # fuzz the parser; FUZZTIME=5m for longer
task lint     # golangci-lint, govulncheck, go mod tidy, markdownlint
```

## License

MIT. Adapted from `applyDiff` in the
[OpenAI Agents SDK for JavaScript](https://github.com/openai/openai-agents-js),
copyright (c) 2025 OpenAI; see [LICENSE](LICENSE).
