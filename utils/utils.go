package util

import "math"

func rounded(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	ouptut := math.Pow(10, float64(precision))
	return float64(rounded(num * ouptut))
}
