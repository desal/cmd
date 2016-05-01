package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/desal/richtext"
	"github.com/mattn/go-shellwords"
)

type (
	Flag uint

	Output interface {
		Verbose(format string, a ...interface{})
		Warning(format string, a ...interface{})
		Error(format string, a ...interface{})
		Format() richtext.Format
	}

	Context struct {
		workingDir string
		output     Output
		must       bool
		warn       bool
		green      func(format string, a ...interface{}) string
		bold       func(format string, a ...interface{}) string
	}
)

const (
	Must Flag = iota
	Warn
)

var (
	shellParser *shellwords.Parser
)

func init() {
	shellParser = shellwords.NewParser()
	shellParser.ParseBackSlash = false
}

func NewContext(workingDir string, output Output, flags ...Flag) *Context {
	result := &Context{workingDir: workingDir, output: output}
	for _, flag := range flags {
		switch flag {
		case Must:
			result.must = true
		case Warn:
			result.warn = true
		}
	}
	format := output.Format()
	result.green = format.String(richtext.Green, richtext.None)
	result.bold = format.String(richtext.None, richtext.None, richtext.Bold)
	return result
}

//TODO move from []string to a command struct

func splitCmd(s string) ([]*exec.Cmd, error) {
	args, err := shellParser.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse '%s': %s", s, err.Error())
	}
	cmds := []*exec.Cmd{}
	next := []string{}
	for _, e := range args {
		if e == "|" {
			cmds = append(cmds, exec.Command(next[0], next[1:]...))
			next = nil
		} else {
			next = append(next, e)
		}
	}
	cmds = append(cmds, exec.Command(next[0], next[1:]...))

	return cmds, nil
}

func quoteJoin(s []string) string {
	quoted := []string{}
	for _, e := range s {
		if strings.Contains(e, " ") {
			quoted = append(quoted, `"`+e+`"`)
		} else {
			quoted = append(quoted, e)
		}
	}
	return strings.Join(quoted, " ")
}

func indentLines(s string, indent string) string {
	lines := strings.Split(s, "\n")
	outLines := []string{}
	for _, line := range lines {
		if line != "" {
			outLines = append(outLines, indent+line)
		}
	}
	return strings.Join(outLines, "\n")
}

func (c *Context) Execf(format string, a ...interface{}) (string, error) {
	if c.workingDir == "" || c.workingDir == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("Can't get working directory: %s", err.Error())
		}
		c.workingDir = cwd
	}

	cmdline := fmt.Sprintf(format, a...)
	cmds, err := splitCmd(cmdline)
	if err != nil {
		return "", err
	}
	errorBufs := make([]bytes.Buffer, len(cmds))

	for i := 0; i < len(cmds)-1; i++ {
		outPipe, err := cmds[i].StdoutPipe()
		if err != nil {
			return "", fmt.Errorf("StdoutPipe Error %v", err.Error())
		}

		errPipe, err := cmds[i].StderrPipe()
		if err != nil {
			return "", fmt.Errorf("StderrPipe Error %v", err.Error())
		}

		inPipe, err := cmds[i+1].StdinPipe()
		if err != nil {
			return "", fmt.Errorf("StdinPipe Error %v", err.Error())
		}

		go func() {
			io.Copy(inPipe, outPipe)
			inPipe.Close()
		}()

		go func(i int) {
			errorBufs[i].ReadFrom(errPipe)
		}(i)
	}

	var result []byte
	msgs := make([]string, len(cmds))
	errors := make([]error, len(cmds))

	for i, cmd := range cmds {
		if i != 0 {
			msgs[i] += c.bold(" -> ")
		}
		msgs[i] += c.workingDir
		msgs[i] += c.green(" $ ")
		msgs[i] += c.bold(quoteJoin(cmds[i].Args))
		cmd.Dir = c.workingDir
		if i != len(cmds)-1 {
			err := cmd.Start()
			if err != nil {
				return "", fmt.Errorf("Failed to start %s: %s", quoteJoin(cmds[i].Args), err.Error())
			}
		} else {
			errPipe, err := cmds[i].StderrPipe()
			if err != nil {
				return "", fmt.Errorf("StderrPipe Error %v", err.Error())
			}
			go func() {
				errorBufs[i].ReadFrom(errPipe)
			}()

			result, errors[i] = cmd.Output()

		}
	}

	for i := 0; i < len(cmds); i++ {
		if i != len(cmds)-1 {
			errors[i] = cmds[i].Wait()
		}

		if c.warn && errors[i] != nil {
			warnMsg := msgs[i] + " returned error " + errors[i].Error()
			stderr := indentLines(string(errorBufs[i].Bytes()), " > ")
			if len(stderr) > 0 {
				warnMsg += "\n" + stderr
			}
			c.output.Warning(warnMsg)

		} else if c.must && errors[i] != nil {
			errMsg := msgs[i] + " returned error " + errors[i].Error()
			stderr := indentLines(string(errorBufs[i].Bytes()), " > ")
			if len(stderr) > 0 {
				errMsg += "\n" + stderr
			}
			c.output.Error(errMsg)
			os.Exit(1)

		} else {
			c.output.Verbose(msgs[i])
		}
	}

	for i := 0; i < len(cmds); i++ {
		if errors[i] != nil {
			return string(result), fmt.Errorf("%s returned error %s", quoteJoin(cmds[i].Args), errors[i])
		}
	}

	return string(result), nil
}
