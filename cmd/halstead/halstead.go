package main

import (
    "fmt"
    "io/ioutil"
    //"bytes"
    //"log"
    "os"
    "github.com/luisantonioig/halstead-metrics"
    "reflect"
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
        _,_, err4 := halstead.AsHTML(dat1)
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
