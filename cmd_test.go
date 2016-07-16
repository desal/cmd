package cmd_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/desal/cmd"
	"github.com/desal/richtext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	o, e, err := cmd.New("", richtext.Silenced()).Execf(`
		rm -f testhelper;
		go build -o testhelper ./_testhelper
	`)

	if err != nil {
		fmt.Println("TEST SETUP FAILED")
		fmt.Println(o)
		fmt.Println(e)
		os.Exit(1)
	}

	result := m.Run()

	o, e, err = cmd.New("", richtext.Silenced()).Execf(`
		rm -f testhelper;
	`)

	if err != nil {
		fmt.Println("TEST CLEANUP FAILED")
		fmt.Println(o)
		fmt.Println(e)
		os.Exit(1)
	}

	os.Exit(result)
}
func TestExecOk(t *testing.T) {
	stdout, stderr, err := cmd.New("", richtext.Test(t)).Execf(`./testhelper "stdout" "stderr" 0`)
	assert.Contains(t, stdout, "stdout")
	assert.Contains(t, stderr, "stderr")
	assert.Nil(t, err)
}

func TestExecErr(t *testing.T) {
	stdout, stderr, err := cmd.New("", richtext.Test(t)).Execf(`./testhelper "stdout" "stderr" 1`)
	assert.Contains(t, stdout, "stdout")
	assert.Contains(t, stderr, "stderr")
	require.NotNil(t, err)
	cmdErr, ok := err.(*cmd.Error)
	require.True(t, ok)
	assert.Equal(t, cmdErr.ExitCode, 1)
}

func TestBadCommand(t *testing.T) {
	_, _, err := cmd.New("", richtext.Test(t)).Execf("aoeusnth")
	require.NotNil(t, err)
	cmdErr, ok := err.(*cmd.Error)
	require.True(t, ok)
	assert.Equal(t, cmdErr.ExitCode, 127)
}

func TestPipe(t *testing.T) {
	var rw bytes.Buffer
	rw.WriteString("a\nb\nc\n")
	stdout, _, err := cmd.New("", richtext.Test(t), cmd.TrimSpace).PipeExecf(&rw, "wc -l")
	assert.Nil(t, err)
	assert.Equal(t, stdout, "3")
}

func TestCheck(t *testing.T) {
	assert.Nil(t, cmd.Check())
	path := os.Getenv("PATH")
	os.Setenv("PATH", "")
	assert.NotNil(t, cmd.Check())
	os.Setenv("PATH", path)
	assert.Nil(t, cmd.Check())
}

func TestTrim(t *testing.T) {
	stdout, _, _ := cmd.New("", richtext.Test(t), cmd.TrimSpace).Execf(
		"./testhelper ' \n  \n \t stdout  \n  \n ' '' 0")
	assert.Equal(t, "stdout", stdout)
}

func TestFirstLine(t *testing.T) {
	ctx := cmd.New("", richtext.Test(t), cmd.FirstLine)

	stdout, _, _ := ctx.Execf("./testhelper 'one\ntwo\nthree\n' '' 0")
	assert.Equal(t, "one", stdout)

	stdout, _, _ = ctx.Execf("./testhelper '\ntwo\nthree\n' '' 0")
	assert.Equal(t, "", stdout)
	stdout, _, _ = ctx.Execf("./testhelper ' \ntwo' '' 0")
	assert.Equal(t, " ", stdout)

	ctx = cmd.New("", richtext.Test(t), cmd.FirstLine, cmd.TrimSpace)
	stdout, _, _ = ctx.Execf("./testhelper '\ntwo\nthree\n' '' 0")
	assert.Equal(t, "two", stdout)
	stdout, _, _ = ctx.Execf("./testhelper ' \ntwo' '' 0")
	assert.Equal(t, "two", stdout)

}

func TestStrict(t *testing.T) {
	ctx := cmd.New("", richtext.Test(t))
	_, _, err := ctx.Execf("./testhelper '' '' 1; ./testhelper '' '' 0")
	assert.Nil(t, err)

	_, _, err = ctx.Execf("./testhelper '' '' 1| wc -l")
	assert.Nil(t, err)

	ctx = cmd.New("", richtext.Test(t), cmd.Strict)
	_, _, err = ctx.Execf("./testhelper '' '' 1; ./testhelper '' '' 0")
	assert.NotNil(t, err)

	_, _, err = ctx.Execf("./testhelper '' '' 1| wc -l")
	assert.NotNil(t, err)

}
