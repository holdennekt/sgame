package custerr

type TooManyRequestsErr struct {
	Msg string
}

func NewTooManyRequestsErr(msg string) TooManyRequestsErr {
	return TooManyRequestsErr{Msg: msg}
}

func (e TooManyRequestsErr) Error() string {
	return e.Msg
}
