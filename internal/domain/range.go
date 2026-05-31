package domain

import "fmt"

var InvalidRangeErr = fmt.Errorf("invalid arguments for range")

// RangeGroup represents a group of ranges.
type RangeGroup struct {
	// Ranges is a list of ranges.
	Ranges []Range
}

// Range represents a range of indices.
type Range struct {
	// FirstIndex is the index of the first element in the range.
	FirstIndex int
	// LastIndex is the index of the last element in the range. Can be -1 if the
	// range is continuous.
	LastIndex int
}

// NewRangeGroup constructs a domain.RangeGroup from a slice of domain.Range.
// Function merges provided ranges if they are adjacent or overlapping.
func NewRangeGroup(ranges []Range) (RangeGroup, error) {
	buf := make([]Range, len(ranges))
	copy(buf, ranges)

	for i := 0; i+1 < len(buf); i++ {
		for j := i + 1; j < len(buf); j++ {
			first := buf[i]
			second := buf[j]

			mergedRange, isMerged := first.TryMerge(second)

			if isMerged {
				buf[i] = mergedRange
				buf[j] = buf[len(buf)-1]
				buf = buf[:len(buf)-1]
				j--
			}
		}
	}

	resultRanges := make([]Range, len(buf))
	copy(resultRanges, buf)

	return RangeGroup{Ranges: resultRanges}, nil
}

// NewRangeFull creates a full range where the first index is 0 and the last
// index is -1.
func NewRangeFull() Range {
	return Range{
		FirstIndex: 0,
		LastIndex:  -1,
	}
}

// NewRangeBySingleIndex creates a range where the first and last index are the
// same.
func NewRangeBySingleIndex(index int) (Range, error) {
	if index < 0 {
		return Range{}, InvalidRangeErr
	}

	return Range{FirstIndex: index, LastIndex: index}, nil
}

// NewRangeByIndices creates a range that contains all indices between the first
// and last index, inclusive. The last index cannot be -1. To create a
// continuous range, use NewRangeFromIndexToEnd.
func NewRangeByIndices(firstIndex, lastIndex int) (Range, error) {
	if firstIndex < 0 || lastIndex < 0 {
		return Range{}, InvalidRangeErr
	}

	if lastIndex < firstIndex {
		return Range{}, InvalidRangeErr
	}

	return Range{FirstIndex: firstIndex, LastIndex: lastIndex}, nil
}

// NewRangeFromStartToIndex creates a range from the start of the list to the
// given index. The last index cannot be -1. To create a continuous range, use
// NewRangeFromIndexToEnd.
func NewRangeFromStartToIndex(lastIndex int) (Range, error) {
	if lastIndex < 0 {
		return Range{}, InvalidRangeErr
	}

	return Range{FirstIndex: 0, LastIndex: lastIndex}, nil
}

// NewRangeFromIndexToEnd creates a range from the first index to the end of the
// list.
func NewRangeFromIndexToEnd(firstIndex int) (Range, error) {
	if firstIndex < 0 {
		return Range{}, InvalidRangeErr
	}

	return Range{FirstIndex: firstIndex, LastIndex: -1}, nil
}

// TryMerge attempts to merge two ranges into a single range. For example,
// [1, 2] and [3, 4] will be merged into [1, 4].
func (r Range) TryMerge(other Range) (Range, bool) {
	if !r.CanMerge(other) {
		return Range{}, false
	}

	if r.IsFullRange() || other.IsFullRange() {
		return NewRangeFull(), true
	}

	var minLeft int
	var maxRight int

	if r.FirstIndex < other.FirstIndex {
		minLeft = r.FirstIndex
	} else {
		minLeft = other.FirstIndex
	}

	if r.IsContinuousRange() || other.IsContinuousRange() {
		maxRight = -1
	} else if r.LastIndex > other.LastIndex {
		maxRight = r.LastIndex
	} else {
		maxRight = other.LastIndex
	}

	mergedRange := Range{
		FirstIndex: minLeft,
		LastIndex:  maxRight,
	}

	return mergedRange, true
}

// CanMerge checks if two ranges can be merged into a single range. For example,
// [1, 2] and [3, 4] can be merged into [1, 4].
func (r Range) CanMerge(other Range) bool {
	// If any of the ranges is full, they can be merged.
	//
	// Examples:
	// [0, -1], [5, 7] -> true

	if r.IsFullRange() || other.IsFullRange() {
		return true
	}

	// If both of the ranges are continuous, they can be merged.
	//
	// Examples:
	// [5, -1], [7, -1] -> true

	if r.IsContinuousRange() && other.IsContinuousRange() {
		return true
	}

	// If one of the ranges is continuous and includes any index of the other
	// range, or they are adjacent, they can be merged.
	//
	// Examples:
	// [5, 7], [7, -1] -> true (the first range contains indexes from the second)
	// [5, 7], [8, -1] -> true (the second range is adjacent to the first)
	// [5, 7], [9, -1] -> false (there is a gap between the ranges)

	if r.IsContinuousRange() {
		return other.LastIndex+1 >= r.FirstIndex
	}

	if other.IsContinuousRange() {
		return r.LastIndex+1 >= other.FirstIndex
	}

	// If one of the ranges includes any index of the other range, or they are
	// adjacent, they can be merged.
	//
	// Examples:
	// [1, 3], [2, 6] -> true (the first range contains indexes from the second)
	// [1, 6], [2, 3] -> true (the second range is subset of the first)
	// [1, 3], [4, 6] -> true (the second range is adjacent to the first)
	// [1, 3], [5, 6] -> false (there is a gap between the ranges)
	if r.LastIndex+1 >= other.FirstIndex && other.LastIndex+1 >= r.FirstIndex {
		return true
	}

	return false
}

// IsFullRange returns true if the range is a full range.
func (r Range) IsFullRange() bool {
	return r.FirstIndex == 0 && r.LastIndex == -1
}

// IsContinuousRange returns true if the range is continuous.
func (r Range) IsContinuousRange() bool {
	return r.LastIndex == -1
}
