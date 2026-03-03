package utils

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/utils/diff"
	"github.com/sergi/go-diff/diffmatchpatch"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	colourReset   = "\x1b[0m"
	colourRed     = "\x1b[91m"
	colourGreen   = "\x1b[92m"
	colourYellow  = "\x1b[93m"
	colourMagenta = "\x1b[95m"
	colourCyan    = "\x1b[96m"
)

// Diff returns a string containing the diff between the two strings.
// This is equivalent to running `git diff` on the two strings.
func Diff(a, b []byte, fileName string) string {
	diffs := diff.Do(string(a), string(b))

	if len(diffs) > 1 {
		return prettyPrintDiff(diffs, fileName)
	}

	return ""
}

// prettyPrintDiff prints the diff for the file out in a style similar to git diff.
// It prints up to 3 lines before and after a change and divides the diff into blocks
// each with it's own header identifying the line numbers and number of changes.
func prettyPrintDiff(diffs []diffmatchpatch.Diff, fileName string) string {
	diffLines := splitDiffsToLines(diffs)
	changedLines := getChangedLines(diffLines)
	filteredLines := filterLines(diffLines, changedLines)
	diffBlocks := splitLinesIntoBlocks(filteredLines)

	buf := bytes.NewBuffer(nil)

	for _, block := range diffBlocks {
		printDiffBlock(buf, block, fileName)
	}

	return buf.String()
}

// diffLine holds the line number and text for a line in a diff.
type diffLine struct {
	number   int
	text     string
	diffType diffmatchpatch.Operation
}

// splitDiffsToLines splits the diffs into lines and adds the line number.
// Deleted lines do not increase the line number as they are then replaced
// by the next line.
// Line numbers should be accurate for the new file.
func splitDiffsToLines(in []diffmatchpatch.Diff) []diffLine {
	var out []diffLine

	lineNumber := 1
	for _, d := range in {
		data := strings.TrimSuffix(d.Text, "\n")
		lines := strings.Split(data, "\n")

		for _, line := range lines {
			out = append(out, diffLine{
				number:   lineNumber,
				text:     line,
				diffType: d.Type,
			})

			// Once we've used a line number, increment it unless it's a delete line.
			if d.Type != diffmatchpatch.DiffDelete {
				lineNumber++
			}
		}
	}

	return out
}

// getChangedLines returns a set of line numbers that have changed.
func getChangedLines(lines []diffLine) sets.Int {
	changeSet := sets.NewInt()

	for _, line := range lines {
		if line.diffType != diffmatchpatch.DiffEqual {
			changeSet.Insert(line.number)
		}
	}

	return changeSet
}

// filterLines removes lines that are not changed and are not within 3 lines of a changed line.
func filterLines(lines []diffLine, changedLines sets.Int) []diffLine {
	var out []diffLine

	linesAhead := 0
	for i, line := range lines {
		if changedLines.Has(line.number) {
			// This line changed, pick it.
			out = append(out, line)
			// We need to pick at least the next 3 lines.
			linesAhead = 3
			continue
		}

		if linesAhead > 0 {
			// We need to pick this line.
			out = append(out, line)
			linesAhead--
			continue
		}

		// This line didn't change so we don't need to pick it unless there is a change ahead.
		for j := 1; j <= 3; j++ {
			if i+j >= len(lines) {
				break
			}

			if changedLines.Has(lines[i+j].number) {
				// There is a change ahead, pick this line.
				out = append(out, line)
				// Pick at least the lines up to the next change.
				linesAhead = j
			}
		}
	}

	return out
}

// splitLinesIntoBlocks splits the lines into contiguous blocks of lines that have changed.
func splitLinesIntoBlocks(lines []diffLine) [][]diffLine {
	var out [][]diffLine

	var block []diffLine
	for i, line := range lines {
		// Every time the line number jumps, switch to a new block.
		if i == 0 || (line.number != lines[i-1].number && line.number != lines[i-1].number+1) {
			if len(block) > 0 {
				out = append(out, block)
			}
			block = []diffLine{}
		}

		block = append(block, line)
	}

	// Make sure to append the final block.
	out = append(out, block)

	return out
}

// printDiffBlock prints a block of lines with a header identifying the line numbers and number of changes.
func printDiffBlock(buf *bytes.Buffer, block []diffLine, fileName string) {
	first, last := block[0], block[len(block)-1]
	inserts, deletes := countChangesIn(block)

	blockHeaderFmt := "%s ----- %s [%d - %d] +%d -%d -----%s\n"
	blockHeader := fmt.Sprintf(blockHeaderFmt, colourCyan, fileName, first.number, last.number, inserts, deletes, colourReset)
	buf.WriteString(blockHeader)

	for _, line := range block {
		printDiffLine(buf, line)
	}
}

// printDiffLine prints a line with the appropriate colour for the diff type.
func printDiffLine(buf *bytes.Buffer, line diffLine) {
	var prefix, suffix string

	switch line.diffType {
	case diffmatchpatch.DiffInsert:
		prefix = fmt.Sprintf("%s+%d ", colourGreen, line.number)
		suffix = colourReset
	case diffmatchpatch.DiffDelete:
		prefix = fmt.Sprintf("%s-%d ", colourRed, line.number)
		suffix = colourReset
	default:
		prefix = fmt.Sprintf(" %d ", line.number)
	}

	_, _ = buf.WriteString(prefix + line.text + suffix + "\n")
}

// countChangesIn counts the number of insertions and deletions in a block of lines.
func countChangesIn(block []diffLine) (int, int) {
	var inserts, deletes int

	for _, line := range block {
		switch line.diffType {
		case diffmatchpatch.DiffInsert:
			inserts++
		case diffmatchpatch.DiffDelete:
			deletes++
		}
	}

	return inserts, deletes
}
