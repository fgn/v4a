// Copyright (c) 2026 Fredrik Gustafsson
// Copyright (c) 2025 OpenAI
// SPDX-License-Identifier: MIT
//
// A strict, update-only Go port of applyDiff from the OpenAI Agents SDK for
// JavaScript:
// https://github.com/openai/openai-agents-js/blob/main/packages/agents-core/src/utils/applyDiff.ts
// See LICENSE.

package v4a

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

const endFile = "*** End of File"

var endSectionMarkers = []string{endFile}

type chunk struct {
	origIndex          int
	delLines, insLines []string
}
type readSectionResult struct {
	nextContext   []string
	sectionChunks []chunk
	endIndex      int
	eof           bool
}

// Apply applies one V4A update diff to input and returns the patched text.
// Creating, deleting, and moving files is left to the caller.
func Apply(input, diff string) (string, error) {
	return ApplyBatch(input, []string{diff})
}

// ApplyBatch applies several update diffs to the same input as one edit.
// Every diff targets the original input, not the output of the diffs before
// it. Overlapping edits fail the whole batch.
func ApplyBatch(input string, diffs []string) (string, error) {
	if len(diffs) == 0 || slices.IndexFunc(diffs, func(diff string) bool { return diff != "" }) == -1 {
		return input, nil
	}
	newline := "\n"
	normalized := strings.ReplaceAll(input, "\r\n", "\n")
	if strings.Contains(input, "\r\n") {
		if strings.Contains(strings.ReplaceAll(input, "\r\n", ""), "\n") {
			return "", errors.New("mixed newline input")
		}
		newline = "\r\n"
	}
	if strings.Contains(normalized, "\r") {
		return "", errors.New("unsupported newline input")
	}
	var chunks []chunk
	for _, diff := range diffs {
		next, err := parseStrict(normalized, diff)
		if err != nil {
			return "", err
		}
		chunks = append(chunks, next...)
	}
	if len(chunks) == 0 {
		return input, nil
	}
	slices.SortStableFunc(chunks, func(a, b chunk) int { return a.origIndex - b.origIndex })
	for i := 1; i < len(chunks); i++ {
		previous := chunks[i-1]
		if chunks[i].origIndex <= previous.origIndex || chunks[i].origIndex < previous.origIndex+len(previous.delLines) {
			return "", errors.New("overlapping edits")
		}
	}
	return applyChunks(normalized, chunks, newline)
}

