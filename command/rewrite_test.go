package command

import (
	"github.com/DEVELOPEST/gtm-core/util"
	"github.com/mitchellh/cli"
	"log"
	"os"
	"strings"
	"testing"
)

func TestRewriteDefaultOptions(t *testing.T) {
	repo := util.NewTestRepo(t, false)
	defer repo.Remove()
	repo.Seed()
	err := os.Chdir(repo.Workdir())
	if err != nil {
		log.Fatal(err)
	}

	(RewriteCmd{UI: new(cli.MockUi)}).Run([]string{})

	ui := new(cli.MockUi)
	c := RewriteCmd{UI: ui}

	var args []string
	rc := c.Run(args)

	// TODO: Write to stdin

	if rc != 0 {
		t.Errorf("gtm rewrite(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}
}

func TestRewriteInvalidOption(t *testing.T) {
	ui := new(cli.MockUi)
	c := RewriteCmd{UI: ui}

	args := []string{"-invalid"}
	rc := c.Run(args)

	// TODO: Write to stdin

	if rc != 1 {
		t.Errorf("gtm rewrite(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter)
	}
	if !strings.Contains(ui.OutputWriter.String(), "Usage:") {
		t.Errorf("gtm rewrite(%+v), want 'Usage:'  got %d, %s", args, rc, ui.OutputWriter.String())
	}
}
