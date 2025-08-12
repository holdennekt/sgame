package custerr

type ForbiddenErr struct {
	Msg string
}

func NewForbiddenErr(msg string) ForbiddenErr {
	return ForbiddenErr{Msg: msg}
}

func (e ForbiddenErr) Error() string {
	return e.Msg
}
