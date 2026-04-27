package cmd

import (
	"context"
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

func (cmd *Command) SetStderr(stderr *os.File) {
	cmd.inner.Stderr = stderr
}

func (cmd *Command) SetStdout(stdout *os.File) {
	cmd.inner.Stdout = stdout
}

func (cmd *Command) SetDir(dir string) {
	cmd.inner.Dir = dir
}

func (cmd *Command) Run(ctx context.Context) error {
	return cmd.inner.Run()
}
