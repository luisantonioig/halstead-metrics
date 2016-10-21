package main

import (
    "github.com/antonio/halstead/goo"
    "fmt"
    "strings"
    "io/ioutil"
    //"bytes"
    //"log"
    "os"
    "github.com/sourcegraph/syntaxhighlight"
    "reflect"
)


func main(){
    //De una consulta de sql dice cuales son las columnas
    //y cual es la tabla a las que se les hace la consulta.
    //También dice si existe un error sintáctico
    var dirs []string
    if len(os.Args) > 1 {
        dirs = os.Args[1:]
    } else {
        dirs = []string{"."}
    }
    _ , err:= goo.NewParser(strings.NewReader("SELECT SELEC from mi_tabla")).Parse()
    if err != nil{
        fmt.Println(err)
    }else{
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
    
        _, err4 := syntaxhighlight.AsHTML(dat1)
        if err4 != nil {
            fmt.Println(err4)
            os.Exit(1)
        }

    


}
func check(e error) {
    if e != nil {
        panic(e)
    }
}
