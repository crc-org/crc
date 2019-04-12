package output

import "fmt"

func Out(args ...interface{}) {
	fmt.Println(args...)
}

func OutF(s string, args ...interface{}) {
	fmt.Printf(s, args...)
}