func parseStrict(input, diff string) ([]chunk, error) {
	if diff == "" {
		return nil, nil
	}
	lines := strings.Split(strings.ReplaceAll(diff, "\r\n", "\n"), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	source := strings.Split(input, "\n")
	if input == "" {
		// Match applyChunks: an empty file has no line to delete or use as
		// context. A phantom empty line would allow an out-of-bounds deletion.
		source = nil
	}
	var chunks []chunk
	cursor, index := 0, 0
	for index < len(lines) {
		anchored := false
		for index < len(lines) && (lines[index] == "@@" || strings.HasPrefix(lines[index], "@@ ")) {
			anchor := strings.TrimPrefix(lines[index], "@@")
			index++
			if anchor == "" {
				continue
			}
			anchor = strings.TrimPrefix(anchor, " ")
			found, count := -1, 0
			for i, line := range source {
				if line == anchor {
					found = i
					count++
				}
			}
			if count != 1 || found < cursor {
				return nil, errors.New("missing, repeated or stale anchor")
			}
			cursor, anchored = found+1, true
		}
		section, err := readSection(lines, index)
		if err != nil {
			return nil, err
		}
		target, count := -1, 0
		limit := len(source)
		if section.eof && limit > 0 && source[limit-1] == "" {
			limit--
		}
		if len(section.nextContext) == 0 {
			switch {
			case section.eof:
				target, count = limit, 1
			case anchored:
				target, count = cursor, 1
			case input == "":
				target, count = 0, 1
			default:
				return nil, errors.New("insertion needs exact context or anchor")
			}
		} else {
			start, stop := cursor, limit-len(section.nextContext)
			if section.eof {
				start = stop
			}
			if start < cursor {
				return nil, errors.New("stale EOF context")
			}
			for i := start; i <= stop; i++ {
				if slices.Equal(source[i:i+len(section.nextContext)], section.nextContext) {
					target = i
					count++
				}
			}
		}
		if count != 1 || target < cursor {
			return nil, errors.New("target must match exactly once")
		}
		for _, c := range section.sectionChunks {
			c.origIndex += target
			chunks = append(chunks, c)
		}
		cursor, index = target+len(section.nextContext), section.endIndex
		if section.eof && index != len(lines) {
			return nil, errors.New("content after EOF marker")
		}
	}
	return chunks, nil
}

func readSection(lines []string, startIndex int) (readSectionResult, error) {
	var context, delLines, insLines []string
	var sectionChunks []chunk
	mode := "keep"
	index := startIndex
	origIndex := index

	flushChunk := func() {
		sectionChunks = append(sectionChunks, chunk{
			origIndex: len(context) - len(delLines),
			delLines:  append([]string(nil), delLines...),
			insLines:  append([]string(nil), insLines...),
		})
	}
	for index < len(lines) {
		raw := lines[index]
		if strings.HasPrefix(raw, "@@") || slices.Contains(endSectionMarkers, raw) {
			break
		}

		if strings.HasPrefix(raw, "***") {
			return readSectionResult{}, fmt.Errorf("invalid line: %s", raw)
		}

		index++
		lastMode := mode
		line := raw
		if line == "" {
			line = " "
		}
		switch line[0] {
		case '+':
			mode = "add"
		case '-':
			mode = "delete"
		case ' ':
			mode = "keep"
		default:
			return readSectionResult{}, fmt.Errorf("invalid line: %s", line)
		}
		lineContent := line[1:]
		switchingToContext := mode == "keep" && lastMode != mode
		if switchingToContext && (len(delLines) > 0 || len(insLines) > 0) {
			flushChunk()
			delLines = nil
			insLines = nil
		}

		switch mode {
		case "delete":
			delLines = append(delLines, lineContent)
			context = append(context, lineContent)
		case "add":
			insLines = append(insLines, lineContent)
		default:
			context = append(context, lineContent)
		}
	}
	if len(delLines) > 0 || len(insLines) > 0 {
		flushChunk()
	}
	if index < len(lines) && lines[index] == endFile {
		return readSectionResult{context, sectionChunks, index + 1, true}, nil
	}
	if index == origIndex {
		nextLine := ""
		if index < len(lines) {
			nextLine = lines[index]
		}
		return readSectionResult{}, fmt.Errorf("nothing in this section - index=%d %s", index, nextLine)
	}
	return readSectionResult{context, sectionChunks, index, false}, nil
}

func applyChunks(input string, chunks []chunk, newline string) (string, error) {
	origLines := strings.Split(input, "\n")
	if input == "" {
		origLines = nil
	}
	var destLines []string
	cursor := 0
	for _, ch := range chunks {
		if ch.origIndex > len(origLines) {
			return "", fmt.Errorf("applyDiff: chunk.origIndex %d > input length %d", ch.origIndex, len(origLines))
		}
		if cursor > ch.origIndex {
			return "", fmt.Errorf("applyDiff: overlapping chunk at %d (cursor %d)", ch.origIndex, cursor)
		}
		destLines = append(destLines, origLines[cursor:ch.origIndex]...)
		cursor = ch.origIndex
		if len(ch.insLines) > 0 {
			destLines = append(destLines, ch.insLines...)
		}
		cursor += len(ch.delLines)
	}
	destLines = append(destLines, origLines[cursor:]...)
	return strings.Join(destLines, newline), nil
}
