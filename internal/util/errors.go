package util

import (
    "errors" // Importa o pacote padrão de erros do Go
)

// ErrNotFound é um erro padrão usado para indicar que um recurso não foi encontrado.
// É uma boa prática definir erros específicos como variáveis para que possam ser
// facilmente comparados usando errors.Is().
var ErrNotFound = errors.New("resource not found")

// ErrInvalidInput é um erro padrão usado para indicar que os parâmetros de entrada
// para uma função ou requisição são inválidos.
var ErrInvalidInput = errors.New("invalid input parameter")

// IsNotFound verifica se o erro fornecido (ou qualquer erro em sua cadeia)
// é o erro ErrNotFound.
// Isso é útil para lidar com erros de forma mais precisa, como retornar um StatusNotFound
// em um handler HTTP.
func IsNotFound(err error) bool {
    return errors.Is(err, ErrNotFound)
}

