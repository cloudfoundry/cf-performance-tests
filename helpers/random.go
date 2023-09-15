package helpers

import (
	"math/rand"
)

func Shuffle(items []string) []string {
	rand.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})

	return items
}
