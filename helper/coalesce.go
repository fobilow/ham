package helper

import (
	"time"
)

func CoalesceString(s ...string) string {
	for _, v := range s {
		if v != "" {
			return v
		}
	}
	return ""
}

func CoalesceFloat(f ...float64) float64 {
	for _, v := range f {
		if v != 0.00 {
			return v
		}
	}
	return 0.00
}

func CoalesceInt(i ...int) int {
	for _, v := range i {
		if v != 0 {
			return v
		}
	}
	return 0
}

func CoalesceTime(s ...*time.Time) *time.Time {
	for _, v := range s {
		if v != nil && !v.IsZero() {
			return v
		}
	}
	return nil
}
