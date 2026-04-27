package core

import "context"

type Pipeline interface {
	Run(context.Context) error
}
