package errors

type CIError interface {
	Error() string
	ExitCode() int
}
