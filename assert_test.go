package ttheader

import "testing"

func assert(t *testing.T, cond bool, val ...interface{}) {
	if !cond {
        args := []interface{}{"assertion failed"}
		if len(val) > 0 {
			args = append(args, val...)
		}
        t.Fatal(args)
	}
}
