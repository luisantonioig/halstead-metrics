package main

import (
    "fmt"
    "io/ioutil"
    //"bytes"
    //"log"
    "os"
    "github.com/luisantonioig/halstead-metrics"
    "reflect"
    "math"
)


func main(){
    var dirs []string
    if len(os.Args) > 1 {
        dirs = os.Args[1:]
    } else {
        dirs = []string{"."}
    }

    src := []byte(`
/* hello, world! */
var a = 3;

// b is a cool function
function b() {
  return 7;
}`)
    reflect.TypeOf(src)
    dat1, err2 := ioutil.ReadFile(dirs[0])
    check(err2)
        operators,operands,totalOperators,totalOperands ,err4 := halstead.AsHTML(dat1)
        if err4 != nil {
            fmt.Println(err4)
            os.Exit(1)
        }
        differentOperands := len(operands)
        differentOperators := len(operators)

        fmt.Println("Operators\n--------------------------------")
        for key, value := range operators {
            fmt.Println("Key:", key, "Value:", value)
        }
        fmt.Println("\n\nOperands\n--------------------------------")
        for key, value := range operands {
            fmt.Println("Key:", key, "Value:", value)
        }
        fmt.Printf("\n\nExisten %d operadores diferentes", len(operators))
        fmt.Println("")
        fmt.Printf("Existen %d operandos diferentes", len(operands))
        fmt.Println("")
        fmt.Printf("El codigo tiene %d operandos y %d operadores",totalOperands,totalOperators)
        fmt.Println("")
        programVocabulary := differentOperands + differentOperators
        programLength := totalOperands + totalOperators
        var hola = (float64(differentOperands)*(math.Log2(float64(differentOperands))))
        fmt.Println(hola)
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
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}
