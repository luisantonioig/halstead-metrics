package goo

// Token representa un token léxico
type Token int

const (
	// Tokens especiales
	ILLEGAL Token = iota
	EOF
	WS

	// literales
	IDENT // identificador

	// Caracteres especiales
	ASTERISK // *
	COMMA    // ,

	// Palabras reservadas
	SELECT
	FROM
)
