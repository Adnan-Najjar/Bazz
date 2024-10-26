package bot

import (
	"math"
)

func Max(slice []float64) float64 {
	if len(slice) == 0 {
		return math.NaN()
	}

	maxValue := slice[0]
	for _, value := range slice {
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}

func Min(slice []float64) float64 {
	if len(slice) == 0 {
		return math.NaN()
	}

	maxValue := slice[0]
	for _, value := range slice {
		if value < maxValue {
			maxValue = value
		}
	}
	return maxValue
}
