package domain

import (
	"testing"
)

func TestCanMerge(t *testing.T) {
	type args struct {
		first  Range
		second Range
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "equal ranges",
			args: args{
				first:  Range{FirstIndex: 1, LastIndex: 2},
				second: Range{FirstIndex: 1, LastIndex: 2},
			},
			want: true,
		},
		{
			name: "neighbour ranges",
			args: args{
				first:  Range{FirstIndex: 1, LastIndex: 2},
				second: Range{FirstIndex: 3, LastIndex: 4},
			},
			want: true,
		},
		{
			name: "ranges with gap",
			args: args{
				first:  Range{FirstIndex: 1, LastIndex: 2},
				second: Range{FirstIndex: 4, LastIndex: 5},
			},
			want: false,
		},
		{
			name: "ranges with overlap",
			args: args{
				first:  Range{FirstIndex: 1, LastIndex: 3},
				second: Range{FirstIndex: 2, LastIndex: 4},
			},
			want: true,
		},
		{
			name: "full range",
			args: args{
				first:  Range{FirstIndex: 0, LastIndex: -1},
				second: Range{FirstIndex: 2, LastIndex: 4},
			},
			want: true,
		},
		{
			name: "two continuous ranges",
			args: args{
				first:  Range{FirstIndex: 5, LastIndex: -1},
				second: Range{FirstIndex: 7, LastIndex: -1},
			},
			want: true,
		},
		{
			name: "continuous range includes other range",
			args: args{
				first:  Range{FirstIndex: 5, LastIndex: 7},
				second: Range{FirstIndex: 2, LastIndex: -1},
			},
			want: true,
		},
		{
			name: "continuous range overlaps other range",
			args: args{
				first:  Range{FirstIndex: 5, LastIndex: 7},
				second: Range{FirstIndex: 7, LastIndex: -1},
			},
			want: true,
		},
		{
			name: "continuous range adjacent to other range",
			args: args{
				first:  Range{FirstIndex: 5, LastIndex: 7},
				second: Range{FirstIndex: 8, LastIndex: -1},
			},
			want: true,
		},
		{
			name: "continuous range has gap with other range",
			args: args{
				first:  Range{FirstIndex: 5, LastIndex: 7},
				second: Range{FirstIndex: 9, LastIndex: -1},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.first.CanMerge(tt.args.second); got != tt.want {
				t.Errorf("CanMerge() = %v, want %v", got, tt.want)
			}

			// Swap parameters to make sure that order of arguments does not
			// affect result.
			if got := tt.args.second.CanMerge(tt.args.first); got != tt.want {
				t.Errorf("CanMerge() = %v, want %v", got, tt.want)
			}
		})
	}
}
