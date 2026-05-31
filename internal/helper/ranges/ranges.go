package ranges

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vlostech/qz/internal/domain"
)

// Parse parses a given string and converts it into domain.RangeGroup.
//
// Supported patterns:
//
//	".."    - All elements.
//	"5"     - Element by index 5.
//	"..5"   - From 0 to 5 (inclusive).
//	"5.."   - From 5 to the end.
//	"5..10" - From 5 to 10 (inclusive).
//
// Parse supports multiple ranges that are separated by ',' (comma). If two
// ranges overlap each other, they will be merged.
//
// The following string:
//
//	"..10, 5..20, 15.." -> ".."
//
// results in full range.
//
// Explanation:
//
//   - "..10" and "5..20" have common elements (5, 6, ..., 10) and will be
//     merged into "..20".
//   - "..20" and "15.." also have common elements (15, 16, ..., 20) and will be
//     merged into "..".
//
// An empty string ("") is interpreted as "..".
func Parse(rangeStr string) (domain.RangeGroup, error) {
	if rangeStr == "" {
		rangeStr = ".."
	}

	strWithoutSpaces := strings.ReplaceAll(rangeStr, " ", "")
	parts := strings.Split(strWithoutSpaces, ",")
	rangeParts := make([]domain.Range, len(parts))

	for i, part := range parts {
		rangePart, err := parseRangePart(part)

		if err != nil {
			return domain.RangeGroup{}, err
		}

		rangeParts[i] = rangePart
	}

	return domain.NewRangeGroup(rangeParts)
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
func parseRangePart(partString string) (domain.Range, error) {
	if partString == ".." {
		return domain.NewRangeFull(), nil
	}

	if strings.HasPrefix(partString, "..") {
		numberString, _ := strings.CutPrefix(partString, "..")
		lastIndex, err := getValue(numberString)

		if err != nil {
			return domain.Range{}, err
		}

		return domain.NewRangeFromStartToIndex(lastIndex)
	}

	if strings.HasSuffix(partString, "..") {
		numberString, _ := strings.CutSuffix(partString, "..")
		firstIndex, err := getValue(numberString)

		if err != nil {
			return domain.Range{}, err
		}

		return domain.NewRangeFromIndexToEnd(firstIndex)
	}

	if strings.Contains(partString, "..") {
		numberStrings := strings.Split(partString, "..")

		firstIndex, err := getValue(numberStrings[0])

		if err != nil {
			return domain.Range{}, err
		}

		lastIndex, err := getValue(numberStrings[1])

		if err != nil {
			return domain.Range{}, err
		}

		return domain.NewRangeByIndices(firstIndex, lastIndex)
	}

	index, err := getValue(partString)

	if err != nil {
		return domain.Range{}, err
	}

	return domain.NewRangeBySingleIndex(index)
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
