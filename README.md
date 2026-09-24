# v4a

[![Go Reference](https://pkg.go.dev/badge/github.com/fgn/v4a.svg)](https://pkg.go.dev/github.com/fgn/v4a)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A strict Go applier for V4A diffs, the patch format OpenAI models emit through
the [`apply_patch`](https://developers.openai.com/api/docs/guides/tools-apply-patch)
tool. It is a Go port of
[`applyDiff`](https://github.com/openai/openai-agents-js/blob/main/packages/agents-core/src/utils/applyDiff.ts)
from the [OpenAI Agents SDK for JavaScript](https://github.com/openai/openai-agents-js),
changed to fail closed instead of guessing.

## Install

```sh
go get github.com/fgn/v4a
```

## Usage

```go
input := "def greet():\n    print(\"Hi\")\n"
diff := "@@ def greet():\n-    print(\"Hi\")\n+    print(\"Hello\")"

output, err := v4a.Apply(input, diff, "update")
```

`ApplyBatch` applies several diffs for one file against the same original
input and fails as a whole if any diff fails or two diffs overlap.

## Behavior

- Anchors, context, and removed lines must match byte for byte and exactly
  once. There is no whitespace trimming or Unicode normalization.
- An insertion without context needs an anchor, `*** End of File`, or an
  empty input.
- Only `update_file` diffs are supported. Creating, deleting, and moving
  files is left to the caller.
- LF and CRLF newlines and a missing final newline are preserved.

## Development

Development tasks run through [Task](https://taskfile.dev), which you can
install from [taskfile.dev/docs/installation](https://taskfile.dev/docs/installation).
Run `task` to format, lint, and test, and `task --list` to see the rest.

## License

MIT, see [LICENSE](LICENSE). Based on the OpenAI Agents SDK for JavaScript,
copyright (c) 2025 OpenAI, also MIT.
