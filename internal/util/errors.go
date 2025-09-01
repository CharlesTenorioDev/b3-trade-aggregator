package util

import (
	"errors"
)

var ErrNotFound = errors.New("resource not found")

var ErrInvalidInput = errors.New("invalid input parameter")

// IsNotFound verifica se o erro fornecido (ou qualquer erro em sua cadeia)
// é o erro ErrNotFound.
// Isso é útil para lidar com erros de forma mais precisa, como retornar um StatusNotFound
// em um handler HTTP.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
