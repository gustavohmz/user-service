package errors

import "fmt"

// Errores de dominio
var (
	ErrUserNotFound      = fmt.Errorf("usuario no encontrado")
	ErrUserAlreadyExists = fmt.Errorf("usuario ya existe")
	ErrInvalidCredentials = fmt.Errorf("credenciales inválidas")
	ErrUserInBlacklist   = fmt.Errorf("usuario está en lista negra")
	ErrUnauthorized      = fmt.Errorf("no autorizado")
	ErrForbidden         = fmt.Errorf("acceso prohibido")
)

// ErrorWithCode representa un error con código HTTP
type ErrorWithCode struct {
	Code    int
	Message string
	Err     error
}

func (e *ErrorWithCode) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func NewErrorWithCode(code int, message string, err error) *ErrorWithCode {
	return &ErrorWithCode{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

