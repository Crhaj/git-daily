package common

import (
	"bytes"
	"strconv"
)

func ParseBoolFromBytes(byteSlice []byte) (bool, error) {
	strVal := string(bytes.TrimSpace(byteSlice))
	return strconv.ParseBool(strVal)
}
