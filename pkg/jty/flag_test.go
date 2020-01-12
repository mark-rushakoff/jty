package jty_test

import (
	"os"
	"reflect"
	"strings"
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
	f.FinishParse("")

	if !f.Zero {
		t.Fatal("expected -z to set Zero, but it didn't")
	}
	if !f.FromStdin {
		t.Fatal("expected -z to set FromStdin, but it didn't")
	}
}

func TestFlags_JPaths(t *testing.T) {
	t.Run("flags only", func(t *testing.T) {
		fs := pflag.NewFlagSet("", pflag.ContinueOnError)
		var f jty.Flags
		f.AddToFlagSet(fs)
		if err := fs.Parse([]string{
			"--jpath=/tmp/1",
			"-J=/tmp/2",
			"--jpath", "/tmp/3",
			"-J", "/tmp/4",
		}); err != nil {
			t.Fatal(err)
		}
		f.FinishParse("")

		expJPaths := []string{"/tmp/1", "/tmp/2", "/tmp/3", "/tmp/4"}
		if !reflect.DeepEqual(f.JPaths, expJPaths) {
			t.Fatalf("expected JPaths %v, got %v", expJPaths, f.JPaths)
		}
	})

	t.Run("JSONNET_PATH only", func(t *testing.T) {
		fs := pflag.NewFlagSet("", pflag.ContinueOnError)
		var f jty.Flags
		f.AddToFlagSet(fs)
		if err := fs.Parse([]string{}); err != nil {
			t.Fatal(err)
		}

		e := strings.Join([]string{"/tmp/1", "/tmp/2"}, string(os.PathListSeparator))
		f.FinishParse(e)

		expJPaths := []string{"/tmp/2", "/tmp/1"}
		if !reflect.DeepEqual(f.JPaths, expJPaths) {
			t.Fatalf("expected JPaths %v, got %v", expJPaths, f.JPaths)
		}
	})

	t.Run("both flags and JSONNET_PATH", func(t *testing.T) {
		fs := pflag.NewFlagSet("", pflag.ContinueOnError)
		var f jty.Flags
		f.AddToFlagSet(fs)
		if err := fs.Parse([]string{"-J", "/tmp/a", "-J", "/tmp/b"}); err != nil {
			t.Fatal(err)
		}

		e := strings.Join([]string{"/tmp/1", "/tmp/2"}, string(os.PathListSeparator))
		f.FinishParse(e)

		expJPaths := []string{"/tmp/2", "/tmp/1", "/tmp/a", "/tmp/b"}
		if !reflect.DeepEqual(f.JPaths, expJPaths) {
			t.Fatalf("expected JPaths %v, got %v", expJPaths, f.JPaths)
		}
	})
}
