package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/influxdata/jty/pkg/jty"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
)

func main() {
	fs := pflag.NewFlagSet("jty", pflag.ExitOnError)
	fs.Usage = func() {
		exe := filepath.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "USAGE: %s [opts] [[INPUT_JSONNET OUTPUT_YAML]...]:\n", exe)
		fmt.Fprintln(os.Stderr, fs.FlagUsages())
		fmt.Fprintf(os.Stderr, `ENVIRONMENT VARIABLES

JSONNET_PATH is a colon-(semicolon on Windows) separated list of directories added
in reverse order before the paths specified by --jpath (i.e. left-most wins).
The follow three invocations are equivalent:
    JSONNET_PATH=a:b jty -J c -J d
    JSONNET_PATH=d:c:a:b jty
    jty -J b -J a -J c -J d

`)

		fmt.Fprintf(os.Stderr, `EXAMPLE USES

Evaluate in.jsonnet and save the resulting YAML as out.yaml:
    %[1]s in.jsonnet out.yaml

Evaluate multiple .jsonnet files and save the resulting YAML in specific locations:
    %[1]s in1.jsonnet out/1.yaml conf.jsonnet conf.yaml

Evaluate each .jsonnet file under the current directory,
and save the .yml file adjacent to the .jsonnet file:
    find . -name '*.jsonnet' \
      -exec bash -c 'for p in "$@"; do
        printf "%%s\n%%s.yml\n" "$p" "${p%%.jsonnet}"
        done' _ {} + |
      %[1]s -i

Evaluate each .jsonnet file under the current directory,
and for each file foo.jsonnet save a relative yml/foo.yml file
(useful for tools that expect only .yml files in a directory):
    find . -name '*.jsonnet' \
      -exec bash -c 'for p in "$@"; do
        printf "%%s\n%%s/yml/%%s.yml\n" "$p" "$(dirname "$p")" "$(basename "$p" .jsonnet)"
        done' _ {} + |
      %[1]s -i
`, exe)
	}
	var flags jty.Flags
	flags.AddToFlagSet(fs)
	if err := fs.Parse(os.Args[1:]); err != nil {
		fs.Usage()
		os.Exit(1)
	}
	flags.FinishParse(os.Getenv("JSONNET_PATH"))

	if flags.HelpRequested {
		fs.Usage()
		os.Exit(0)
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
}
