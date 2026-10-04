package main

import (
	"testing"
)

func TestColumnLayout_DivisibleByThree(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"even width", 90, 30},
		{"remainder 1", 80, 30},
		{"remainder 2", 82, 30},
		{"small width", 30, 20},
		{"large width", 200, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := ColumnLayout(tt.width, tt.height)

			// Board width must be divisible by 3
			if l.BoardWidth%3 != 0 {
				t.Errorf("BoardWidth %d not divisible by 3", l.BoardWidth)
			}

			// ColTotalWidth * 3 must be <= BoardWidth
			if l.ColTotalWidth*3 > l.BoardWidth {
				t.Errorf("ColTotalWidth*3 = %d exceeds BoardWidth %d", l.ColTotalWidth*3, l.BoardWidth)
			}

			// ColInternalWidth must be positive
			if l.ColInternalWidth <= 0 {
				t.Errorf("ColInternalWidth = %d, want > 0", l.ColInternalWidth)
			}

			// ColContentWidth must be <= ColInternalWidth
			if l.ColContentWidth > l.ColInternalWidth {
				t.Errorf("ColContentWidth %d > ColInternalWidth %d", l.ColContentWidth, l.ColInternalWidth)
			}

			// ColHeight must be positive and >= minimum
			if l.ColHeight < 10 {
				t.Errorf("ColHeight = %d, want >= 10", l.ColHeight)
			}
		})
	}
}

func TestColumnLayout_MinimumWidth(t *testing.T) {
	l := ColumnLayout(10, 15)
	if l.ColInternalWidth < 10 {
		t.Errorf("ColInternalWidth = %d, want >= 10 (minimum)", l.ColInternalWidth)
	}
	if l.ColHeight < 10 {
		t.Errorf("ColHeight = %d, want >= 10 (minimum)", l.ColHeight)
	}
}

func TestColumnLayout_HeightOverhead(t *testing.T) {
	// Height should account for board margin(1) + title(2) + column
	// padding/border(4) + help margin(1) + help line(1) = 9
	l := ColumnLayout(100, 30)
	expected := 30 - 9
	if l.ColHeight != expected {
		t.Errorf("ColHeight = %d, want %d", l.ColHeight, expected)
	}
}
