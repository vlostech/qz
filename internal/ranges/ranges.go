package ranges

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseRange parses a given string and converts it into RangeQuery model.
//
// Supported patterns:
//
//	".."    - All elements.
//	"5"     - Element by index 5.
//	"..5"   - From 0 to 5 (exclusively).
//	"5.."   - From 5 to the end.
//	"5..10" - From 5 to 10 (exclusively).
//
// ParseRange supports multiple ranges that are separated by ',' (comma). If two ranges overlap each other, they will
// be merged.
//
// Example:
//
//	"..10,5..20,15.." -> ".."
//
//	Explanation:
//	- "..10" and "5..20" has common elements (5, 6, ..., 9) and will be merged into "..20".
//	- "..20" and "15.." also has common elements (15, 16, ..., 19) and will be merged into "..".
//
// An empty string ("") is interpreted as "..".
func ParseRange(rangeStr string) (RangeQuery, error) {
	if rangeStr == "" {
		rangeStr = ".."
	}

	parts := strings.Split(rangeStr, ",")
	rangeParts := make([][2]int, len(parts))

	for i, part := range parts {
		rangePart, err := parseRangePart(part)

		if err != nil {
			return RangeQuery{}, err
		}

		rangeParts[i] = rangePart
	}

	outputRange, err := buildRange(rangeParts)

	if err != nil {
		return RangeQuery{}, err
	}

	return outputRange, nil
}

func parseRangePart(partString string) ([2]int, error) {
	if partString == ".." {
		return [2]int{0, -1}, nil
	}

	if strings.HasPrefix(partString, "..") {
		numberString, _ := strings.CutPrefix(partString, "..")
		closeIndex, err := getValue(numberString)

		if err != nil {
			return [2]int{}, err
		}

		return [2]int{0, closeIndex}, nil
	}

	if strings.HasSuffix(partString, "..") {
		numberString, _ := strings.CutSuffix(partString, "..")
		openIndex, err := getValue(numberString)

		if err != nil {
			return [2]int{}, err
		}

		return [2]int{openIndex, -1}, nil
	}

	if strings.Contains(partString, "..") {
		numberStrings := strings.Split(partString, "..")

		openIndex, err := getValue(numberStrings[0])

		if err != nil {
			return [2]int{}, err
		}

		closeIndex, err := getValue(numberStrings[1])

		if err != nil {
			return [2]int{}, err
		}

		return [2]int{openIndex, closeIndex}, nil
	}

	index, err := getValue(partString)

	if err != nil {
		return [2]int{}, err
	}

	return [2]int{index, index}, nil
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

func buildRange(parts [][2]int) (RangeQuery, error) {
	buf := make([][2]int, len(parts))
	copy(buf, parts)

	shouldRepeat := true

	for shouldRepeat {
		shouldRepeat = false

		for i := 0; i+1 < len(buf); i++ {
			for j := i + 1; j < len(buf); j++ {
				mergeResult := tryMerge(buf[i], buf[j])

				if len(mergeResult) == 1 {
					buf[i] = mergeResult[0]
					buf[j] = buf[len(buf)-1]
					buf = buf[:len(buf)-1]
					shouldRepeat = true
				}
			}
		}
	}

	rangeQuery := RangeQuery{
		Parts: make([]RangeQueryPart, len(buf)),
	}

	for i, r := range buf {
		rangeQuery.Parts[i] = RangeQueryPart{
			OpenIndex:  r[0],
			CloseIndex: r[1],
		}
	}

	return rangeQuery, nil
}

func tryMerge(first, second [2]int) [][2]int {
	if !shouldMerge(first, second) {
		return [][2]int{first, second}
	}

	var minLeft int
	var maxRight int

	if first[0] < second[0] {
		minLeft = first[0]
	} else {
		minLeft = second[0]
	}

	if first[1] == -1 || second[1] == -1 {
		maxRight = -1
	} else if first[1] > second[1] {
		maxRight = first[1]
	} else {
		maxRight = second[1]
	}

	if first[0] == first[1] && second[0] == second[1] {
		return [][2]int{{minLeft, maxRight + 1}}
	} else if first[0] == first[1] && first[0] == maxRight {
		return [][2]int{{minLeft, maxRight + 1}}
	} else if second[0] == second[1] && second[0] == maxRight {
		return [][2]int{{minLeft, maxRight + 1}}
	}

	return [][2]int{{minLeft, maxRight}}
}

func shouldMerge(first, second [2]int) bool {
	if first[0] == first[1] && second[0] == second[1] {
		delta := first[0] - second[0]
		return delta >= -1 && delta <= 1
	}

	if first[1] == -1 || first[1] >= second[0] {
		if second[1] == -1 || first[0] <= second[1] {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}
