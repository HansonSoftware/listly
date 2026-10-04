package main

// Layout holds computed column dimensions
type Layout struct {
	BoardWidth       int // Total width available for all 3 columns
	ColTotalWidth    int // Total width per column (including padding+border)
	ColInternalWidth int // Width for list internal rendering
	ColContentWidth  int // Width for MaxWidth constraint (title + list)
	ColHeight        int // Height for list content
}

// ColumnLayout computes column dimensions from terminal size.
// All padding/overhead is accounted for:
// - ColumnStyle: Padding(1,2) + Border = 6 chars horizontal, 4 vertical
// - Title style: Padding(0,1) = 2 chars horizontal
// - Gap between columns: 2 chars
// - Help bar + top margin: ~4 lines
func ColumnLayout(width, height int) Layout {
	// Board takes (width - remainder) to be divisible by 3
	boardWidth := width - (width % 3)
	// Each column gets an equal third, flush with its neighbors.
	colTotalWidth := boardWidth / 3
	// Internal width for list: total - column padding(2+2) - border(1+1)
	colInternalWidth := colTotalWidth - 6
	if colInternalWidth < 10 {
		colInternalWidth = 10
	}
	// Content width: further subtract title padding
	colContentWidth := colInternalWidth - 2
	if colContentWidth < 10 {
		colContentWidth = 10
	}
	// Budget: session title(1) + board top margin(1) + column title(2)
	// + column border/padding(4) + footer(1) + footer margin(1).
	colHeight := height - 10
	if colHeight < 10 {
		colHeight = 10
	}
	return Layout{
		BoardWidth:       boardWidth,
		ColTotalWidth:    colTotalWidth,
		ColInternalWidth: colInternalWidth,
		ColContentWidth:  colContentWidth,
		ColHeight:        colHeight,
	}
}
