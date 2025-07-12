package timeutil

import (
	"math"
	"time"
)

func FromTo(from, to time.Time) []time.Time {
	daysAmount := int(math.Ceil(to.Sub(from).Hours() / 24))

	var dates = make([]time.Time, daysAmount)

	for i := 0; i < daysAmount; i++ {
		dates[i] = from.Add(time.Duration(i) * time.Hour * 24)
	}

	return dates
}
