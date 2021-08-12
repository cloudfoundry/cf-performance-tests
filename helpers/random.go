package helpers

import (
	"math"
	"math/rand"
)

func Shuffle(items []string) []string {
	rand.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})

	return items
}

func SelectRandom(items []string, count int) []string {
	max := int(math.Min(float64(len(items)), float64(count)))

	items = Shuffle(items)

	return items[:max]
}
