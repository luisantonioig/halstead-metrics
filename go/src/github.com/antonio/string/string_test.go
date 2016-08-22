package string

import "testing"

func Test(t *testing.T){
    var test = []struct{
        s,want string
    }{
        {"Backward","drawkcaB"},
        {"Hooola","aloooH"},
        {"",""},
    }
    for _, c:=range test{
        got :=Reverse(c.s)
        if got !=c.want {
            t.Errorf("Reverse(%q) == %q, want %q", c.s, got, c.want)
        }
    }
}
