// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package afero

import (
	"fmt"
	"github.com/hanagantig/goro/pkg/afero/mem"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const chmodBits = os.ModePerm | os.ModeSetuid | os.ModeSetgid | os.ModeSticky // Only a subset of bits are allowed to be changed. Documented under os.Chmod()

type MemMapFs struct {
	mu   sync.RWMutex
	data map[string]*mem.FileData
	init sync.Once
}

func NewMemMapFs() Fs {
	return &MemMapFs{}
}

func (m *MemMapFs) getData() map[string]*mem.FileData {
	m.init.Do(func() {
		m.data = make(map[string]*mem.FileData)
		// Root should always exist, right?
		// TODO: what about windows?
		root := mem.CreateDir(FilePathSeparator)
		mem.SetMode(root, os.ModeDir|0755)
		m.data[FilePathSeparator] = root
	})
	return m.data
}

func (*MemMapFs) Name() string { return "MemMapFS" }

func (m *MemMapFs) Create(name string) (File, error) {
	name = normalizePath(name)
	m.mu.Lock()
	file := mem.CreateFile(name)
	m.getData()[name] = file
	m.registerWithParent(file, 0)
	m.mu.Unlock()
	return mem.NewFileHandle(file), nil
}

func (m *MemMapFs) unRegisterWithParent(fileName string) error {
	f, err := m.lockfreeOpen(fileName)
	if err != nil {
		return err
	}
	parent := m.findParent(f)
	if parent == nil {
		log.Panic("parent of ", f.Name(), " is nil")
	}

	parent.Lock()
	mem.RemoveFromMemDir(parent, f)
	parent.Unlock()
	return nil
}

func (m *MemMapFs) findParent(f *mem.FileData) *mem.FileData {
	pdir, _ := filepath.Split(f.Name())
	pdir = filepath.Clean(pdir)
	pfile, err := m.lockfreeOpen(pdir)
	if err != nil {
		return nil
	}
	return pfile
}

func (m *MemMapFs) findDescendants(name string) []*mem.FileData {
	fData := m.getData()
	descendants := make([]*mem.FileData, 0, len(fData))
	for p, dFile := range fData {
		if strings.HasPrefix(p, name+FilePathSeparator) {
			descendants = append(descendants, dFile)
		}
	}

	sort.Slice(descendants, func(i, j int) bool {
		cur := len(strings.Split(descendants[i].Name(), FilePathSeparator))
		next := len(strings.Split(descendants[j].Name(), FilePathSeparator))
		return cur < next
	})

	return descendants
}

func (m *MemMapFs) registerWithParent(f *mem.FileData, perm os.FileMode) {
	if f == nil {
		return
	}
	parent := m.findParent(f)
	if parent == nil {
		pdir := filepath.Dir(filepath.Clean(f.Name()))
		err := m.lockfreeMkdir(pdir, perm)
		if err != nil {
			//log.Println("Mkdir error:", err)
			return
		}
		parent, err = m.lockfreeOpen(pdir)
		if err != nil {
			//log.Println("Open after Mkdir error:", err)
			return
		}
	}

	parent.Lock()
	mem.InitializeDir(parent)
	mem.AddToMemDir(parent, f)
	parent.Unlock()
}

func (m *MemMapFs) lockfreeMkdir(name string, perm os.FileMode) error {
	name = normalizePath(name)
	x, ok := m.getData()[name]
	if ok {
		// Only return ErrFileExists if it's a file, not a directory.
		i := mem.FileInfo{FileData: x}
		if !i.IsDir() {
			return ErrFileExists
		}
	} else {
		item := mem.CreateDir(name)
		mem.SetMode(item, os.ModeDir|perm)
		m.getData()[name] = item
		m.registerWithParent(item, perm)
	}
	return nil
}

func (m *MemMapFs) Mkdir(name string, perm os.FileMode) error {
	perm &= chmodBits
	name = normalizePath(name)

	m.mu.RLock()
	_, ok := m.getData()[name]
	m.mu.RUnlock()
	if ok {
		return &os.PathError{Op: "mkdir", Path: name, Err: ErrFileExists}
	}

	m.mu.Lock()
	item := mem.CreateDir(name)
	mem.SetMode(item, os.ModeDir|perm)
	m.getData()[name] = item
	m.registerWithParent(item, perm)
	m.mu.Unlock()

	return m.setFileMode(name, perm|os.ModeDir)
}

func (m *MemMapFs) MkdirAll(path string, perm os.FileMode) error {
	err := m.Mkdir(path, perm)
	if err != nil {
		if err.(*os.PathError).Err == ErrFileExists {
			return nil
		}
		return err
	}
	return nil
}

// Handle some relative paths
func normalizePath(path string) string {
	path = filepath.Clean(path)

	switch path {
	case ".":
		return FilePathSeparator
	case "..":
		return FilePathSeparator
	default:
		return path
	}
}

func (m *MemMapFs) Open(name string) (File, error) {
	f, err := m.open(name)
	if f != nil {
		return mem.NewReadOnlyFileHandle(f), err
	}
	return nil, err
}

func (m *MemMapFs) openWrite(name string) (File, error) {
	f, err := m.open(name)
	if f != nil {
		return mem.NewFileHandle(f), err
	}
	return nil, err
}

func (m *MemMapFs) open(name string) (*mem.FileData, error) {
	name = normalizePath(name)

	m.mu.RLock()
	f, ok := m.getData()[name]
	m.mu.RUnlock()
	if !ok {
		return nil, &os.PathError{Op: "open", Path: name, Err: ErrFileNotFound}
	}
	return f, nil
}

func (m *MemMapFs) lockfreeOpen(name string) (*mem.FileData, error) {
	name = normalizePath(name)
	f, ok := m.getData()[name]
	if ok {
		return f, nil
	} else {
		return nil, ErrFileNotFound
	}
}

func (m *MemMapFs) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	perm &= chmodBits
	chmod := false
	file, err := m.openWrite(name)
	if err == nil && (flag&os.O_EXCL > 0) {
		return nil, &os.PathError{Op: "open", Path: name, Err: ErrFileExists}
	}
	if os.IsNotExist(err) && (flag&os.O_CREATE > 0) {
		file, err = m.Create(name)
		chmod = true
	}
	if err != nil {
		return nil, err
	}
	if flag == os.O_RDONLY {
		file = mem.NewReadOnlyFileHandle(file.(*mem.File).Data())
	}
	if flag&os.O_APPEND > 0 {
		_, err = file.Seek(0, os.SEEK_END)
		if err != nil {
			file.Close()
			return nil, err
		}
	}
	if flag&os.O_TRUNC > 0 && flag&(os.O_RDWR|os.O_WRONLY) > 0 {
		err = file.Truncate(0)
		if err != nil {
			file.Close()
			return nil, err
		}
	}
	if chmod {
		return file, m.setFileMode(name, perm)
	}
	return file, nil
}

