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
	"regexp"
	"runtime"
	"strings"
	"text/template"

	"github.com/DEVELOPEST/gtm-core/scm"
	"github.com/DEVELOPEST/gtm-core/util"
	"github.com/mattn/go-isatty"
)

var (
	// ErrNotInitialized is raised when a git repo not initialized for time tracking
	ErrNotInitialized = errors.New("Git Time Metric is not initialized")
	// ErrFileNotFound is raised when record an event for a file that does not exist
	ErrFileNotFound = errors.New("File does not exist")
	// AppEventFileContentRegex regex for app event files
	AppEventFileContentRegex = regexp.MustCompile(`\.gtm[\\/](?P<appName>.*)\.(?P<eventType>app|run|build)`)
)

var (
	// GitHooks is map of hooks to apply to the git repo
	GitHooks = map[string]scm.GitHook{
		"post-commit": {
			Exe:     "gtm",
			Command: "gtm commit --yes",
			RE:      regexp.MustCompile(`(?s)[/:a-zA-Z0-9$_=()"\.\|\-\\ ]*gtm(.exe"|)\s+commit\s+--yes\.*`),
		},
		// "post-rewrite": {
		// 	Exe:     "gtm",
		// 	Command: "gtm rewrite",
		// 	RE: regexp.MustCompile(
		// 		`(?s)[/:a-zA-Z0-9$_=()"\.\|\-\\ ]*gtm\s+rewrite\.*`),
		// },
	}
	// GitConfig is map of git configuration settings
	GitConfig = map[string]string{
		"alias.pushgtm":        "push origin refs/notes/gtm-data",
		"alias.fetchgtm":       "fetch origin refs/notes/gtm-data:refs/notes/gtm-data",
		"notes.rewriteRef":     "refs/notes/gtm-data",
		"notes.rewriteMode":    "concatenate",
		"notes.rewrite.rebase": "true",
		"notes.rewrite.amend":  "true"}
	// GitIgnore is file ignore to apply to git repo
	GitIgnore = "/.gtm/"

	GitFetchRefs = []string{
		"+refs/notes/gtm-data:refs/notes/gtm-data",
	}

	GitPushRefsHooks = map[string]scm.GitHook{
		"pre-push": {
			Exe:     "git",
			Command: "git push origin refs/notes/gtm-data --no-verify",
			RE: regexp.MustCompile(
				`(?s)[/:a-zA-Z0-9$_=()"\.\|\-\\ ]*git\s+push\s+origin\s+refs/notes/gtm-data\s+--no-verify\.*`),
		},
	}

	GitLabHooks = map[string]scm.GitHook{
		"prepare-commit-msg": {
			Exe:     "gtm",
			Command: "gtm status --auto-log=gitlab >> $1",
			RE: regexp.MustCompile(
				`(?s)[/:a-zA-Z0-9$_=()"\.\|\-\\ ]*gtm(.exe"|)\s+status\s+--auto-log=gitlab\s+>>\s+\$1\.*`),
		},
	}

	JiraHooks = map[string]scm.GitHook{
		"prepare-commit-msg": {
			Exe:     "gtm",
			Command: "gtm status --auto-log=jira >> $1",
			RE: regexp.MustCompile(
				`(?s)[/:a-zA-Z0-9$_=()"\.\|\-\\ ]*gtm(.exe"|)\s+status\s+--auto-log=jira\s+>>\s+\$1\.*`),
		},
	}
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
	{{- $hook | printf "%20s" }}: {{ $command.Command }}
{{ end -}}
{{ range $key, $val := .GitConfig -}}
	{{- $key | printf "%20s" }}: {{ $val }}
{{end -}}
{{ range $ref := .GitFetchRefs -}}
    {{ print "add fetch ref:" | printf "%21s" }} {{ $ref}}
{{end -}}
{{ print "terminal:" | printf "%21s" }} {{ .Terminal }}
{{ print ".gitignore:" | printf "%21s" }} {{ .GitIgnore }}
{{ print "tags:" | printf "%21s" }} {{.Tags }}
`
const removeMsgTpl string = `
{{print "Git Time Metric uninitialized for " (.ProjectPath) | printf (.HeaderFormat) }}

The following items have been removed.

{{ range $hook, $command := .GitHooks -}}
	{{- $hook | printf "%20s" }}: {{ $command.Command }}
{{ end -}}
{{ range $key, $val := .GitConfig -}}
	{{- $key | printf "%20s" }}: {{ $val }}
{{end -}}
{{ print ".gitignore:" | printf "%21s" }} {{ .GitIgnore }}
`

// Initialize initializes a git repo for time tracking
func Initialize(
	terminal bool,
	tags []string,
	clearTags bool,
	autoLog string,
	local bool,
	cwd string,
) (string, error) {
	var (
		wd  string
		err error
	)
	gitRepoPath, workDirRoot, gtmPath, err := SetUpPaths(cwd, wd, err)
	if err != nil {
		return "", err
	}

	if clearTags {
		err = removeTags(gtmPath)
		if err != nil {
			return "", err
		}
	}
	tags, err = SetupTags(err, tags, gtmPath)
	if err != nil {
		return "", err
	}

	if terminal {
		if err := ioutil.WriteFile(filepath.Join(gtmPath, "terminal.app"), []byte(""), 0644); err != nil {
			return "", err
		}
	} else {
		// try to remove terminal.app, it may not exist
		_ = os.Remove(filepath.Join(gtmPath, "terminal.app"))
	}

	err = SetupHooks(local, gitRepoPath, autoLog)
	if err != nil {
		return "", err
	}

	if err := scm.ConfigSet(GitConfig, gitRepoPath); err != nil {
		return "", err
	}

	if err := scm.IgnoreSet(GitIgnore, workDirRoot); err != nil {
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
			GitHooks     map[string]scm.GitHook
			GitConfig    map[string]string
			GitFetchRefs []string
			GitIgnore    string
			Terminal     bool
		}{
			strings.Join(tags, " "),
			headerFormat,
			workDirRoot,
			GitHooks,
			GitConfig,
			GitFetchRefs,
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

	index.add(workDirRoot)
	err = index.save()
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func SetupHooks(local bool, gitRepoPath, autoLog string) error {
	if !local {
		if err := scm.FetchRemotesAddRefSpecs(GitFetchRefs, gitRepoPath); err != nil {
			return err
		}
		for k, v := range GitPushRefsHooks {
			GitHooks[k] = v
		}
	}

	switch autoLog {
	case "gitlab":
		for k, v := range GitLabHooks {
			GitHooks[k] = v
		}
	case "jira":
		for k, v := range JiraHooks {
			GitHooks[k] = v
		}
	}

	if err := scm.SetHooks(GitHooks, gitRepoPath); err != nil {
		return err
	}
	return nil
}

func SetupTags(err error, tags []string, gtmPath string) ([]string, error) {
	err = saveTags(tags, gtmPath)
	if err != nil {
		return nil, err
	}
	tags, err = LoadTags(gtmPath)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func SetUpPaths(cwd, wd string, err error) (
	gitRepoPath, workDirRoot, gtmPath string, setupError error) {
	if cwd == "" {
		wd, err = os.Getwd()
	} else {
		wd = cwd
	}

	if err != nil {
		return "", "", "", err
	}

	gitRepoPath, err = scm.GitRepoPath(wd)
	if err != nil {
		return "", "", "", fmt.Errorf(
			"Unable to initialize Git Time Metric, Git repository not found in '%s'", wd)
	}
	if _, err := os.Stat(gitRepoPath); os.IsNotExist(err) {
		return "", "", "", fmt.Errorf(
			"Unable to initialize Git Time Metric, Git repository not found in %s", gitRepoPath)
	}

	workDirRoot, err = scm.Workdir(gitRepoPath)
	if err != nil {
		return "", "", "", fmt.Errorf(
			"Unable to initialize Git Time Metric, Git working tree root not found in %s", workDirRoot)

	}

	if _, err := os.Stat(workDirRoot); os.IsNotExist(err) {
		return "", "", "", fmt.Errorf(
			"Unable to initialize Git Time Metric, Git working tree root not found in %s", workDirRoot)
	}

	gtmPath = filepath.Join(workDirRoot, GTMDir)
	if _, err := os.Stat(gtmPath); os.IsNotExist(err) {
		if err := os.MkdirAll(gtmPath, 0700); err != nil {
			return "", "", "", err
		}
	}
	return gitRepoPath, workDirRoot, gtmPath, nil
}

//Uninitialize remove GTM tracking from the project in the current working directory
func Uninitialize() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	gitRepoPath, err := scm.GitRepoPath(wd)
	if err != nil {
		return "", fmt.Errorf(
			"Unable to uninitialize Git Time Metric, Git repository not found in %s", gitRepoPath)
	}

	workDir, _ := scm.Workdir(gitRepoPath)
	gtmPath := filepath.Join(workDir, GTMDir)
	if _, err := os.Stat(gtmPath); os.IsNotExist(err) {
		return "", fmt.Errorf(
			"Unable to uninitialize Git Time Metric, %s directory not found", gtmPath)
	}
	if err := scm.RemoveHooks(GitLabHooks, gitRepoPath); err != nil {
		return "", err
	}
	if err := scm.RemoveHooks(GitHooks, gitRepoPath); err != nil {
		return "", err
	}
	if err := scm.RemoveHooks(GitPushRefsHooks, gitRepoPath); err != nil {
		return "", err
	}
	if err := scm.ConfigRemove(GitConfig, gitRepoPath); err != nil {
		return "", err
	}
	if err := scm.FetchRemotesRemoveRefSpecs(GitFetchRefs, gitRepoPath); err != nil {
		return "", err
	}
	if err := scm.IgnoreRemove(GitIgnore, workDir); err != nil {
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
			GitHooks     map[string]scm.GitHook
			GitConfig    map[string]string
			GitIgnore    string
		}{
			headerFormat,
			workDir,
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

	index.remove(workDir)
	err = index.save()
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

//Clean removes any event or metrics files from project in the current working directory
func Clean(dr util.DateRange, terminalOnly bool, appOnly bool) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	gitRepoPath, err := scm.GitRepoPath(wd)
	if err != nil {
		return fmt.Errorf("Unable to clean, Git repository not found in %s", gitRepoPath)
	}

	workDir, err := scm.Workdir(gitRepoPath)
	if err != nil {
		return err
	}

	gtmPath := filepath.Join(workDir, GTMDir)
	if _, err := os.Stat(gtmPath); os.IsNotExist(err) {
		return fmt.Errorf("Unable to clean GTM data, %s directory not found", gtmPath)
	}

	files, err := ioutil.ReadDir(gtmPath)
	if err != nil {
		return err
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".event") &&
			!strings.HasSuffix(f.Name(), ".metric") {
			continue
		}
		if !dr.Within(f.ModTime()) {
			continue
		}

		fp := filepath.Join(gtmPath, f.Name())
		if (terminalOnly || appOnly) && strings.HasSuffix(f.Name(), ".event") {
			b, err := ioutil.ReadFile(fp)
			if err != nil {
				return err
			}

			if terminalOnly {
				if !strings.Contains(string(b), "terminal.app") {
					continue
				}
			} else if appOnly {
				if !AppEventFileContentRegex.MatchString(string(b)) {
					continue
				}
			}
		}

		if err := os.Remove(fp); err != nil {
			return err
		}
	}
	return nil
}

// Paths returns the root git repo and gtm paths
func Paths(wd ...string) (string, string, error) {
	defer util.Profile()()

	var (
		gitRepoPath string
		err         error
	)
	if len(wd) > 0 {
		gitRepoPath, err = scm.GitRepoPath(wd[0])
	} else {
		gitRepoPath, err = scm.GitRepoPath()
	}
	if err != nil {
		return "", "", ErrNotInitialized
	}

	workDir, err := scm.Workdir(gitRepoPath)
	if err != nil {
		return "", "", ErrNotInitialized
	}

	gtmPath := filepath.Join(workDir, GTMDir)
	if _, err := os.Stat(gtmPath); os.IsNotExist(err) {
		return "", "", ErrNotInitialized
	}
	return workDir, gtmPath, nil
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
	var tags []string
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
			if err := ioutil.WriteFile(
				filepath.Join(gtmPath, fmt.Sprintf("%s.tag", t)), []byte(""), 0644); err != nil {
				return err
			}
		}
	}
	return nil
}
