package custerr

import "errors"

type InternalErr struct {
	Msg string
}

func NewInternalErr(err error) InternalErr {
	return InternalErr{Msg: errors.Join(errors.New("internal server error"), err).Error()}
}

func (e InternalErr) Error() string {
	return e.Msg
}
