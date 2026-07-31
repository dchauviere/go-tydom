package tydom

type Error struct {
	Msg string
	Err error
}

func (e *Error) Error() string {
	var errStr string
	if e.Err != nil {
		errStr = " : " + e.Err.Error()
	}

	return e.Msg + errStr
}
