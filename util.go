package cmd

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/desal/richtext"
)

var (
	shellErr     error
	shellChecked bool
	shellMutex   sync.Mutex
)

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

func Check() error {
	shellMutex.Lock()
	defer shellMutex.Unlock()
	if shellChecked {
		return shellErr
	}
	cmd := exec.Command("sh", "--version")
	_, shellErr = cmd.Output()
	if shellErr != nil {
		return shellErr
	}

	var stdout string
	var stderr string
	stdout, stderr, shellErr = newContext("", richtext.Silenced()).Execf("echo 'stdout'; echo 'stderr' 1>&2")
	if shellErr != nil {
		return shellErr
	}

	if !strings.Contains(stdout, "stdout") {
		shellErr = fmt.Errorf("ShellExec did not capture stdout correctly")
	} else if !strings.Contains(stderr, "stderr") {
		shellErr = fmt.Errorf("ShellExec did not capture stderr correctly")
	}

	return shellErr
}
