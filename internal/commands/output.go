package commands

import (
	"fmt"
	"io"
	"strings"
)

// actionWidth fits the longest action name, skip-unmarked, so the paths line up.
const actionWidth = len(ActionSkipUnmarked)

func printAction(w io.Writer, f plannedFile) {
	fmt.Fprintf(w, "%-*s %s\n", actionWidth, f.action, f.path)
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
		fmt.Fprintln(w, "orphan: goro.yaml doesn't list this code. goro keeps it and stops wiring it; delete it yourself when you no longer need it.")
	}
}
