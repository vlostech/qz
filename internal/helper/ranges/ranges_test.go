package ranges

import (
	"reflect"
	"testing"

	"github.com/vlostech/qz/internal/domain"
)

func TestParseRange(t *testing.T) {
	tests := []struct {
		name     string
		rangeStr string
		want     domain.RangeGroup
		wantErr  bool
	}{
		{
			name:     "empty string",
			rangeStr: "",
			want: domain.RangeGroup{
				Ranges: []domain.Range{{FirstIndex: 0, LastIndex: -1}},
			},
			wantErr: false,
		},
		{
			name:     "single index",
			rangeStr: "42",
			want: domain.RangeGroup{
				Ranges: []domain.Range{{FirstIndex: 42, LastIndex: 42}},
			},
			wantErr: false,
		},
		{
			name:     "multiple adjacent indexes",
			rangeStr: "5, 6, 7",
			want: domain.RangeGroup{
				Ranges: []domain.Range{{FirstIndex: 5, LastIndex: 7}},
			},
			wantErr: false,
		},
		{
			name:     "multiple indexes with gaps",
			rangeStr: "5, 7, 9",
			want: domain.RangeGroup{
				Ranges: []domain.Range{
					{FirstIndex: 5, LastIndex: 5},
					{FirstIndex: 7, LastIndex: 7},
					{FirstIndex: 9, LastIndex: 9},
				},
			},
			wantErr: false,
		},
		{
			name:     "full range",
			rangeStr: "..",
			want: domain.RangeGroup{
				Ranges: []domain.Range{{FirstIndex: 0, LastIndex: -1}},
			},
			wantErr: false,
		},
		{
			name:     "range without last index",
			rangeStr: "42..",
			want: domain.RangeGroup{
				Ranges: []domain.Range{{FirstIndex: 42, LastIndex: -1}},
			},
			wantErr: false,
		},
		{
			name:     "range without first index",
			rangeStr: "..42",
			want: domain.RangeGroup{
				Ranges: []domain.Range{{FirstIndex: 0, LastIndex: 42}},
			},
			wantErr: false,
		},
		{
			name:     "range with both indexes",
			rangeStr: "3..6",
			want: domain.RangeGroup{
				Ranges: []domain.Range{{FirstIndex: 3, LastIndex: 6}},
			},
			wantErr: false,
		},
		{
			name:     "multiple ranges",
			rangeStr: "..10, 15, 20, 30..40, 50..",
			want: domain.RangeGroup{
				Ranges: []domain.Range{
					{FirstIndex: 0, LastIndex: 10},
					{FirstIndex: 15, LastIndex: 15},
					{FirstIndex: 20, LastIndex: 20},
					{FirstIndex: 30, LastIndex: 40},
					{FirstIndex: 50, LastIndex: -1},
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
