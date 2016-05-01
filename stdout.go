package cmd

import (
	"fmt"

	"github.com/desal/richtext"
)

type stdOutput struct {
	verbose  bool
	format   richtext.Format
	warnText string
	errText  string
}

var _ Output = &stdOutput{}

func NewStdOutput(verbose bool, format richtext.Format) *stdOutput {
	warn := format.String(richtext.Orange, richtext.None, richtext.Bold)
	err := format.String(richtext.Red, richtext.None, richtext.Bold)
	return &stdOutput{verbose, format, warn("WARNING: "), err("ERROR: ")}
}

func (o *stdOutput) Verbose(format string, a ...interface{}) {
	if !o.verbose {
		return
	}
	fmt.Printf(format, a...)
	fmt.Print("\n")
}

func (o *stdOutput) Warning(format string, a ...interface{}) {
	fmt.Print(o.warnText)
	fmt.Printf(format, a...)
	fmt.Print("\n")
}
func (o *stdOutput) Error(format string, a ...interface{}) {
	fmt.Print(o.errText)
	fmt.Printf(format, a...)
	fmt.Print("\n")
}

func (o *stdOutput) Format() richtext.Format { return o.format }
