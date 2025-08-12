package custerr

type ConflictErr struct {
	Msg string
}

func NewConflictErr(msg string) ConflictErr {
	return ConflictErr{Msg: msg}
}

func (e ConflictErr) Error() string {
	return e.Msg
}
