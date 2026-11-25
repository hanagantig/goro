package commands

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hanagantig/goro/internal/generator/chains"
	"github.com/hanagantig/goro/pkg/afero"
)

// plannedFile is one entry of a plan: a rendered file and what to do with it, or an orphan.
type plannedFile struct {
	// path is relative to the work dir, with forward slashes.
	path   string
	action Action
	// rendered is what goro generates for path; empty for an orphan.
	rendered []byte
	// current is what the work dir holds at path, read when the file exists.
	current []byte
}

// planFiles decides an action for every file in rendered, the output of the generator, against
// what workDir holds, and lists the orphans in workDir (ADR 0001):
//
//   - a missing file is created;
//   - a rendered file with the marker is goro's: it is overwritten when its content changes,
//     unless the file in workDir lacks the marker, which makes it the user's;
//   - a rendered file without the marker is the user's: once it exists, goro skips it.
//
// The plan is sorted by path.
func planFiles(rendered *afero.Afero, workDir string) ([]plannedFile, error) {
	files := make([]plannedFile, 0)
	renderedDirs := make(map[string]bool)
	renderedFiles := make(map[string]bool)

	err := rendered.Walk("/", func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(filepath.ToSlash(p), "/")
		if rel == "" {
			return nil
		}
		if info.IsDir() {
			renderedDirs[rel] = true
			return nil
		}
		renderedFiles[rel] = true

		content, err := rendered.ReadFile(p)
		if err != nil {
			return err
		}
		f, err := planFile(workDir, rel, content)
		if err != nil {
			return err
		}
		files = append(files, f)

		return nil
	})
	if err != nil {
		return nil, err
	}

	files = dropDepFilesOfUserOwners(files)

	orphans, err := findOrphans(workDir, renderedDirs, renderedFiles)
	if err != nil {
		return nil, err
	}
	files = append(files, orphans...)

	sort.Slice(files, func(i, j int) bool { return files[i].path < files[j].path })

	return files, nil
}

func planFile(workDir, rel string, rendered []byte) (plannedFile, error) {
	f := plannedFile{path: rel, rendered: rendered}

	current, err := os.ReadFile(filepath.Join(workDir, filepath.FromSlash(rel)))
	if errors.Is(err, fs.ErrNotExist) {
		f.action = ActionCreate
		return f, nil
	}
	if err != nil {
		return f, err
	}
	f.current = current

	switch {
	case !hasMarker(rendered):
		f.action = ActionSkip
	case !hasMarker(current):
		f.action = ActionSkipUnmarked
	case bytes.Equal(rendered, current):
		f.action = ActionUnchanged
	default:
		f.action = ActionOverwrite
	}

	return f, nil
}

// dropDepFilesOfUserOwners leaves out the dep files goro would create next to a usecase.go or
// service.go that lost its marker. That file belongs to the user and may still declare the
// dependency interfaces, as files made by older goro versions do, so new dep files would declare
// them twice (ADR 0001).
func dropDepFilesOfUserOwners(files []plannedFile) []plannedFile {
	owned := make(map[string]bool)
	for _, f := range files {
		if f.action == ActionSkipUnmarked && isDepOwner(f.path) {
			owned[path.Dir(f.path)] = true
		}
	}

	kept := make([]plannedFile, 0, len(files))
	for _, f := range files {
		if f.action == ActionCreate && owned[path.Dir(f.path)] && strings.HasPrefix(path.Base(f.path), chains.DepFilePrefix) {
			continue
		}
		kept = append(kept, f)
	}

	return kept
}

// isDepOwner reports whether rel is the generated file whose struct uses the dep files next to
// it: internal/usecase/usecase.go or internal/service/<service>/service.go.
func isDepOwner(rel string) bool {
	if rel == "internal/usecase/usecase.go" {
		return true
	}
	ok, _ := path.Match("internal/service/*/service.go", rel)

	return ok
}

// findOrphans lists the code in workDir that goro.yaml doesn't list, so goro didn't render it:
// use case files, service packages, adapter packages (or whole storage folders), and the method
// and dep files of listed services and adapters. Test files and other non-Go files don't count.
func findOrphans(workDir string, renderedDirs, renderedFiles map[string]bool) ([]plannedFile, error) {
	orphans := make([]plannedFile, 0)
	add := func(rel string) {
		orphans = append(orphans, plannedFile{path: rel, action: ActionOrphan})
	}
	addGoFiles := func(dir string) error {
		names, err := goFiles(workDir, dir)
		if err != nil {
			return err
		}
		for _, name := range names {
			if rel := dir + "/" + name; !renderedFiles[rel] {
				add(rel)
			}
		}
		return nil
	}

	if err := addGoFiles("internal/usecase"); err != nil {
		return nil, err
	}

	services, err := subdirs(workDir, "internal/service")
	if err != nil {
		return nil, err
	}
	for _, svc := range services {
		if !renderedDirs[svc] {
			add(svc)
			continue
		}
		if err := addGoFiles(svc); err != nil {
			return nil, err
		}
	}

	storages, err := subdirs(workDir, "internal/adapter")
	if err != nil {
		return nil, err
	}
	for _, storage := range storages {
		if !renderedDirs[storage] {
			add(storage)
			continue
		}
		adapters, err := subdirs(workDir, storage)
		if err != nil {
			return nil, err
		}
		for _, adapter := range adapters {
			if !renderedDirs[adapter] {
				add(adapter)
				continue
			}
			if err := addGoFiles(adapter); err != nil {
				return nil, err
			}
		}
	}

	return orphans, nil
}

// goFiles returns the names of the non-test Go files directly in dir, a slash-separated path
// relative to workDir. A missing dir has none.
func goFiles(workDir, dir string) ([]string, error) {
	entries, err := readDir(workDir, dir)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		names = append(names, name)
	}

	return names, nil
}

// subdirs returns the directories directly in dir as slash-separated paths relative to workDir,
// leaving out hidden ones. A missing dir has none.
func subdirs(workDir, dir string) ([]string, error) {
	entries, err := readDir(workDir, dir)
	if err != nil {
		return nil, err
	}

	dirs := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			dirs = append(dirs, dir+"/"+e.Name())
		}
	}

	return dirs, nil
}

func readDir(workDir, dir string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(filepath.Join(workDir, filepath.FromSlash(dir)))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	return entries, err
}

// writeFiles carries out the create and overwrite actions of a plan. It never deletes anything.
func writeFiles(workDir string, files []plannedFile) error {
	for _, f := range files {
		if f.action != ActionCreate && f.action != ActionOverwrite {
			continue
		}

		p := filepath.Join(workDir, filepath.FromSlash(f.path))
		if err := os.MkdirAll(filepath.Dir(p), 0o775); err != nil {
			return err
		}
		if err := os.WriteFile(p, f.rendered, 0o644); err != nil {
			return err
		}
	}

	return nil
}

func toReport(files []plannedFile) Report {
	r := Report{Files: make([]FileAction, 0, len(files))}
	for _, f := range files {
		r.Files = append(r.Files, FileAction{Path: f.path, Action: f.action})
	}

	return r
}
