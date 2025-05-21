package utils

import "math/rand"

func RandomIntRange(min, max int) int {
	return min + rand.Intn(max-min+1)
}
