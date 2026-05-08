package downsample

import (
	"time"

	"github.com/example/timeseries-downsample/pkg/common"
)

func AlignToWindow(t time.Time, window common.WindowSize) time.Time {
	d := window.Duration
	switch d {
	case time.Minute:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
	case 5 * time.Minute:
		minutes := t.Minute() / 5 * 5
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), minutes, 0, 0, t.Location())
	case time.Hour:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	case 24 * time.Hour:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	default:
		ts := t.UnixNano() / int64(d) * int64(d)
		return time.Unix(0, ts).In(t.Location())
	}
}

func NextWindow(t time.Time, window common.WindowSize) time.Time {
	return AlignToWindow(t, window).Add(window.Duration)
}
