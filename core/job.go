package core

import "context"

type Job interface {
	Run(context.Context) error
}
