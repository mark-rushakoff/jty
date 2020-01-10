package jty

import (
	"fmt"
	"io"
	"runtime"

	"github.com/spf13/afero"
)

// Command represents a running CLI environment.
type Command struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer

	FS afero.Fs
}

// Run inputs all the CLI-specified files to a new Processor.
func (c *Command) Run(f *Flags) error {
	p := NewProcessor(runtime.GOMAXPROCS(-1), c.FS, c.Stderr)
	if f.DryRun {
		p.DryRunDest = c.Stdout
	}

	for i := 0; i < len(f.Args); i += 2 {
		p.Process(f.Args[i], f.Args[i+1])
	}

	p.Close()

	// Don't need to take lock, as we have finished all goroutines which may access the field.
	if p.didLogError {
		return fmt.Errorf("encountered errors during processing; failing")
	}

	return nil
}
