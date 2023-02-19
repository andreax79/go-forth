package fcpu

type Halt struct {
}

func (e *Halt) Error() string {
	return "Halt"
}

type ExecFormatError struct {
}

func (e *ExecFormatError) Error() string {
	return "Exec format error"
}
