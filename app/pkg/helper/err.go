package helper

import (
	"errors"
	"fmt"
	"strings"
)

func ErrMapToError(m map[string]error) error {
	s := strings.Builder{}
	for k, v := range m {
		s.WriteString(fmt.Sprintf("%s: %v\n", k, v))
	}
	return errors.New(s.String())
}
