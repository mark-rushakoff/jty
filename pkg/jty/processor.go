package jty

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	jsonnet "github.com/google/go-jsonnet"
	"github.com/spf13/afero"
	yaml "gopkg.in/yaml.v3"
)

// processRequest is a request to compile the jsonnet at inPath
// to a YAML file saved at outPath.
type processRequest struct {
	InPath, OutPath string
}

// evalRequest is a request to evaluate the jsonnetContent
// and store it as YAML saved at outPath.
// inPath is only used as a string to identify the source file.
type evalRequest struct {
	InPath, OutPath string

	JsonnetContent string
}

// writeRequest is a request to convert the slice of JSON-encoded values
// to YAML, saved as OutPath.
type writeRequest struct {
	OutPath string

	Jsons []string
}

// Processor handles concurrent requests to process input Jsonnet files and save their output as YAML.
type Processor struct {
	// If not nil, Processor will operate in dry run mode and write messages here.
	// Must be set before any calls to Process.
	DryRunDest io.Writer

	vm *jsonnet.VM
	fs afero.Fs

	reqCh   chan processRequest
	evalCh  chan evalRequest
	writeCh chan writeRequest

	reqWG, evalWG, writeWG sync.WaitGroup

	logMu       sync.Mutex
	logDest     io.Writer
	didLogError bool
}

// NewProcessor returns a new Processor that has ioWorkers goroutines to handle reading input files
// and another ioWorkers goroutines to handle writing output files.
func NewProcessor(ioWorkers int, fs afero.Fs, logDest io.Writer) *Processor {
	if ioWorkers < 1 {
		panic(errors.New("NewProcessor: ioWorkers must be positive"))
	}

	p := &Processor{
		vm: jsonnet.MakeVM(),
		fs: fs,

		reqCh:   make(chan processRequest, ioWorkers),
		evalCh:  make(chan evalRequest),
		writeCh: make(chan writeRequest, ioWorkers),

		logDest: logDest,
	}

	p.reqWG.Add(ioWorkers)
	p.writeWG.Add(ioWorkers)
	for i := 0; i < ioWorkers; i++ {
		go p.readFiles()
		go p.writeFiles()
	}

	p.evalWG.Add(1)
	go p.evaluate()
	return p
}

// Close stops processing requests and blocks until outstanding requests have completed.
// After calling Close, calling Process again will panic.
func (p *Processor) Close() {
	close(p.reqCh)
	p.reqWG.Wait()

	close(p.evalCh)
	p.evalWG.Wait()

	close(p.writeCh)
	p.writeWG.Wait()
}

// Process enqueues a request to compile the jsonnet at inPath
// and write the resulting YAML to outPath.
func (p *Processor) Process(inPath, outPath string) {
	p.reqCh <- processRequest{InPath: inPath, OutPath: outPath}
}

func (p *Processor) readFiles() {
	defer p.reqWG.Done()

	for req := range p.reqCh {
		if p.DryRunDest != nil {
			// TODO: does this need a mutex?
			_, _ = fmt.Fprintf(p.DryRunDest, "would process %s and save YAML output to %s\n", req.InPath, req.OutPath)
			continue
		}
		content, err := afero.ReadFile(p.fs, req.InPath)
		if err != nil {
			p.log(fmt.Errorf("failed to read %s: %v", req.InPath, err))
			continue
		}
		p.evalCh <- evalRequest{
			InPath:  req.InPath,
			OutPath: req.OutPath,

			JsonnetContent: string(content),
		}
	}
}

func (p *Processor) evaluate() {
	defer p.evalWG.Done()

	for req := range p.evalCh {
		jsons, err := p.vm.EvaluateSnippetStream(req.InPath, req.JsonnetContent)
		if err != nil {
			p.log(fmt.Errorf("failed to evaluate jsonnet at %s: %v", req.InPath, err))
			continue
		}

		p.writeCh <- writeRequest{
			OutPath: req.OutPath,

			Jsons: jsons,
		}
	}
}

func (p *Processor) writeFiles() {
	defer p.writeWG.Done()

	for req := range p.writeCh {
		if err := p.writeFile(req); err != nil {
			p.log(fmt.Errorf("failed to write output file %s: %v", req.OutPath, err))
			continue
		}
	}
}

func (p *Processor) writeFile(req writeRequest) error {
	outF, err := p.fs.Create(req.OutPath)
	if err != nil {
		return err
	}
	defer outF.Close()

	enc := yaml.NewEncoder(outF)

	for i, j := range req.Jsons {
		var obj interface{}
		if err := json.Unmarshal([]byte(j), &obj); err != nil {
			return fmt.Errorf("error unmarshaling JSON object %d when writing %s: %v", i, req.OutPath, err)
		}

		if i == 0 {
			// Emit a document separator line, because the encoder doesn't do so for the first document.
			if _, err := io.WriteString(outF, "---\n"); err != nil {
				return fmt.Errorf("error writing first document separator when writing %s: %v", req.OutPath, err)
			}
		}
		if err := enc.Encode(obj); err != nil {
			return fmt.Errorf("error encoding YAML document %d when writing %s: %v", i, req.OutPath, err)
		}
	}

	// Must have completely decoded.
	if err := enc.Close(); err != nil {
		return fmt.Errorf("error closing YAML encoder when writing %s: %v", req.OutPath, err)
	}

	// Closing the encoder doesn't emit a stream terminator, so do that ourselves.
	if _, err := io.WriteString(outF, "...\n"); err != nil {
		return fmt.Errorf("error writing YAML stream terminator when writing %s: %v", req.OutPath, err)
	}

	return nil
}

func (p *Processor) log(err error) {
	p.logMu.Lock()
	defer p.logMu.Unlock()

	_, _ = fmt.Fprintln(p.logDest, err.Error())
	p.didLogError = true
}
