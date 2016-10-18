// Package syntaxhighlight provides syntax highlighting for code. It currently
// uses a language-independent lexer and performs decently on JavaScript, Java,
// Ruby, Python, Go, and C.
package syntaxhighlight

import (
	"bytes"
	"io"
	"math"
	"text/scanner"
	"text/template"
	"unicode"
	//"unicode/utf8"
    "fmt"
	"github.com/sourcegraph/annotate"
)

// Kind represents a syntax highlighting kind (class) which will be assigned to tokens.
// A syntax highlighting scheme (style) maps text style properties to each token kind.
type Kind uint8
var totalOperators = 0
var differentOperators = 0
var totalOperands = 0
var differentOperands = 0
var comments = 0

var tokens []string
var operators map[string]int
var operands map[string]int

var wait_eq = false
var wait_eqq = false

var state = "contando"
var func_call = 0
var var_call = 0

var variable = ""

var i = 0

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

//go:generate GoStringer -type=Kind

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
	fmt.Println("results",func_call,var_call)
		i++
		tokens = append(tokens,tokText)
	
	switch kind {
	case String:
		func_call=0
		var_call=0
		variable = ""
	    operands[tokText]++
	    totalOperands = totalOperands + 1
	    fmt.Println("Cadena\t",tokText)
	    return c.String
	case Keyword:
		func_call=0
		var_call=0
		variable = ""
	    operators[tokText]++
	    totalOperators = totalOperators + 1
	    fmt.Println("Keyword\t",tokText)
	    return c.Keyword
	case Comment:
		func_call=0
		var_call=0
		variable = ""
	    comments++
	    fmt.Println("Comentario")
	    return c.Comment
	case Type:
		if var_call == 2{
			var_call++
			operands[variable]--
			if operands[variable] == 0{
				delete(operands,variable)
			}
			operands[variable+"."+tokText]++
			variable = variable+"."+tokText
			break;
		}
		/*if func_call == 2{
			fmt.Println("Llamada a una funcion")
			func_call = 0
			var_call = 0
			operands[variable]--
			totalOperands--
			if operands[variable] == 0{
				delete(operands,variable)
			}
			operators[variable+"."+tokText+"()"]++
			totalOperators++
		    variable = ""
			break
		}else{
		    func_call=0
		    var_call=0
		    variable = ""
		}*/
		var_call++
		variable = tokText
	    operands[tokText]++
	    totalOperands = totalOperands + 1
	    fmt.Println("Tipo\t",tokText)
	    return c.Type
	case Literal:
		func_call=0
		var_call=0
		variable = ""
	    fmt.Println("Literal\t",tokText)
	    return c.Literal
	case Punctuation:
		if tokText=="("&&(var_call==3||(var_call==1 && tokens[i-3] != "func")){
			operands[variable]--
			totalOperands--
			if operands[variable] == 0{
				delete(operands,variable)
			}
			operators[variable+"()"]++
			totalOperators++
		}
		if(tokText=="."){
			func_call++
			var_call++
		}else{
		    func_call=0
		    var_call=0
		    variable = ""
		}
	    if(tokText!="}" && tokText!="]" && tokText!=")" && tokText!="." && tokText!="{" && tokText!="("){
	        operators[tokText]++
	        totalOperators = totalOperators + 1
                if(tokText==":"){
               	    wait_eq = true
                }
                if(tokText =="=" && wait_eq){
            	    operators[":"]--
            	    operators["="]--
            	    character:=":"+"="
            	    operators[character]++
            	if operators[":"] == 0{
            	    delete(operators,":")
            	}
            	if operators["="] == 0{
            	    delete(operators,"=")
            	}
            	operators["declaracion"]++
            	totalOperators = totalOperators - 1
            	wait_eq = false
            }
            if(tokText=="!"){
               	wait_eqq = true
            }
            if(tokText =="=" && wait_eqq){
            	operators["!"]--
            	operators["="]--
            	operators["comparacion"]++
            	totalOperators = totalOperators - 1
            	wait_eqq = false
            }
		    fmt.Println("Punt\t",tokText)
        }
		return c.Punctuation
	case Plaintext:
		func_call++
		var_call++
		if var_call ==3{
			func_call = 0
			var_call = 0
			fmt.Println("Llamada a una variable")
			operands[variable]--
			if operands[variable] == 0{
				delete(operands,variable)
			}
			operands[variable+"."+tokText]++
		    variable = ""
			break
		}
		if var_call ==1{
			variable = tokText
		}
		operands[tokText]++
		totalOperands = totalOperands + 1
		fmt.Println("Text plan\t",tokText)
	    return c.Plaintext
	case Tag:
		//fmt.Println("Etiqueta")
	    return c.Tag
	case HTMLTag:
		//fmt.Println("Etiqueta html")
		return c.HTMLTag
	case HTMLAttrName:
		//fmt.Println("Atributo html")
	    return c.HTMLAttrName
	case HTMLAttrValue:
		//fmt.Println("Valor html")
	    return c.HTMLAttrValue
	case Decimal:
		operands[tokText]++
		totalOperands = totalOperands + 1
		fmt.Println("Decimal\t",tokText)
		return c.Decimal
	}
	fmt.Println("El tamaño de tokens son:",len(tokens))
	i--
	tokens = tokens[:len(tokens)-1]
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
	operators = make(map[string]int)
    operands = make(map[string]int)
	for tok != scanner.EOF {
		tokText := s.TokenText()
		err := p.Print(w, tokenKind(tok, tokText), tokText)
		if err != nil {
			return err
		}

		tok = s.Scan()
		if len(tokens) > 0{
		    fmt.Println("-----------------------------",tokens[i-1])
		}
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
	differentOperands = len(operands)
	differentOperators = len(operators)

    fmt.Println("Operators\n--------------------------------")
	for key, value := range operators {
        fmt.Println("Key:", key, "Value:", value)
    }
    fmt.Println("Operands\n--------------------------------")
    for key, value := range operands {
        fmt.Println("Key:", key, "Value:", value)
    }
	fmt.Printf("Existen %d operadores diferentes", len(operators))
	fmt.Println("")
	fmt.Printf("Existen %d operandos diferentes", len(operands))
	fmt.Println("")
	fmt.Printf("El codigo tiene %d comentarios, %d operandos y %d operadores",comments,totalOperands,totalOperators)
	fmt.Println("")
	/*differentOperators = 10
	differentOperands = 7
	totalOperands = 15
	totalOperators = 16

    differentOperators = 31
	differentOperands = 13
	totalOperands = 15
	totalOperators = 109*/

	//logarithm1 = math.Log2(float64(differentOperators))
	//logarithm2 = math.Log2(float64(differentOperands))
	programVocabulary := differentOperands + differentOperators
	programLength := totalOperands + totalOperators
	var hola = (float64(differentOperands)*(math.Log2(float64(differentOperands))))
	fmt.Println(hola,"                                  Hoooola")
    calculatedProgramLength := (float64(differentOperators)*(math.Log2(float64(differentOperators))))+(float64(differentOperands)*(math.Log2(float64(differentOperands))))
    volume := float64(programLength) * math.Log2(float64(programVocabulary))
    difficulty := (float64(differentOperators)/2)*(float64(totalOperands)/float64(differentOperands))
    effort := difficulty * volume
    timeRequiredToProgram := effort / 18
    numberOfDeliveredBugs := math.Pow(effort,2.0/3.0) / 3000

    fmt.Printf("El tamaño calculado del programa es %f y el volumen es %f\n",calculatedProgramLength,volume)
    fmt.Printf("La dificultad del programa es %f\n",difficulty)
    fmt.Printf("El esfuerzo del programa es %f\n",effort)
    fmt.Printf("El tiempo requerido para programar es %f\n",timeRequiredToProgram)
    fmt.Printf("El numero de bugs es %f\n",numberOfDeliveredBugs)
    fmt.Println("",)
    fmt.Println("",)
    for i := 1; i <= 52 ; i++{
    	for j := 1; j <= 52 ; j++{
    		tam := (float64(i)*(math.Log2(float64(i))))+(float64(j)*(math.Log2(float64(j))))
    		if tam > 48 && tam < 49{
    			fmt.Printf("i = %d\n",i)
    	    	fmt.Printf("j = %d\n",j)
    			fmt.Printf("El tamaño calculado es %f\n",tam)
    		}else{
    			//fmt.Printf("El tamaño calculado es %f\n",tam)
    		}
    	}
    }


    

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
		}else{
			return Type
		}
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
