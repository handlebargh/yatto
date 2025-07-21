package items

import "time"

func IsToday(t *time.Time) bool {
	if t == nil {
		return false
	}

	now := time.Now()
	y1, m1, d1 := t.Date()
	y2, m2, d2 := now.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
