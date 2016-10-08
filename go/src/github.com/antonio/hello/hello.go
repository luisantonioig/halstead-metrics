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
    _ , err:= goo.NewParser(strings.NewReader("SELECT SELEC from mi_tabla")).Parse()
    if err != nil{
        fmt.Println(err)
    }else{
        //fmt.Println("La tabla es",stmt.TableName,"y las columnas son",stmt.Fields)
    }

    src := []byte(`
/* hello, world! */
var a = 3;

// b is a cool function
function b() {
  return 7;
}`)
    reflect.TypeOf(src)

    dat1, err2 := ioutil.ReadFile("/home/vagrant/archivos/codigo.txt")
    check(err2)
    //fmt.Print(string(dat1))
    
        _, err4 := syntaxhighlight.AsHTML(dat1)
        if err4 != nil {
            fmt.Println(err4)
            os.Exit(1)
        }

        //fmt.Println(string(highlighted))
        fmt.Println("Hola")

    
/*
    //Este código es para abrir un archivo con una ruta e imprimir el contenido del archivo
    //Si existe algun error dice el error que ocurrió

    dat, err := ioutil.ReadFile("/home/antonio/go/src/github.com/antonio/sql/consulta.sql")
    check(err)
    fmt.Print(string(dat))

    //Del archivo abierto en la parte anterior leé el contenido para ver si tiene alguna
    //consulta de SQL con la palabra SELECT si si la tiene escribe cuales son las columnas
    //y cual es la tabla, sino, manda a pantalla el error
    stmt1 , err1 := goo.NewParser(bytes.NewReader(dat)).Parse()
    if err1 != nil{
        fmt.Println(err1)
    }else{
        fmt.Println("La tabla es",stmt1.TableName,"y las columnas son",stmt1.Fields)
    }

    //Este código nos muestra el nombre de los archivos que contiene una carpeta
    // https://golang.org/pkg/io/ioutil/
    files, err := ioutil.ReadDir("/home/antonio/go/src/github.com/antonio/")
    if err != nil {
        log.Fatal(err)
    }
    for _, file := range files {
        fmt.Println(file.Name())
    }

*/


}
func check(e error) {
    if e != nil {
        panic(e)
    }
}
