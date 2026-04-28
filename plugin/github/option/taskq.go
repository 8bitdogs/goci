package option

import "github.com/antonmashko/taskq"

var _ Option = (*TaskQOption)(nil)

type TaskQOption struct {
	taskq *taskq.TaskQ
}

func (o *TaskQOption) Apply(opts *Options) {
	opts.taskq = o.taskq
}

func WithTaskQ(tq *taskq.TaskQ) Option {
	if tq == nil {
		const defaultTaskQSize = 100
		return &TaskQOption{taskq: taskq.New(defaultTaskQSize)}
	}
	return &TaskQOption{taskq: tq}
}
