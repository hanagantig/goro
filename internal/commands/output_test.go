package commands_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hanagantig/goro/internal/commands"
)

// actionLine is a dry-run report line: an action, then a path.
var actionLine = regexp.MustCompile(`^(create|overwrite|skip|skip-unmarked|orphan) +\S+$`)

// diffPrefixes start every line of a unified diff.
var diffPrefixes = []string{"--- ", "+++ ", "@@ ", " ", "+", "-", `\ No newline at end of file`}

func TestUpdateDryRunDiffsApply(t *testing.T) {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("needs git to apply the dry-run diffs")
	}

	dir := t.TempDir()
	initProject(t, newConfig(t, dir))

	// Hand edits at the end of marked files, which update throws away: one ending in a newline,
	// one without a final newline.
	appended := map[string]string{
		"internal/app/http.go":   "// HAND\n",
		"internal/app/logger.go": "// HAND",
	}
	for path, tail := range appended {
		writeFile(t, dir, path, readFile(t, dir, path)+tail)
	}

	// goro.yaml: add pgsqlx to storages, which changes app.go, database.go and container.go
	// in the middle and at the end.
	cfg := newConfig(t, dir)
	cfg.Storages = append(cfg.Storages, "pgsqlx")

	var out bytes.Buffer
	dry, err := commands.Update(cfg, commands.Options{SkipGoMod: true, DryRun: true, Out: &out})
	if err != nil {
		t.Fatalf("update --dry-run: %v", err)
	}

	overwritten := make([]string, 0)
	for _, f := range dry.Files {
		if f.Action == commands.ActionOverwrite {
			overwritten = append(overwritten, f.Path)
		}
	}
	for path := range appended {
		if actionOf(dry, path) != commands.ActionOverwrite {
			t.Fatalf("dry run gives %s action %q, want overwrite", path, actionOf(dry, path))
		}
	}

	// The patch is the output without the action lines.
	var patch strings.Builder
	for _, line := range strings.SplitAfter(out.String(), "\n") {
		if !actionLine.MatchString(strings.TrimSuffix(line, "\n")) {
			patch.WriteString(line)
		}
	}
	patchFile := filepath.Join(t.TempDir(), "update.patch")
	if err := os.WriteFile(patchFile, []byte(patch.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command(gitBin, args...)
		cmd.Dir = dir
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s\npatch:\n%s", strings.Join(args, " "), err, b, patch.String())
		}
	}
	// A repository of its own, so git apply doesn't resolve paths against an enclosing one.
	git("init", "-q")
	git("apply", "--check", patchFile)
	git("apply", patchFile)

	// Applied, the patch gives exactly what update writes.
	report := update(t, cfg)
	for _, path := range overwritten {
		assertAction(t, report, path, commands.ActionUnchanged)
	}
}

func TestUpdateDryRunPrintsReportWithDiffs(t *testing.T) {
	dir := t.TempDir()
	initProject(t, newConfig(t, dir))

	const unmarked = "internal/app/container.go"
	writeFile(t, dir, unmarked, removeMarker(readFile(t, dir, unmarked)))

	// goro.yaml: add pgsqlx to storages, which changes app.go, database.go and container.go.
	cfg := newConfig(t, dir)
	cfg.Storages = append(cfg.Storages, "pgsqlx")

	var out bytes.Buffer
	if _, err := commands.Update(cfg, commands.Options{SkipGoMod: true, DryRun: true, Out: &out}); err != nil {
		t.Fatalf("update --dry-run: %v", err)
	}
	printed := out.String()

	t.Run("each overwrite has a diff", func(t *testing.T) {
		for _, path := range []string{"internal/app/app.go", "internal/app/database.go"} {
			if !strings.Contains(printed, "--- a/"+path+"\n+++ b/"+path+"\n") {
				t.Errorf("output has no diff header for %s:\n%s", path, printed)
			}
		}
		if !hasDiffLine(printed, "+", "pgsqlx *sqlx.DB") {
			t.Errorf("app.go diff doesn't add the pgsqlx connection:\n%s", printed)
		}
	})

	t.Run("a skipped file has no diff", func(t *testing.T) {
		if strings.Contains(printed, "+++ b/"+unmarked) {
			t.Errorf("output has a diff for %s, which goro won't write", unmarked)
		}
	})

	t.Run("the output is only the report", func(t *testing.T) {
		for _, line := range strings.Split(strings.TrimSuffix(printed, "\n"), "\n") {
			if actionLine.MatchString(line) {
				continue
			}
			isDiff := false
			for _, p := range diffPrefixes {
				if strings.HasPrefix(line, p) {
					isDiff = true
					break
				}
			}
			if !isDiff {
				t.Errorf("line %q is neither an action nor part of a diff", line)
			}
		}
	})
}

// hasDiffLine reports whether a diff in text has a line that starts with sign and contains part.
func hasDiffLine(text, sign, part string) bool {
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, sign) && !strings.HasPrefix(line, sign+sign+sign+" ") && strings.Contains(line, part) {
			return true
		}
	}

	return false
}

func TestUpdatePrintsSummary(t *testing.T) {
	dir := t.TempDir()
	initProject(t, newConfig(t, dir))

	const unmarked = "internal/app/container.go"
	writeFile(t, dir, unmarked, removeMarker(readFile(t, dir, unmarked)))
	writeFile(t, dir, "internal/service/mailer/mailer.go", "package mailer\n")

	// goro.yaml: add pgsqlx to storages and a SignOut use case method.
	cfg := newConfig(t, dir)
	cfg.Storages = append(cfg.Storages, "pgsqlx")
	cfg.UseCase.Methods = append(cfg.UseCase.Methods, "SignOut")

	var out bytes.Buffer
	if _, err := commands.Update(cfg, commands.Options{SkipGoMod: true, Out: &out}); err != nil {
		t.Fatalf("update: %v", err)
	}
	printed := out.String()

	for _, want := range []struct{ action, path string }{
		{"create", "internal/usecase/sign_out.go"},
		{"overwrite", "internal/app/app.go"},
		{"skip-unmarked", unmarked},
		{"orphan", "internal/service/mailer"},
	} {
		if !hasLineWith(printed, want.action, want.path) {
			t.Errorf("output has no line with %q and %s:\n%s", want.action, want.path, printed)
		}
	}
	if hasLineWith(printed, "main.go") || hasLineWith(printed, "internal/app/logger.go") {
		t.Errorf("output lists skipped user files or unchanged files:\n%s", printed)
	}
	if !strings.Contains(printed, "// Code generated by goro. DO NOT EDIT.") {
		t.Errorf("output doesn't tell how to give an unmarked file back to goro:\n%s", printed)
	}
	if !hasLineWith(printed, "orphan:", "left over") {
		t.Errorf("output doesn't explain orphans:\n%s", printed)
	}
}
