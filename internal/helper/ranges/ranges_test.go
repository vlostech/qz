package ranges

import (
	"reflect"
	"testing"

	"github.com/vlostech/qz/internal/domain"
)

func Test_ParseRange(t *testing.T) {
	tests := []struct {
		name     string
		rangeStr string
		want     domain.RangeQuery
		wantErr  bool
	}{
		{
			name:     "ReturnsFullRangeWhenEmptyStringProvided",
			rangeStr: "",
			want: domain.RangeQuery{
				Parts: []domain.RangeQueryPart{{OpenIndex: 0, CloseIndex: -1}},
			},
			wantErr: false,
		},
		{
			name:     "ReturnsSingleRangeWhenSingleIndexProvided",
			rangeStr: "42",
			want: domain.RangeQuery{
				Parts: []domain.RangeQueryPart{{OpenIndex: 42, CloseIndex: 43}},
			},
			wantErr: false,
		},
		{
			name:     "ReturnsSingleRangeWhenConsecutiveNumbersProvided",
			rangeStr: "5, 6, 7",
			want: domain.RangeQuery{
				Parts: []domain.RangeQueryPart{{OpenIndex: 5, CloseIndex: 8}},
			},
			wantErr: false,
		},
		{
			name:     "ReturnsTwoRangesWhenGapBetweenNumbers",
			rangeStr: "5, 6, 8",
			want: domain.RangeQuery{
				Parts: []domain.RangeQueryPart{
					{OpenIndex: 5, CloseIndex: 7},
					{OpenIndex: 8, CloseIndex: 9},
				},
			},
			wantErr: false,
		},
		{
			name:     "ReturnsFullRangeWhenTwoDotsProvided",
			rangeStr: "..",
			want: domain.RangeQuery{
				Parts: []domain.RangeQueryPart{{OpenIndex: 0, CloseIndex: -1}},
			},
			wantErr: false,
		},
		{
			name:     "ReturnsRangeWhenHalfOpenRangeFromIndexToEnd",
			rangeStr: "42..",
			want: domain.RangeQuery{
				Parts: []domain.RangeQueryPart{{OpenIndex: 42, CloseIndex: -1}},
			},
			wantErr: false,
		},
		{
			name:     "ReturnsRangeWhenHalfOpenRangeFromBeginningToIndex",
			rangeStr: "..42",
			want: domain.RangeQuery{
				Parts: []domain.RangeQueryPart{{OpenIndex: 0, CloseIndex: 43}},
			},
			wantErr: false,
		},
		{
			name:     "ReturnsRangeWhenSimpleRangeProvided",
			rangeStr: "3..6",
			want: domain.RangeQuery{
				Parts: []domain.RangeQueryPart{{OpenIndex: 3, CloseIndex: 7}},
			},
			wantErr: false,
		},
		{
			name:     "ReturnsMultipleRangesWhenComplexStringProvided",
			rangeStr: "..10, 15, 20, 30..40, 50..",
			want: domain.RangeQuery{
				Parts: []domain.RangeQueryPart{
					{OpenIndex: 0, CloseIndex: 11},
					{OpenIndex: 15, CloseIndex: 16},
					{OpenIndex: 20, CloseIndex: 21},
					{OpenIndex: 30, CloseIndex: 41},
					{OpenIndex: 50, CloseIndex: -1},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.rangeStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRange() got = %v, want %v", got, tt.want)
			}
		})
	}
}
