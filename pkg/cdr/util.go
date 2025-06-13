package cdr

import (
	"strconv"
)

func strToUint(s string) uint {
	int, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return uint(int)
}
