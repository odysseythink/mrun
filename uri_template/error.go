package uritemplate

import (
	"fmt"
)

func errorf(pos int, format string, a ...interface{}) error {
	msg := fmt.Sprintf(format, a...)
	return fmt.Errorf("uritemplate:%d:%s", pos, msg)
}
