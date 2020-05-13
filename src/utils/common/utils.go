package common

import (
	"bytes"
	"strconv"
)

func ParseBoolFromBytes(byteSlice []byte) (bool, error) {
	strVal := ParseStringFromBytes(byteSlice)
	return strconv.ParseBool(strVal)
}

func ParseStringFromBytes(byteSlice []byte) string {
	return string(bytes.TrimSpace(byteSlice))
}
