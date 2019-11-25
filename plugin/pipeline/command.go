package pipeline

import (
	"context"
	"os"
	"os/exec"

	"goci/core"

	"github.com/8bitdogs/log"
)

type Command struct {
	Name string
	Args []string
}

func (c *Command) run(ctx context.Context, dir string) error {
	cmd := exec.Command(c.Name, c.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Dir = dir
	log.Infof("%d running cmd %s %v", core.RequestID(ctx), c.Name, c.Args)
	return cmd.Run()
}
