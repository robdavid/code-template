package main

import (
	"errors"
	"io"
	"os"

	. "github.com/robdavid/genutil-go/errors/handler"
)

var errEmptyOutputName = errors.New("empty output name")

type outputCache struct {
	outputs map[string]io.Writer
}

func (oc *outputCache) ensure() {
	if oc.outputs == nil {
		oc.outputs = make(map[string]io.Writer)
	}
}

func (oc *outputCache) preOpen(name string, output io.Writer) {
	oc.ensure()
	oc.outputs[name] = output
}

func (oc *outputCache) get(name string) (output io.Writer, err error) {
	var ok bool
	defer Catch(&err)
	if name == "" {
		return nil, errEmptyOutputName
	}
	if output, ok = oc.outputs[name]; !ok {
		oc.ensure()
		if name == "-" {
			output = os.Stdout
		} else {
			output = Try(os.Create(name))
		}
		oc.outputs[name] = output
	}
	return
}

func (oc *outputCache) close() {
	for _, out := range oc.outputs {
		if cw, ok := out.(io.WriteCloser); ok {
			cw.Close()
		}
	}
}
