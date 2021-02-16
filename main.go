// pwdgo changes Go toolchains as you change directories.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/josharian/cdup"
	"golang.org/x/mod/modfile"
)

var verbose int

func main() {
	log.SetFlags(0)

	forDir := make(map[string]string)
	flag.Func("dir", "override Go version for `dir` and all subdirs, in form dir:go-version-string", func(s string) error {
		parts := strings.Split(s, ":")
		if len(parts) != 2 {
			return fmt.Errorf("-dir arguments must be of form dir:go-version-string, got %q", s)
		}
		forDir[parts[0]] = parts[1]
		return nil
	})

	forModulePath := make(map[string]string)
	flag.Func("path", "override go.mod's Go version for module path `path`, in form module/path:go-version-string", func(s string) error {
		parts := strings.Split(s, ":")
		if len(parts) != 2 {
			return fmt.Errorf("-path arguments must be of form module/path:go-version-string, got %q", s)
		}
		forModulePath[parts[0]] = parts[1]
		return nil
	})

	forGoVersion := make(map[string]string)
	flag.Func("go", "toolchain path for Go version `go`, in form go-version-string:/path/to/toolchain", func(s string) error {
		parts := strings.Split(s, ":")
		if len(parts) != 2 {
			return fmt.Errorf("-go arguments must be of form go-version-string:/path/to/toolchain, got %q", s)
		}
		forGoVersion[parts[0]] = parts[1]
		return nil
	})

	var def string
	flag.StringVar(&def, "default", "", "default Go version to use")
	flag.IntVar(&verbose, "v", 0, "verbosity level")
	flag.Parse()

	if verbose >= 2 {
		log.Printf("Go version by module path: %v", forModulePath)
		log.Printf("toolchain by Go version: %v", forGoVersion)
	}

	pwd, err := os.Getwd()
	check(err)

	var modulePath, goVersion string

	dir, err := cdup.Find(pwd, "go.mod")
	if !errors.Is(err, os.ErrNotExist) {
		check(err)
	}
	if dir != "" {
		file := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(file)
		if !errors.Is(err, os.ErrNotExist) {
			check(err)
		}
		fix := func(path, version string) (string, error) { return version, nil }
		f, err := modfile.ParseLax(file, data, fix)
		check(err)

		modulePath = f.Module.Mod.Path
		goVersion = f.Go.Version
	}

	if verbose >= 2 {
		log.Printf("module path: %q, go version %q", modulePath, goVersion)
	}

	if gv, ok := forModulePath[modulePath]; ok {
		goVersion = gv
		if verbose >= 2 {
			log.Printf("setting Go version to %q based on module path %q", gv, modulePath)
		}
	}
	// Dir overrides override go.mod overrides.
	anc := filepath.Clean(pwd)
	for {
		if gv, ok := forDir[anc]; ok {
			goVersion = gv
			if verbose >= 2 {
				log.Printf("setting Go version to %q based on directory %q", gv, anc)
			}
			break
		}
		parent := filepath.Dir(anc)
		if parent == anc {
			// Hit root.
			break
		}
		// Pop up a directory.
		anc = parent
	}

	if goVersion == "" && def != "" {
		if verbose >= 2 {
			log.Printf("setting Go version to default %q", goVersion)
		}
		goVersion = def
	}

	var toolchain string
	if tc, ok := forGoVersion[goVersion]; ok {
		toolchain = tc
		if verbose >= 2 {
			log.Printf("setting toolchain to %q based on Go version %q", toolchain, goVersion)
		}
	} else {
		if verbose > 0 {
			log.Printf("failed to find toolchain for Go version %s", goVersion)
		}
		// Fall back to default toolchain, if it exists.
		if def != "" {
			goVersion = def
			toolchain = forGoVersion[def]
			if verbose >= 1 && toolchain != "" {
				log.Printf("fell back to default Go version %s", def)
			}
		}
	}

	if verbose >= 1 {
		log.Printf("preferred toolchain: %q", toolchain)
	}

	if verbose >= 1 {
		log.Printf("switching to Go version %s", goVersion)
	}

	toolchains := make(map[string]bool)
	for _, tc := range forGoVersion {
		toolchains[tc] = true
	}
	paths := filepath.SplitList(os.Getenv("PATH"))
	newPaths := make([]string, 0, len(paths))
	if toolchain != "" {
		newPaths = append(newPaths, toolchain)
	}
	for _, path := range paths {
		if toolchains[path] {
			// Drop known toolchain from path.
			continue
		}
		newPaths = append(newPaths, path)
	}
	listSep := string([]rune{filepath.ListSeparator})
	fmt.Print(strings.Join(newPaths, listSep))
}

func check(err error) {
	if err != nil {
		if verbose > 0 {
			log.Fatal(err)
		}
		os.Exit(1)
	}
}