func (m *MemMapFs) Remove(name string) error {
	name = normalizePath(name)

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.getData()[name]; ok {
		err := m.unRegisterWithParent(name)
		if err != nil {
			return &os.PathError{Op: "remove", Path: name, Err: err}
		}
		delete(m.getData(), name)
	} else {
		return &os.PathError{Op: "remove", Path: name, Err: os.ErrNotExist}
	}
	return nil
}

func (m *MemMapFs) RemoveAll(path string) error {
	path = normalizePath(path)
	m.mu.Lock()
	m.unRegisterWithParent(path)
	m.mu.Unlock()

	m.mu.RLock()
	defer m.mu.RUnlock()

	for p := range m.getData() {
		if p == path || strings.HasPrefix(p, path+FilePathSeparator) {
			m.mu.RUnlock()
			m.mu.Lock()
			delete(m.getData(), p)
			m.mu.Unlock()
			m.mu.RLock()
		}
	}
	return nil
}

func (m *MemMapFs) Rename(oldName, newName string) error {
	oldName = normalizePath(oldName)
	newName = normalizePath(newName)

	if oldName == newName {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.getData()[oldName]; ok {
		m.mu.RUnlock()
		m.mu.Lock()
		err := m.unRegisterWithParent(oldName)
		if err != nil {
			return err
		}

		fileData := m.getData()[oldName]
		mem.ChangeFileName(fileData, newName)
		m.getData()[newName] = fileData

		err = m.renameDescendants(oldName, newName)
		if err != nil {
			return err
		}

		delete(m.getData(), oldName)

		m.registerWithParent(fileData, 0)
		m.mu.Unlock()
		m.mu.RLock()
	} else {
		return &os.PathError{Op: "rename", Path: oldName, Err: ErrFileNotFound}
	}
	return nil
}

func (m *MemMapFs) renameDescendants(oldName, newName string) error {
	descendants := m.findDescendants(oldName)
	removes := make([]string, 0, len(descendants))
	for _, desc := range descendants {
		descNewName := strings.Replace(desc.Name(), oldName, newName, 1)
		err := m.unRegisterWithParent(desc.Name())
		if err != nil {
			return err
		}

		removes = append(removes, desc.Name())
		mem.ChangeFileName(desc, descNewName)
		m.getData()[descNewName] = desc

		m.registerWithParent(desc, 0)
	}
	for _, r := range removes {
		delete(m.getData(), r)
	}

	return nil
}

func (m *MemMapFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	fileInfo, err := m.Stat(name)
	return fileInfo, false, err
}

func (m *MemMapFs) Stat(name string) (os.FileInfo, error) {
	f, err := m.Open(name)
	if err != nil {
		return nil, err
	}
	fi := mem.GetFileInfo(f.(*mem.File).Data())
	return fi, nil
}

func (m *MemMapFs) Chmod(name string, mode os.FileMode) error {
	mode &= chmodBits

	m.mu.RLock()
	f, ok := m.getData()[name]
	m.mu.RUnlock()
	if !ok {
		return &os.PathError{Op: "chmod", Path: name, Err: ErrFileNotFound}
	}
	prevOtherBits := mem.GetFileInfo(f).Mode() & ^chmodBits

	mode = prevOtherBits | mode
	return m.setFileMode(name, mode)
}

func (m *MemMapFs) setFileMode(name string, mode os.FileMode) error {
	name = normalizePath(name)

	m.mu.RLock()
	f, ok := m.getData()[name]
	m.mu.RUnlock()
	if !ok {
		return &os.PathError{Op: "chmod", Path: name, Err: ErrFileNotFound}
	}

	m.mu.Lock()
	mem.SetMode(f, mode)
	m.mu.Unlock()

	return nil
}

func (m *MemMapFs) Chown(name string, uid, gid int) error {
	name = normalizePath(name)

	m.mu.RLock()
	f, ok := m.getData()[name]
	m.mu.RUnlock()
	if !ok {
		return &os.PathError{Op: "chown", Path: name, Err: ErrFileNotFound}
	}

	mem.SetUID(f, uid)
	mem.SetGID(f, gid)

	return nil
}

func (m *MemMapFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	name = normalizePath(name)

	m.mu.RLock()
	f, ok := m.getData()[name]
	m.mu.RUnlock()
	if !ok {
		return &os.PathError{Op: "chtimes", Path: name, Err: ErrFileNotFound}
	}

	m.mu.Lock()
	mem.SetModTime(f, mtime)
	m.mu.Unlock()

	return nil
}

func (m *MemMapFs) List() {
	for _, x := range m.data {
		y := mem.FileInfo{FileData: x}
		fmt.Println(x.Name(), y.Size())
	}
}
