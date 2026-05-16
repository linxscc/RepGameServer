package service

import "math"

func DollarsToCents(d float64) int64 {
	return int64(math.Round(d * 100))
}

func CentsToDollars(c int64) float64 {
	return float64(c) / 100
}
