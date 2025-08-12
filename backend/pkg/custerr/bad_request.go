package custerr

type BadRequestErr struct {
	Msg string
}

func NewBadRequestErr(msg string) BadRequestErr {
	return BadRequestErr{Msg: msg}
}

func (e BadRequestErr) Error() string {
	return e.Msg
}
