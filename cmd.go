package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/desal/richtext"
)

//go:generate stringer -type Flag

type (
	empty   struct{}
	Flag    uint
	flagSet map[Flag]empty

	Context struct {
		workingDir string
		format     richtext.Format
		flags      flagSet
		regular    func(format string, a ...interface{}) (int, error)
		green      func(format string, a ...interface{}) (int, error)
		bold       func(format string, a ...interface{}) (int, error)
	}

	//Returned when the execution was started correctly, but program returned an error
	Error struct {
		flags    flagSet
		ExitCode int
		CmdLine  string
		StdErr   string
	}
)

const (
	_                 Flag = iota
	MustPanic              //panic on error
	MustExit               //Exit on error
	Warn                   //Errors get a 'Warning' level output
	NoAnnotate             //When outputting stdout/stderr, the output is passed through unchanged
	StdErrInResult         //StdErr is included in 'output'
	PassThrough            //Alias for PassThroughStdOut, PassThroughStdErr & PassThroughStdIn
	PassThroughStdOut      //Passes through stdout, output will always be an empty string
	PassThroughStdErr      //StdErr passthrough, error will only contain exit code.
	PassThroughStdIn
	TrimSpace
	FirstLine
	Strict //does 'set -eu -o pipefail'
	Verbose
)

func (fs flagSet) Checked(flag Flag) bool {
	_, ok := fs[flag]

	return ok
}

func newContext(workingDir string, format richtext.Format, flags ...Flag) *Context {
	c := &Context{workingDir: workingDir, format: format, flags: flagSet{}}

	for _, flag := range flags {
		c.flags[flag] = empty{}
	}

	if c.flags.Checked(PassThrough) {
		c.flags[PassThroughStdIn] = empty{}
		c.flags[PassThroughStdOut] = empty{}
		c.flags[PassThroughStdErr] = empty{}
	}

	if c.flags.Checked(Verbose) {
		c.green = format.MakePrintf(richtext.Green, richtext.None)
		c.bold = format.MakePrintf(richtext.None, richtext.None, richtext.Bold)
		c.regular = format.MakePrintf(richtext.None, richtext.None)

	}

	if c.workingDir == "" || c.workingDir == "." {
		cwd, err := os.Getwd()
		if err != nil {
			panic(fmt.Errorf("Can't get current working directory: %s", err.Error()))
		}
		c.workingDir = cwd
	}
	return c
}

func New(workingDir string, format richtext.Format, flags ...Flag) *Context {
	if err := Check(); err != nil {
		panic(err)
	}
	return newContext(workingDir, format, flags...)
}

func (c *Context) Execf(format string, a ...interface{}) (string, string, error) {
	return c.PipeExecf(nil, format, a...)
}

func (c *Context) PipeExecf(r io.Reader, format string, a ...interface{}) (string, string, error) {
	cmdLine := fmt.Sprintf(format, a...)

	var cmd *exec.Cmd
	if c.flags.Checked(Strict) {
		cmd = exec.Command("sh", "-c", "set -eu -o pipefail; "+cmdLine)
	} else {
		cmd = exec.Command("sh", "-c", cmdLine)
	}

	cmd.Dir = c.workingDir

	if r != nil {
		pipeIn, err := cmd.StdinPipe()
		//Can only fail if pipe is already set, or process started
		if err != nil {
			panic(err)
		}

		go func() {
			io.Copy(pipeIn, r)
			pipeIn.Close()
		}()
	} else if c.flags.Checked(PassThroughStdIn) {
		cmd.Stdin = os.Stdin
	}

	if c.flags.Checked(Verbose) {
		c.regular("%s", c.workingDir)
		c.green(" $ ")
		c.bold("%s", cmdLine)
		c.regular("\n")
	}

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer

	if c.flags.Checked(PassThroughStdOut) {
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdout = &stdoutBuf
	}

	if c.flags.Checked(PassThroughStdErr) {
		cmd.Stderr = os.Stderr
	} else if c.flags.Checked(StdErrInResult) {
		cmd.Stderr = &stdoutBuf
	} else {
		cmd.Stderr = &stderrBuf
	}

	err := c.newError(cmdLine, &stderrBuf, cmd.Run())
	if err != nil {
		if c.flags.Checked(MustExit) {
			c.format.ErrorLine(err.Error())
			os.Exit(1)
		} else if c.flags.Checked(MustPanic) {
			panic(err.Error())
		} else if c.flags.Checked(Warn) {
			c.format.WarningLine(err.Error())
		}
	}

	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	if c.flags.Checked(TrimSpace) {
		stdout = strings.TrimSpace(stdout)
		stderr = strings.TrimSpace(stderr)
	}

	if c.flags.Checked(FirstLine) {
		stdout = strings.Replace(strings.SplitN(stdout, "\n", 2)[0], "\r", "", -1)
	}

	return stdout, stderr, err
}

func (c *Context) newError(cmdLine string, errorBuf *bytes.Buffer, cmdErr error) error {
	if cmdErr == nil {
		return nil
	}
	exitErr, isExit := cmdErr.(*exec.ExitError)
	if !isExit {
		return cmdErr
	}

	//Windows and Linux only
	waitStatus, isWaitStatus := exitErr.Sys().(syscall.WaitStatus)
	if !isWaitStatus {
		return fmt.Errorf("exec error was of type ExitError, but could not get WaitStatus")
	}

	return &Error{
		flags:    c.flags,
		ExitCode: waitStatus.ExitStatus(),
		CmdLine:  cmdLine,
		StdErr:   errorBuf.String(),
	}
}

func (e *Error) Error() string {
	return e.warnText(false)
}

func (e *Error) warnText(verbose bool) string {
	if e.flags.Checked(NoAnnotate) {
		return e.StdErr
	}
	lines := []string{}
	if e.flags.Checked(Verbose) {
		//cmdLine would have already been output
		lines = append(lines, fmt.Sprintf("  returned %d", e.ExitCode))
	} else {
		lines = append(lines, fmt.Sprintf("%s returned %d", e.CmdLine, e.ExitCode))
	}

	stderr := indentLines(e.StdErr, " > ")

	if len(stderr) > 0 {
		lines = append(lines, stderr)
	}

	return strings.Join(lines, "\n")
}
