package progressbar

import (
	"math"
	"strconv"
	"time"
)

func fltfmt(f float64) string {
	return strconv.FormatFloat(f, 'f', 1, 64)
}

func fltfmtpercent(f float64) string {
	return strconv.FormatFloat(f*100, 'f', 1, 64) + "%"
}

func intfmt(i int64) string {
	return strconv.FormatInt(i, 10)
}

func durfmt(d time.Duration) string {
	return d.String()
}

func humanizeBytes(s float64) (value, suffix string) {
	sizes := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB"}
	base := 1024.0
	if s < 10 {
		// return fmt.Sprintf("%2.0f", s), "B"
		return strconv.FormatFloat(s, 'f', 0, 64), "B"
		// return fltfmt(s), "B"
	}

	e := math.Floor(logn(s, base))
	suffix = sizes[int(e)]
	val := math.Floor(s/math.Pow(base, e)*10+0.5) / 10

	// f := "%.0f"
	// if val < 10 {
	// 	f = "%.1f"
	// }
	// value = fmt.Sprintf(f, val)

	// precision := 1
	// if val > 10 {
	// 	precision = 0
	// }
	const precision = -1
	value = strconv.FormatFloat(val, 'f', precision, 64)
	// value = fltfmt(val)
	return
}

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}
