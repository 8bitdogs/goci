package cmd

import (
	"context"
	"io"
	"os"
	"os/exec"
)

type Command struct {
	inner *exec.Cmd
}

func NewCommand(name string, args []string) *Command {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	return &Command{inner: cmd}
}

func (cmd *Command) SetStderr(w io.Writer) {
	cmd.inner.Stderr = w
}

func (cmd *Command) SetStdout(w io.Writer) {
	cmd.inner.Stdout = w
}

func (cmd *Command) SetDir(dir string) {
	cmd.inner.Dir = dir
}

func (cmd *Command) Run(ctx context.Context) error {
	return cmd.inner.Run()
}
