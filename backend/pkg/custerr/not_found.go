package custerr

type NotFoundErr struct {
	Msg string
}

func NewNotFoundErr(msg string) NotFoundErr {
	return NotFoundErr{Msg: msg}
}

func (e NotFoundErr) Error() string {
	return e.Msg
}
