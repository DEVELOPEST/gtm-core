// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package command

import (
	"bytes"
	"testing"

	"github.com/mitchellh/cli"
)

func TestCheck(t *testing.T) {
	cases := []struct {
		input string
		cmd   VerifyCmd
		valid bool
		err   bool
	}{
		{">= 1.2", VerifyCmd{Version: "1.0.0"}, false, false},
		{">= 1.0.0", VerifyCmd{Version: "v1.0.0"}, true, false},
		{">= 1.0.0", VerifyCmd{Version: "V1.0.0"}, true, false},
		{">= 1.0.0", VerifyCmd{Version: "1.0.0"}, true, false},
		{">= 1.0-beta.5", VerifyCmd{Version: "v1.0-beta.5"}, true, false},
		{">= 1.0.0", VerifyCmd{Version: "1.0.xxx"}, false, true},
	}

	for _, tc := range cases {
		valid, err := tc.cmd.check(tc.input)
		if tc.err && err == nil {
			t.Fatalf("expected error for input: '%s' Version: %s", tc.input, tc.cmd.Version)
		} else if !tc.err && err != nil {
			t.Fatalf("error for for input: '%s' Version: %s: %s", tc.input, tc.cmd.Version, err)
		}
		if valid != tc.valid {
			t.Fatalf("input: '%s' Version: %s\nexpected  %t\nactual: %t",
				tc.input, tc.cmd.Version, tc.valid, valid)
		}
	}

}

func TestVerify(t *testing.T) {
	ui := new(cli.MockUi)
	c := VerifyCmd{UI: ui, Version: "1.0.0", Out: new(bytes.Buffer)}

	args := []string{">= 1.0.0"}
	rc := c.Run(args)

	if rc != 0 {
		t.Errorf("gtm verify(%+v), want 0 got %d, %s", args, rc, ui.ErrorWriter.String())
	}

	want := "true"
	if want != c.Out.String() {
		t.Errorf("gtm verify(%+v), want '%s' got '%s', %s", args, want, ui.OutputWriter.String(), ui.ErrorWriter.String())
	}
}
