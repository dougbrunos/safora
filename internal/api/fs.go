package api

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const maxFsEntries = 2000

type fsEntry struct {
	Name string
	Path string
}

type fsListing struct {
	Path    string // "" means the list of roots
	Parent  string // "" means go to the roots
	Entries []fsEntry
}

// handleFsList lists the sub-directories of ?path= for the dashboard's folder
// picker. Directories only: file names and contents are never exposed.
func (s *Server) handleFsList(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		json.NewEncoder(w).Encode(fsListing{Entries: fsRoots()})
		return
	}
	if !filepath.IsAbs(path) {
		http.Error(w, "path must be absolute", http.StatusBadRequest)
		return
	}
	path = filepath.Clean(path)

	if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
		http.Error(w, "not a directory", http.StatusBadRequest)
		return
	}

	items, err := os.ReadDir(path)
	if err != nil {
		code := http.StatusForbidden
		if os.IsNotExist(err) {
			code = http.StatusNotFound
		}
		http.Error(w, err.Error(), code)
		return
	}

	list := fsListing{Path: path, Entries: []fsEntry{}}
	if parent := filepath.Dir(path); parent != path {
		list.Parent = parent
	}
	for _, it := range items {
		name := it.Name()
		if runtime.GOOS != "windows" && strings.HasPrefix(name, ".") {
			continue
		}
		full := filepath.Join(path, name)
		if !isDir(it, full) {
			continue
		}
		list.Entries = append(list.Entries, fsEntry{Name: name, Path: full})
		if len(list.Entries) >= maxFsEntries {
			break
		}
	}
	sort.Slice(list.Entries, func(i, j int) bool {
		return strings.ToLower(list.Entries[i].Name) < strings.ToLower(list.Entries[j].Name)
	})
	json.NewEncoder(w).Encode(list)
}

// isDir reports whether the entry is a directory, following symlinks.
func isDir(it fs.DirEntry, full string) bool {
	if it.IsDir() {
		return true
	}
	if it.Type()&fs.ModeSymlink != 0 {
		fi, err := os.Stat(full)
		return err == nil && fi.IsDir()
	}
	return false
}

func fsRoots() []fsEntry {
	roots := []fsEntry{}
	if runtime.GOOS == "windows" {
		for c := 'A'; c <= 'Z'; c++ {
			drive := string(c) + `:\`
			if _, err := os.Stat(drive); err == nil {
				roots = append(roots, fsEntry{Name: drive, Path: drive})
			}
		}
		return roots
	}
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots, fsEntry{Name: "~ (" + filepath.Base(home) + ")", Path: home})
	}
	return append(roots, fsEntry{Name: "/", Path: "/"})
}
