package faults

type BadRequestError struct {
	Msg string
}

type NotFoundError struct {
	Msg string
}

func (e BadRequestError) Error() string {
	return e.Msg
}

func (e NotFoundError) Error() string {
	return e.Msg
}

func NewBadRequest(msg string) error {
	return BadRequestError{Msg: msg}
}

func NewNotFound(msg string) error {
	return NotFoundError{Msg: msg}
}
