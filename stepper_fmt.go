package progressbar

import (
	"strconv"
	"time"
)

func fltfmt(f float64) string {
	return strconv.FormatFloat(f, 'f', 1, 64) + "%"
}
func durfmt(d time.Duration) string {
	return d.String()
}
