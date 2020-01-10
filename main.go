package main

import (
	"fmt"
	"os"

	"github.com/influxdata/jty/pkg/jty"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
)

func main() {
	fs := pflag.NewFlagSet("jty", pflag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "USAGE: %s [opts] [[INPUT_JSONNET OUTPUT_YAML]...]:\n", os.Args[0])
		fs.PrintDefaults()
	}
	var flags jty.Flags
	flags.AddToFlagSet(fs)
	if err := fs.Parse(os.Args[1:]); err != nil {
		fs.Usage()
		os.Exit(1)
	}
	flags.Args = fs.Args()

	c := &jty.Command{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,

		FS: afero.NewOsFs(),
	}
	if err := c.Run(&flags); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	return
}
