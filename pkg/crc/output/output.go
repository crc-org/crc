package output

import (
	"fmt"
)

func Out(s string, args ... interface{}) {
	fmt.Println(fmt.Sprintf(s, args...))
}
