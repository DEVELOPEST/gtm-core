// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package project

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/git-time-metric/gtm/scm"
	isatty "github.com/mattn/go-isatty"
)

var (
	// ErrNotInitialized is raised when a git repo not initialized for time tracking
	ErrNotInitialized = errors.New("Git Time Metric is not initialized")
	// ErrFileNotFound is raised when record an event for a file that does not exist
	ErrFileNotFound = errors.New("File does not exist")
)

var (
	// GitHooks is map of hooks to apply to the git repo
	GitHooks = map[string]string{
		"post-commit": "gtm commit --yes"}
	// GitConfig is map of git configuration settings
	GitConfig = map[string]string{
		"alias.pushgtm":    "push origin refs/notes/gtm-data",
		"alias.fetchgtm":   "fetch origin refs/notes/gtm-data:refs/notes/gtm-data",
		"notes.rewriteref": "refs/notes/gtm-data"}
	// GitIgnore is file ignore to apply to git repo
	GitIgnore = "/.gtm/"
)

const (
	// NoteNameSpace is the gtm git note namespace
	NoteNameSpace = "gtm-data"
	// GTMDir is the subdir for gtm within the git repo root directory
	GTMDir = ".gtm"
)

const initMsgTpl string = `
{{print "Git Time Metric initialized for " (.ProjectPath) | printf (.HeaderFormat) }}

{{ range $hook, $command := .GitHooks -}}
	{{- $hook | printf "%16s" }}: {{ $command }}
{{ end -}}
{{ range $key, $val := .GitConfig -}}
	{{- $key | printf "%16s" }}: {{ $val }}
{{end -}}
{{ print "terminal:" | printf "%17s" }} {{ .Terminal }}
{{ print ".gitignore:" | printf "%17s" }} {{ .GitIgnore }}
{{ print "tags:" | printf "%17s" }} {{.Tags }}
`
const removeMsgTpl string = `
{{print "Git Time Metric uninitialized for " (.ProjectPath) | printf (.HeaderFormat) }}

The following items have been removed.

{{ range $hook, $command := .GitHooks -}}
	{{- $hook | printf "%16s" }}: {{ $command }}
{{ end -}}
{{ range $key, $val := .GitConfig -}}
	{{- $key | printf "%16s" }}: {{ $val }}
{{end -}}
{{ print ".gitignore:" | printf "%17s" }} {{ .GitIgnore }}
`

// Initialize initializes a git repo for time tracking
func Initialize(terminal bool, tags []string, clearTags bool) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	projRoot, err := scm.RootPath(wd)
	if err != nil {
		return "", fmt.Errorf(
			"Unable to intialize Git Time Metric, Git repository not found in %s", projRoot)
	}

	gitPath := filepath.Join(projRoot, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		return "", fmt.Errorf(
			"Unable to intialize Git Time Metric, Git repository not found in %s", gitPath)
	}

	gtmPath := filepath.Join(projRoot, GTMDir)
	if _, err := os.Stat(gtmPath); os.IsNotExist(err) {
		if err := os.MkdirAll(gtmPath, 0700); err != nil {
			return "", err
		}
	}

	if clearTags {
		err = removeTags(gtmPath)
		if err != nil {
			return "", err
		}
	}
	err = saveTags(tags, gtmPath)
	if err != nil {
		return "", err
	}
	tags, err = LoadTags(gtmPath)
	if err != nil {
		return "", err
	}

	if terminal {
		if err := ioutil.WriteFile(filepath.Join(gtmPath, "terminal.app"), []byte(""), 0644); err != nil {
			return "", err
		}
	} else {
		// file may not exist, ignore error
		os.Remove(filepath.Join(gtmPath, "terminal.app"))
	}

	if err := scm.SetHooks(GitHooks, projRoot); err != nil {
		return "", err
	}

	if err := scm.ConfigSet(GitConfig, projRoot); err != nil {
		return "", err
	}

	if err := scm.IgnoreSet(GitIgnore, projRoot); err != nil {
		return "", err
	}

	headerFormat := "%s"
	if isatty.IsTerminal(os.Stdout.Fd()) && runtime.GOOS != "windows" {
		headerFormat = "\x1b[1m%s\x1b[0m"
	}

	b := new(bytes.Buffer)
	t := template.Must(template.New("msg").Parse(initMsgTpl))
	err = t.Execute(b,
		struct {
			Tags         string
			HeaderFormat string
			ProjectPath  string
			GitHooks     map[string]string
			GitConfig    map[string]string
			GitIgnore    string
			Terminal     bool
		}{
			strings.Join(tags, " "),
			headerFormat,
			projRoot,
			GitHooks,
			GitConfig,
			GitIgnore,
			terminal,
		})

	if err != nil {
		return "", err
	}

	index, err := NewIndex()
	if err != nil {
		return "", err
	}

	index.add(projRoot)
	err = index.save()
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

