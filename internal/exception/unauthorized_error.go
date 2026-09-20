// internal/exception/unauthorized_error.go
package exception

type UnauthorizedError struct {
	Message string
}

func NewUnauthorizedError(message string) UnauthorizedError {
	return UnauthorizedError{Message: message}
}

func (e UnauthorizedError) Error() string {
	return e.Message
}