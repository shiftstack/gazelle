package job

type ErrNoStepgraph struct {
	wrapped error
}

func (err ErrNoStepgraph) Unwrap() error {
	return err.wrapped
}

func (err ErrNoStepgraph) Error() string {
	return "no stepgraph: " + err.wrapped.Error()
}

func NewErrNoStepgraph(err error) ErrNoStepgraph {
	return ErrNoStepgraph{err}
}
