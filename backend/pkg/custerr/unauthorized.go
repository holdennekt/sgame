package custerr

type UnauthorizedErr struct {
	Msg string
}

func NewUnauthorizedErr(msg string) UnauthorizedErr {
	return UnauthorizedErr{Msg: msg}
}

func (e UnauthorizedErr) Error() string {
	return e.Msg
}
