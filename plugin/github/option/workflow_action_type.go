package option

import "strings"

var _ Option = (*workflowActionTypeOption)(nil)

const (
	WorkflowActionQueued     = "queued"
	WorkflowActionInProgress = "in_progress"
	WorkflowActionCompleted  = "completed"
)

type workflowActionTypeOption struct {
	workflowActionType string
}

func (o *workflowActionTypeOption) Apply(opts *Options) {
	opts.workflowAction = o.workflowActionType
}

func WithWorkflowActionType(workflowActionType string) Option {
	lwWAT := strings.ToLower(workflowActionType)
	switch lwWAT {
	case WorkflowActionQueued, WorkflowActionInProgress, WorkflowActionCompleted:
		return &workflowActionTypeOption{workflowActionType: lwWAT}
	default:
		panic("invalid workflow action type: " + workflowActionType)
	}
}
