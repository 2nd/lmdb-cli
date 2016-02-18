package commands

import "strconv"

var (
	OK        = []byte("ok")
	TRUE      = []byte("true")
	FALSE     = []byte("false")
	SCAN_MORE = []byte(`"it" for more`)

	units = []string{"KB", "MB", "GB", "TB", "PB"}
)

func labelUint(label string, value uint64) []byte {
	return labelString(label, strconv.FormatUint(value, 10))
}

func labelInt(label string, value int64) []byte {
	return labelString(label, strconv.FormatInt(value, 10))
}

func labelString(label string, value string) []byte {
	return []byte(label + ": " + value)
}

func readableBytes(size int64) string {
	if size < 1024 {
		return strconv.FormatInt(size, 10) + "B"
	}
	for _, unit := range units {
		size = size / 1024
		if size < 1024 {
			return strconv.FormatInt(size, 10) + unit
		}
	}
	return ""
}
