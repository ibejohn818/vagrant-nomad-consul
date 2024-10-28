package cli

import (
	"errors"
	"gotooling/johnhardy.io/pkg/utils"
	"strings"
)

type ListFlag []string

func (l *ListFlag) String() string {
	return strings.Join(*l, ", ")
}

func (l *ListFlag) Set(val string) error {
	if len(*l) >= 2 {
		return errors.New("too many values")
	}

	if !utils.StrInArray(val, []string{"stderr", "stdout"}) {
		return errors.New("invalid value!")
	}

	*l = append(*l, val)

	return nil
}