//Uninitialize remove GTM tracking from the project in the current working directory
func Uninitialize() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	projRoot, err := scm.RootPath(wd)
	if err != nil {
		return "", fmt.Errorf(
			"Unable to unintialize Git Time Metric, Git repository not found in %s", projRoot)
	}

	gtmPath := filepath.Join(projRoot, GTMDir)
	if _, err := os.Stat(gtmPath); os.IsNotExist(err) {
		return "", fmt.Errorf(
			"Unable to uninitialize Git Time Metric, %s directory not found", gtmPath)
	}
	if err := scm.RemoveHooks(GitHooks, projRoot); err != nil {
		return "", err
	}
	if err := scm.ConfigRemove(GitConfig, projRoot); err != nil {
		return "", err
	}
	if err := scm.IgnoreRemove(GitIgnore, projRoot); err != nil {
		return "", err
	}
	if err := os.RemoveAll(gtmPath); err != nil {
		return "", err
	}

	headerFormat := "%s"
	if isatty.IsTerminal(os.Stdout.Fd()) && runtime.GOOS != "windows" {
		headerFormat = "\x1b[1m%s\x1b[0m"
	}
	b := new(bytes.Buffer)
	t := template.Must(template.New("msg").Parse(removeMsgTpl))
	err = t.Execute(b,
		struct {
			HeaderFormat string
			ProjectPath  string
			GitHooks     map[string]string
			GitConfig    map[string]string
			GitIgnore    string
		}{
			headerFormat,
			projRoot,
			GitHooks,
			GitConfig,
			GitIgnore})

	if err != nil {
		return "", err
	}

	index, err := NewIndex()
	if err != nil {
		return "", err
	}

	index.remove(projRoot)
	err = index.save()
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

//Clean removes any event or metrics files from project in the current working directory
func Clean() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	projRoot, err := scm.RootPath(wd)
	if err != nil {
		return "", fmt.Errorf(
			"Unable to clean, Git repository not found in %s", projRoot)
	}

	gtmPath := filepath.Join(projRoot, GTMDir)
	if _, err := os.Stat(gtmPath); os.IsNotExist(err) {
		return "", fmt.Errorf(
			"Unable to clean GTM data, %s directory not found", gtmPath)
	}

	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".event") &&
			!strings.HasSuffix(f.Name(), ".metric") {
			continue
		}
		if err := os.Remove(filepath.Join(gtmPath, f.Name())); err != nil {
			return "", err
		}
	}

	return "", nil
}

// Paths returns the root git repo and gtm paths
func Paths(wd ...string) (string, string, error) {
	var (
		repoPath string
		err      error
	)
	if len(wd) > 0 {
		repoPath, err = scm.RootPath(wd[0])
	} else {
		repoPath, err = scm.RootPath()
	}
	if err != nil {
		return "", "", ErrNotInitialized
	}

	gtmPath := filepath.Join(repoPath, GTMDir)
	if _, err := os.Stat(gtmPath); os.IsNotExist(err) {
		return "", "", ErrNotInitialized
	}
	return repoPath, gtmPath, nil
}

func removeTags(gtmPath string) error {
	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		return err
	}
	for i := range files {
		if strings.HasSuffix(files[i].Name(), ".tag") {
			tagFile := filepath.Join(gtmPath, files[i].Name())
			if err := os.Remove(tagFile); err != nil {
				return err
			}
		}
	}
	return nil
}

// LoadTags returns the tags for the project in the gtmPath directory
func LoadTags(gtmPath string) ([]string, error) {
	tags := []string{}
	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		return []string{}, err
	}
	for i := range files {
		if strings.HasSuffix(files[i].Name(), ".tag") {
			tags = append(tags, strings.TrimSuffix(files[i].Name(), filepath.Ext(files[i].Name())))
		}
	}
	return tags, nil
}

func saveTags(tags []string, gtmPath string) error {
	if len(tags) > 0 {
		for _, t := range tags {
			if strings.TrimSpace(t) == "" {
				continue
			}
			if err := ioutil.WriteFile(filepath.Join(gtmPath, fmt.Sprintf("%s.tag", t)), []byte(""), 0644); err != nil {
				return err
			}
		}
	}
	return nil
}
