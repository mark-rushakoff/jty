package jty_test

import (
	"testing"

	"github.com/influxdata/jty/pkg/jty"
	"github.com/spf13/pflag"
)

func TestFlags_ZeroSetsStdin(t *testing.T) {
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	var f jty.Flags
	f.AddToFlagSet(fs)
	if err := fs.Parse([]string{"-z"}); err != nil {
		t.Fatal(err)
	}
	f.FinishParse()

	if !f.Zero {
		t.Fatal("expected -z to set Zero, but it didn't")
	}
	if !f.FromStdin {
		t.Fatal("expected -z to set FromStdin, but it didn't")
	}
}
