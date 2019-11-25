package pipeline

type Error struct {
	Step Step
	Err  error
}

func (e *Error) Error() string {
	return e.Err.Error()
}
