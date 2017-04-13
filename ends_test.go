package halstead
import (
"testing"
"io/ioutil"
"strconv"
"fmt"
)

func TestCases(t *testing.T) {
	en1, _ := ioutil.ReadFile("testdata/ejem_01.go")
	en2, _ := ioutil.ReadFile("testdata/ejem_02.go")
	en3, _ := ioutil.ReadFile("testdata/ejem_03.go")
	en4, _ := ioutil.ReadFile("testdata/ejem_04.go")
	en5, _ := ioutil.ReadFile("testdata/ejem_05.go")
	en6, _ := ioutil.ReadFile("testdata/ejem_06.go")
	en7, _ := ioutil.ReadFile("testdata/ejem_07.go")
	en8, _ := ioutil.ReadFile("testdata/ejem_08.go")
	en9, _ := ioutil.ReadFile("testdata/ejem_09.go")
	en0, _ := ioutil.ReadFile("testdata/ejem_10.go")
    cases := []struct {
		in, want []byte
	}{
		{en1 ,[]byte("4 3 4 4")},
		{en2 ,[]byte("6 5 7 6")},
		{en3 ,[]byte("6 5 6 6")},
		{en4 ,[]byte("6 5 6 6")},
		{en5 ,[]byte("4 4 4 5")},
		{en6 ,[]byte("9 7 13 10")},
		{en7 ,[]byte("9 7 12 10")},
		{en8 ,[]byte("9 10 14 17")},
		{en9 ,[]byte("12 9 15 15")},
		{en0 ,[]byte("8 6 13 11")},
	}
	for _, c := range cases {
		operators,operands,toperators,toperands,_ := AsHTML(c.in)
		fmt.Print(toperators)
		fmt.Print(toperands)
		got := strconv.Itoa(len(operators))+ " "+ strconv.Itoa(len(operands)) + " "+
		strconv.Itoa(toperators)+ " "+strconv.Itoa(toperands)
		if string(got) != string(c.want) {
			t.Errorf("AsHMTL(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
