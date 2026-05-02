package ranges

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vlostech/qz/internal/domain"
)

// Parse parses a given string and converts it into domain.RangeQuery model.
//
// Supported patterns:
//
//	".."    - All elements.
//	"5"     - Element by index 5.
//	"..5"   - From 0 to 5 (exclusively).
//	"5.."   - From 5 to the end.
//	"5..10" - From 5 to 10 (exclusively).
//
// Parse supports multiple ranges that are separated by ',' (comma). If two
// ranges overlap each other, they will be merged.
//
// Example:
//
//	"..10, 5..20, 15.." -> ".."
//
// Explanation for the example above:
//
//   - "..10" and "5..20" have common elements (5, 6, ..., 9) and will be merged
//     into "..20".
//   - "..20" and "15.." also have common elements (15, 16, ..., 19) and will be
//     merged into "..".
//
// An empty string ("") is interpreted as "..".
func Parse(rangeStr string) (domain.RangeQuery, error) {
	if rangeStr == "" {
		rangeStr = ".."
	}

	strWithoutSpaces := strings.ReplaceAll(rangeStr, " ", "")
	parts := strings.Split(strWithoutSpaces, ",")
	rangeParts := make([][2]int, len(parts))

	for i, part := range parts {
		rangePart, err := parseRangePart(part)

		if err != nil {
			return domain.RangeQuery{}, err
		}

		rangeParts[i] = rangePart
	}

	outputRange, err := buildRange(rangeParts)

	if err != nil {
		return domain.RangeQuery{}, err
	}

	return outputRange, nil
}

// parseRangePart parses a single range part represented by a string and turns
// it into an array with open and close indices:
//
// Examples:
//
//	".."    -> [0, -1]
//	"5"     -> [5, 5]
//	"..5"   -> [0, 5]
//	"5.."   -> [5, -1]
//	"5..10" -> [5, 10]
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

		return [2]int{0, closeIndex + 1}, nil
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

		return [2]int{openIndex, closeIndex + 1}, nil
	}

	index, err := getValue(partString)

	if err != nil {
		return [2]int{}, err
	}

	return [2]int{index, index + 1}, nil
}

// getValue extracts an integer value from str and validates it.
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

// buildRange constructs a [domain.RangeQuery] from a slice of integer ranges.
func buildRange(parts [][2]int) (domain.RangeQuery, error) {
	buf := make([][2]int, len(parts))
	copy(buf, parts)

	shouldRepeat := true

	for shouldRepeat {
		shouldRepeat = false

		for i := 0; i+1 < len(buf); i++ {
			for j := i + 1; j < len(buf); j++ {
				isMerged, mergedRange := tryMerge(buf[i], buf[j])

				if isMerged {
					buf[i] = mergedRange
					buf[j] = buf[len(buf)-1]
					buf = buf[:len(buf)-1]
					shouldRepeat = true
				}
			}
		}
	}

	rangeQuery := domain.RangeQuery{
		Parts: make([]domain.RangeQueryPart, len(buf)),
	}

	for i, r := range buf {
		rangeQuery.Parts[i] = domain.RangeQueryPart{
			OpenIndex:  r[0],
			CloseIndex: r[1],
		}
	}

	return rangeQuery, nil
}

// tryMerge attempts to merge two ranges if they are adjacent or overlapping.
// For example, [1, 2] and [2, 3] will be merged into [1, 3].
func tryMerge(first, second [2]int) (bool, [2]int) {
	if !shouldMerge(first, second) {
		return false, [2]int{}
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
		return true, [2]int{minLeft, maxRight + 1}
	} else if first[0] == first[1] && first[0] == maxRight {
		return true, [2]int{minLeft, maxRight + 1}
	} else if second[0] == second[1] && second[0] == maxRight {
		return true, [2]int{minLeft, maxRight + 1}
	}

	return true, [2]int{minLeft, maxRight}
}

// shouldMerge checks if two ranges can be merged into a single range. For
// example, [1, 2] and [2, 3] can be merged into [1, 3].
func shouldMerge(first, second [2]int) bool {
	if first[0] == first[1] && second[0] == second[1] {
		delta := first[0] - second[0]
		return delta >= -1 && delta <= 1
	}

	if first[1] == -1 || first[1] >= second[0] {
		if second[1] == -1 || first[0] <= second[1] {
			return true
		}

		return false
	}

	return false
}
