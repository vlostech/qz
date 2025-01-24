package random

import "math/rand"

// Randomize returns count numbers from nums in random order.
func Randomize(nums []int, count int) []int {
	lastIndex := len(nums) - 1
	leftIndexes := nums

	var actualCount int

	if len(nums) < count {
		actualCount = len(nums)
	} else {
		actualCount = count
	}

	output := make([]int, actualCount)

	for i := 0; i < len(output); i++ {
		if lastIndex == 0 {
			output[i] = leftIndexes[0]
			break
		}

		randomIndex := rand.Intn(lastIndex + 1)
		randomValue := leftIndexes[randomIndex]
		output[i] = randomValue

		leftIndexes[randomIndex] = leftIndexes[lastIndex]
		lastIndex--
	}

	return output
}
