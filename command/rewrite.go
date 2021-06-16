package command

import (
	"bufio"
	"github.com/DEVELOPEST/gtm-core/project"
	"github.com/DEVELOPEST/gtm-core/scm"
	"io"
	"log"
	"os"
	"strings"

	"github.com/mitchellh/cli"
)

// CommitCmd struct contain methods for commit command
type RewriteCmd struct {
	UI cli.Ui
}

// NewCommit returns new CommitCmd struct
func NewRewrite() (cli.Command, error) {
	return RewriteCmd{}, nil
}

// Help returns help for commit command
func (c RewriteCmd) Help() string {
	helpText := `
Usage: gtm rewrite [options]

  Automatically called on git history rewrite. Do not use manually!.
`
	return strings.TrimSpace(helpText)
}

// Run executes commit commands with args
func (c RewriteCmd) Run(args []string) int {

	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		input = strings.TrimSpace(input)
		hashes := strings.Split(input, " ")

		if len(hashes) < 2 {
			log.Fatal("Unexpected input!")
		}

		err = scm.RewriteNote(hashes[0], hashes[1], project.NoteNameSpace)

		if err != nil {
			log.Println(err)
		}
	}

	return 0
}

// Synopsis return help for commit command
func (c RewriteCmd) Synopsis() string {
	return "Update git notes on history rewrite"
}
