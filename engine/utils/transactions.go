package utils

import (
	"strings"
)

const efMarker = "0000000000ef"

// IsEf checks if the transaction hex is in Extended Format
func IsEf(txHex string) bool {
	return len(txHex) > 20 && strings.EqualFold(txHex[8:20], efMarker)
}
