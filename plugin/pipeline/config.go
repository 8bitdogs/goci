package pipeline

import "strings"

type StepConfig struct {
	Name string   `json:"name" yaml:"name"`
	Cmd  string   `json:"cmd" yaml:"cmd"`
	Args []string `json:"args" yaml:"args"`
	Dir  string   `json:"dir" yaml:"dir"`
	// Timeout time.Duration `json:"timeout" yaml:"timeout"`
}

func (s *StepConfig) CmdString() string {
	return s.Cmd + " " + strings.Join(s.Args, " ")
}

type JobConfig struct {
	Name  string       `json:"name" yaml:"name"`
	Steps []StepConfig `json:"steps" yaml:"steps"`
}

type Config struct {
	Jobs []JobConfig `json:"jobs" yaml:"jobs"`
}
