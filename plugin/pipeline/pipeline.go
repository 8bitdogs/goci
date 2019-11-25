package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const defaultPipelineFile = ".pipeline"

var (
	ErrPipelineFileNotFound = errors.New("pipeline file not found")
)

type Step struct {
	Command Command `json:"command"`
}

type Pipeline struct {
	projectDir string
}

func New(dir string) *Pipeline {
	return &Pipeline{
		projectDir: dir,
	}
}

func (p *Pipeline) Run(ctx context.Context) error {
	// applying steps from .pipeline file
	pf := filepath.Join(p.projectDir, defaultPipelineFile)
	file, err := os.Open(pf)
	if err != nil {
		return err
	}
	steps := make([]Step, 0)
	err = json.NewDecoder(file).Decode(&steps)
	if err != nil {
		return err
	}
	for _, step := range steps {
		if err := step.Command.run(ctx, p.projectDir); err != nil {
			return &Error{Err: err, Step: step}
		}
	}
	return nil
}
