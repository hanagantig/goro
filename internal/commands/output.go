package commands

import (
	"fmt"
	"io"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

// actionWidth fits the longest action name, skip-unmarked, so the paths line up.
const actionWidth = len(ActionSkipUnmarked)

func printAction(w io.Writer, f plannedFile) {
	fmt.Fprintf(w, "%-*s %s\n", actionWidth, f.action, f.path)
}

// printDryRun prints the plan and nothing else: one line per file whose action isn't unchanged,
// each overwrite followed by a unified diff from the file in the work dir to the regenerated one.
func printDryRun(w io.Writer, files []plannedFile) error {
	for _, f := range files {
		if f.action == ActionUnchanged {
			continue
		}
		printAction(w, f)

		if f.action != ActionOverwrite {
			continue
		}
		diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:        splitLines(f.current),
			B:        splitLines(f.rendered),
			FromFile: "a/" + f.path,
			ToFile:   "b/" + f.path,
			Context:  3,
		})
		if err != nil {
			return fmt.Errorf("diff %s: %w", f.path, err)
		}
		fmt.Fprint(w, diff)
	}

	return nil
}

// noNewline follows a last line without a newline in a unified diff, as git writes it.
const noNewline = "\n\\ No newline at end of file\n"

// splitLines splits content into the lines of a unified diff, each ending in "\n". difflib's
// SplitLines would add an empty last line to content that ends in a newline, which gives wrong
// hunk ranges and a fake line at the end of a file. A last line without a newline gets the
// noNewline marker, so it differs from the same line with a newline and the patch stays exact.
func splitLines(content []byte) []string {
	lines := strings.SplitAfter(string(content), "\n")
	last := len(lines) - 1
	if lines[last] == "" {
		return lines[:last]
	}
	lines[last] += noNewline

	return lines
}

// printSummary prints the outcome of a run that wrote its plan. init lists only what it didn't
// create; update lists what it created or overwrote. Both list the files they skipped for lack
// of the marker and the orphans, then a line of counts and a hint for each kind of warning.
func printSummary(w io.Writer, command, workDir string, files []plannedFile) {
	listed := map[Action]bool{ActionSkipUnmarked: true, ActionOrphan: true}
	if command == "init" {
		listed[ActionSkip] = true
	} else {
		listed[ActionCreate] = true
		listed[ActionOverwrite] = true
	}

	counts := make(map[Action]int)
	printed := false
	for _, f := range files {
		counts[f.action]++
		if listed[f.action] {
			printAction(w, f)
			printed = true
		}
	}
	if printed {
		fmt.Fprintln(w)
	}

	parts := make([]string, 0, 6)
	for _, c := range []struct {
		action Action
		label  string
	}{
		{ActionCreate, "created"},
		{ActionOverwrite, "overwritten"},
		{ActionUnchanged, "unchanged"},
		{ActionSkip, "skipped as yours"},
		{ActionSkipUnmarked, "skipped without the marker"},
		{ActionOrphan, "orphaned"},
	} {
		if n := counts[c.action]; n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, c.label))
		}
	}
	if len(parts) == 0 {
		parts = append(parts, "nothing to do")
	}
	fmt.Fprintf(w, "goro %s in %s: %s.\n", command, workDir, strings.Join(parts, ", "))

	if counts[ActionSkipUnmarked] > 0 {
		fmt.Fprintf(w, "skip-unmarked: these files are at paths goro generates but lack the line %q, so they are yours and goro left them alone. "+
			"Put the line back to let goro regenerate a file.\n", marker)
	}
	if counts[ActionOrphan] > 0 {
		fmt.Fprintln(w, "orphan: goro.yaml doesn't describe these files and packages; goro leaves them alone. Delete any that are left over from removed entries.")
	}
}
