package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/desal/richtext"
	"github.com/stretchr/testify/assert"
)

func TestGoGrepGrep(t *testing.T) {
	output := NewStdOutput(richtext.Ansi())
	ctx := NewContext("", output)
	res, err := ctx.Execf("go help |grep -v testing|grep description|wc -l")
	assert.Equal(t, nil, err)
	assert.Equal(t, []byte{50, 10}, []byte(res))
}

func TestGoGrefGrep(t *testing.T) {
	output := NewStdOutput(richtext.Ansi())
	ctx := NewContext("", output, Warn)
	res, err := ctx.Execf("go help |gref -v testing|grep description|wc -l")
	assert.NotEqual(t, nil, err)
	fmt.Println(res)
	fmt.Println(err)
}

func TestGoGrepErrGrep(t *testing.T) {
	output := NewStdOutput(richtext.Ansi())
	ctx := NewContext("", output, Warn)
	res, err := ctx.Execf("go help |grep --fail|grep description|wc -l")
	assert.NotEqual(t, nil, err)
	fmt.Println(res)
	fmt.Println(err)
}

func testGoGrepErrGrepMust(t *testing.T) {
	output := NewStdOutput(richtext.Ansi())
	ctx := NewContext("", output, Must)
	res, err := ctx.Execf("go help |grep --fail|grep description|wc -l")
	assert.NotEqual(t, nil, err)
	fmt.Println(res)
	fmt.Println(err)
}

func TestGoGrepErrGrepMust(t *testing.T) {
	if os.Getenv("TEST_EXIT") == "1" {
		testGoGrepErrGrepMust(t)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestGoGrepErrGrepMust")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	err := cmd.Run()
	e, ok := err.(*exec.ExitError)
	assert.Equal(t, true, ok)
	assert.Equal(t, false, e.Success())
}
