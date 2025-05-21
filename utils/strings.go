package utils

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrStringsParseInt = errors.New(fmt.Sprintf("strings parse int error"))
)

func StrToInt(str string) (int, error) {
	num, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, ErrStringsParseInt
	} else {
		return int(num), nil
	}
}
