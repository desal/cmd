package cmd

import (
	"testing"

	"github.com/desal/richtext"
)

type testOutput struct {
	t      *testing.T
	format richtext.Format
}

var _ Output = &stdOutput{}

func NewTestOutput(t *testing.T) *testOutput {
	return &testOutput{t, richtext.Ascii()}
}

func (o *testOutput) Verbose(format string, a ...interface{}) {
	o.t.Logf(format, a...)
}

func (o *testOutput) Warning(format string, a ...interface{}) {
	o.t.Logf("WARNING "+format, a...)
}

func (o *testOutput) Error(format string, a ...interface{}) {
	o.t.Logf("ERROR "+format, a...)
}

func (o *testOutput) Format() richtext.Format { return o.format }
