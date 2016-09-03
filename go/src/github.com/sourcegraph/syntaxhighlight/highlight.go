// Package syntaxhighlight provides syntax highlighting for code. It currently
// uses a language-independent lexer and performs decently on JavaScript, Java,
// Ruby, Python, Go, and C.
package syntaxhighlight

import (
	"bytes"
	"io"
	"text/scanner"
	"text/template"
	"unicode"
	"unicode/utf8"
    "fmt"
	"github.com/sourcegraph/annotate"
)

// Kind represents a syntax highlighting kind (class) which will be assigned to tokens.
// A syntax highlighting scheme (style) maps text style properties to each token kind.
type Kind uint8
var totalOperators = 0
var totalOperands = 0
var comments = 0

const (
	Whitespace Kind = iota
	String
	Keyword
	Comment
	Type
	Literal
	Punctuation
	Plaintext
	Tag
	HTMLTag
	HTMLAttrName
	HTMLAttrValue
	Decimal
)

//go:generate gostringer -type=Kind

type Printer interface {
	Print(w io.Writer, kind Kind, tokText string) error
}

// HTMLConfig holds the HTML class configuration to be used by annotators when
// highlighting code.
type HTMLConfig struct {
	String        string
	Keyword       string
	Comment       string
	Type          string
	Literal       string
	Punctuation   string
	Plaintext     string
	Tag           string
	HTMLTag       string
	HTMLAttrName  string
	HTMLAttrValue string
	Decimal       string
	Whitespace    string
}

type HTMLPrinter HTMLConfig

// Class returns the set class for a given token Kind.
func (c HTMLConfig) Class(kind Kind,tokText string) string {
	if(tokText!=" " && tokText!="\n"){
	    fmt.Print(tokText,"\t")
    }
	switch kind {
	case String:
		totalOperands = totalOperands + 1
		fmt.Println("Cadena")
		return c.String
	case Keyword:
		totalOperators = totalOperators + 1
		fmt.Println("Palabra reservada")
		return c.Keyword
	case Comment:
		comments++
		fmt.Println("Comentario")
		return c.Comment
	case Type:
		totalOperators = totalOperators + 1
		fmt.Println("Tipo")
		return c.Type
	case Literal:
		totalOperands = totalOperands + 1
		fmt.Println("Literal")
		return c.Literal
	case Punctuation:
		totalOperators = totalOperators + 1
		fmt.Println("Signo de puntiacion")
		return c.Punctuation
	case Plaintext:
		totalOperands = totalOperands + 1
		fmt.Println("Texto plano")
		return c.Plaintext
	case Tag:
		fmt.Println("Etiqueta")
		return c.Tag
	case HTMLTag:
		fmt.Println("Etiqueta html")
		return c.HTMLTag
	case HTMLAttrName:
		fmt.Println("Atributo html")
		return c.HTMLAttrName
	case HTMLAttrValue:
		fmt.Println("Valor html")
		return c.HTMLAttrValue
	case Decimal:
		totalOperands = totalOperands + 1
		fmt.Println("Decimal")
		return c.Decimal
	}
	return ""
}

func (p HTMLPrinter) Print(w io.Writer, kind Kind, tokText string) error {
	class := ((HTMLConfig)(p)).Class(kind,tokText)
	if class != "" {
		_, err := w.Write([]byte(`<span class="`))
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, class)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(`">`))
		if err != nil {
			return err
		}
	}
	template.HTMLEscape(w, []byte(tokText))
	if class != "" {
		_, err := w.Write([]byte(`</span>`))
		if err != nil {
			return err
		}
	}
	return nil
}

type Annotator interface {
	Annotate(start int, kind Kind, tokText string) (*annotate.Annotation, error)
}

type HTMLAnnotator HTMLConfig


//Esta es la función que devuelve las etiquetas en código html con span------------------------------------------------------------------------
func (a HTMLAnnotator) Annotate(start int, kind Kind, tokText string) (*annotate.Annotation, error) {
	class := ((HTMLConfig)(a)).Class(kind,tokText)
	if class != "" {
		left := []byte(`<span class="`)
		left = append(left, []byte(class)...)
		left = append(left, []byte(`">`)...)
		return &annotate.Annotation{
			Start: start, End: start + len(tokText),
			Left: left, Right: []byte("</span>"),
		}, nil
	}
	return nil, nil
}

// DefaultHTMLConfig's class names match those of google-code-prettify
// (https://code.google.com/p/google-code-prettify/).
var DefaultHTMLConfig = HTMLConfig{
	String:        "str",
	Keyword:       "kwd",
	Comment:       "com",
	Type:          "typ",
	Literal:       "lit",
	Punctuation:   "pun",
	Plaintext:     "pln",
	Tag:           "tag",
	HTMLTag:       "htm",
	HTMLAttrName:  "atn",
	HTMLAttrValue: "atv",
	Decimal:       "dec",
	Whitespace:    "",
}


func Print(s *scanner.Scanner, w io.Writer, p Printer) error {
	tok := s.Scan()
	for tok != scanner.EOF {
		tokText := s.TokenText()
		err := p.Print(w, tokenKind(tok, tokText), tokText)
		if err != nil {
			return err
		}

		tok = s.Scan()
	}

	return nil
}

func Annotate(src []byte, a Annotator) (annotate.Annotations, error) {
	s := NewScanner(src)

	var anns annotate.Annotations
	read := 0

	tok := s.Scan()
	for tok != scanner.EOF {
		tokText := s.TokenText()

		ann, err := a.Annotate(read, tokenKind(tok, tokText), tokText)
		if err != nil {
			return nil, err
		}
		read += len(tokText)
		if ann != nil {
			anns = append(anns, ann)
		}

		tok = s.Scan()
	}

	return anns, nil
}

//Retorna en bytes en código html los tokens con etiquetas de spam
func AsHTML(src []byte) ([]byte, error) {
	var buf bytes.Buffer
	err := Print(NewScanner(src), &buf, HTMLPrinter(DefaultHTMLConfig))
	if err != nil {
		return nil, err
	}
	fmt.Printf("El codigo tiene %d comentarios, %d operandos y %d operadores",comments,totalOperands,totalOperators)
	fmt.Println("")
	return buf.Bytes(), nil
}

// NewScanner is a helper that takes a []byte src, wraps it in a reader and creates a Scanner.
func NewScanner(src []byte) *scanner.Scanner {
	return NewScannerReader(bytes.NewReader(src))
}

// NewScannerReader takes a reader src and creates a Scanner.
func NewScannerReader(src io.Reader) *scanner.Scanner {
	var s scanner.Scanner
	s.Init(src)
	s.Error = func(_ *scanner.Scanner, _ string) {}
	s.Whitespace = 0
	s.Mode = s.Mode ^ scanner.SkipComments
	return &s
}

func tokenKind(tok rune, tokText string) Kind {
	switch tok {
	case scanner.Ident:
		if _, isKW := keywords[tokText]; isKW {
			return Keyword
		}
		if r, _ := utf8.DecodeRuneInString(tokText); unicode.IsUpper(r) {
			return Type
		}
		return Plaintext
	case scanner.Float, scanner.Int:
		return Decimal
	case scanner.Char, scanner.String, scanner.RawString:
		return String
	case scanner.Comment:
		return Comment
	}
	if unicode.IsSpace(tok) {
		return Whitespace
	}
	return Punctuation
}
