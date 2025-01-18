package ranges

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

func ParseRange(rangeStr string, count int) ([]int, error) {
	if count < 0 {
		return nil, fmt.Errorf("count cannot be less than 0")
	}

	if rangeStr == "" {
		rangeStr = ".."
	}

	parts := strings.Split(rangeStr, ",")

	var questionIndexes []int

	for _, part := range parts {
		partIndexes, err := parseRangePart(part, count)

		if err != nil {
			return nil, err
		}

		questionIndexes = append(questionIndexes, partIndexes...)
	}

	slices.Sort(questionIndexes)
	output := slices.Compact(questionIndexes)

	return output, nil
}

func parseRangePart(partString string, count int) ([]int, error) {
	if partString == ".." {
		indexes := make([]int, count)

		for i := range indexes {
			indexes[i] = i
		}

		return indexes, nil
	}

	if strings.HasPrefix(partString, "..") {
		numberString, _ := strings.CutPrefix(partString, "..")
		closeIndex, err := getValue(numberString)

		if err != nil {
			return nil, err
		}

		if closeIndex > count {
			closeIndex = count
		}

		indexes := make([]int, closeIndex)

		for i := range indexes {
			indexes[i] = i
		}

		return indexes, nil
	}

	if strings.HasSuffix(partString, "..") {
		numberString, _ := strings.CutSuffix(partString, "..")
		openIndex, err := getValue(numberString)

		if err != nil {
			return nil, err
		}

		if openIndex >= count {
			return []int{}, nil
		}

		updatedCount := count - openIndex

		indexes := make([]int, updatedCount)

		for i := range updatedCount {
			indexes[i] = openIndex + i
		}

		return indexes, nil
	}

	if strings.Contains(partString, "..") {
		numberStrings := strings.Split(partString, "..")

		openIndex, err := getValue(numberStrings[0])

		if err != nil {
			return nil, err
		}

		closeIndex, err := getValue(numberStrings[1])

		if err != nil {
			return nil, err
		}

		if openIndex >= closeIndex {
			return []int{}, nil
		}

		if openIndex >= count {
			return []int{}, nil
		}

		if closeIndex > count {
			closeIndex = count
		}

		questionCount := closeIndex - openIndex

		indexes := make([]int, questionCount)

		for i := range questionCount {
			indexes[i] = openIndex + i
		}

		return indexes, nil
	}

	index, err := getValue(partString)

	if err != nil {
		return nil, err
	}

	if index >= count {
		return []int{}, nil
	}

	return []int{index}, nil
}

func getValue(str string) (int, error) {
	val, err := strconv.Atoi(str)

	if err != nil {
		return 0, err
	}

	if val < 0 {
		return 0, fmt.Errorf("value cannot be less than 0")
	}

	return val, nil
}
