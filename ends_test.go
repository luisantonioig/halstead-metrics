package halstead
import (
"testing"
"io/ioutil"
"strconv"
)

func TestTimeConsuming(t *testing.T) {
	en1, _ := ioutil.ReadFile("testdata/ejem_01.go")
    cases := []struct {
		in, want []byte
	}{
		{en1 ,[]byte("4 3 4 4")},
	}
	for _, c := range cases {
		operators,operands,toperators,toperands,_ := AsHTML(c.in)
		got := strconv.Itoa(len(operators))+ " "+ strconv.Itoa(len(operands)) + " "+
		strconv.Itoa(toperators)+ " "+strconv.Itoa(toperands)
		if string(got) != string(c.want) {
			t.Errorf("AsHMTL(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
