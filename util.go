package lmdbcli

import "strconv"

func labelUint(label string, value uint64) []byte {
	return labelString(label, strconv.FormatUint(value, 10))
}

func labelString(label string, value string) []byte {
	return []byte(label + ": " + value)
}

func readableBytes(size uint64) string {
	if size < 1024 {
		return strconv.FormatUint(size, 10) + "B"
	}
	for _, unit := range units {
		size = size / 1024
		if size < 1024 {
			return strconv.FormatUint(size, 10) + unit
		}
	}
	return ""
}
