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
		{en1 ,[]byte("3 4")},
	}
	for _, c := range cases {
		operators,operands,_ := AsHTML(c.in)
		got := strconv.Itoa(operators)+ " "+ strconv.Itoa(operands)
		if string(got) != string(c.want) {
			t.Errorf("AsHMTL(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
