package scrabble

import (
	"fmt"
)

// Board represents the view of the scrabble board
type Board [Size][Size]Square

var wordMult = map[string]int{
	"DW": 2,
	"TW": 3,
}
var letterMult = map[string]int{
	"__": 1,
	"DL": 2,
	"TL": 3,
}

var board = [][]string{
	[]string{"TW", "__", "__", "DL", "__", "__", "__", "TW", "__", "__", "__", "DL", "__", "__", "TW"},
	[]string{"__", "DW", "__", "__", "__", "TL", "__", "__", "__", "TL", "__", "__", "__", "DW", "__"},
	[]string{"__", "__", "DW", "__", "__", "__", "DL", "__", "DL", "__", "__", "__", "DW", "__", "__"},
	[]string{"DL", "__", "__", "DW", "__", "__", "__", "DL", "__", "__", "__", "DW", "__", "__", "DL"},
	[]string{"__", "__", "__", "__", "DW", "__", "__", "__", "__", "__", "DW", "__", "__", "__", "__"},
	[]string{"__", "TL", "__", "__", "__", "TL", "__", "__", "__", "TL", "__", "__", "__", "TL", "__"},
	[]string{"__", "__", "DL", "__", "__", "__", "DL", "__", "DL", "__", "__", "__", "DL", "__", "__"},
	[]string{"TW", "__", "__", "DL", "__", "__", "__", "DW", "__", "__", "__", "DL", "__", "__", "TW"},
	[]string{"__", "__", "DL", "__", "__", "__", "DL", "__", "DL", "__", "__", "__", "DL", "__", "__"},
	[]string{"__", "TL", "__", "__", "__", "TL", "__", "__", "__", "TL", "__", "__", "__", "TL", "__"},
	[]string{"__", "__", "__", "__", "DW", "__", "__", "__", "__", "__", "DW", "__", "__", "__", "__"},
	[]string{"DL", "__", "__", "DW", "__", "__", "__", "DL", "__", "__", "__", "DW", "__", "__", "DL"},
	[]string{"__", "__", "DW", "__", "__", "__", "DL", "__", "DL", "__", "__", "__", "DW", "__", "__"},
	[]string{"__", "DW", "__", "__", "__", "TL", "__", "__", "__", "TL", "__", "__", "__", "DW", "__"},
	[]string{"TW", "__", "__", "DL", "__", "__", "__", "TW", "__", "__", "__", "DL", "__", "__", "TW"},
}

// NewBoard returns an instantiated board with the proper multipliers
func NewBoard() Board {
	var Board Board
	for i, row := range board {
		for j, val := range row {
			Board[i][j] = Square{
				Multiplier: val,
				Coordinate: Coordinate{i, j},
			}
		}
	}

	return Board
}

func (b Board) String() string {
	var str string
	str = fmt.Sprint("   ")
	for i := 1; i <= Size; i++ {
		if i > 9 {
			str = fmt.Sprintf("%s|%2v| ", str, i)
		} else {
			str = fmt.Sprintf("%s|_%v| ", str, i)
		}
	}
	str = fmt.Sprintf("%s\n", str)
	for i, a := range b {
		str = fmt.Sprintf("%s%s |", str, string(toRune(i+1)))
		for _, y := range a {
			str = fmt.Sprintf("%s%s ", str, y)
		}
		str = fmt.Sprintf("%s\n", str)
	}
	return str
}

// Square represents an individual unit on the board
// @Value occupying Tile
// @Multiplier multiplier to apply to letter or word
// @Used whether multiplier has been applied to a score
// @Coordinate represents location on the board square occupies
type Square struct {
	Value      Tile
	Multiplier string
	Used       bool
	Coordinate Coordinate
}

func (s Square) String() string {
	if s.IsEmpty() {
		return fmt.Sprintf("[%s]", s.Multiplier)
	}
	return fmt.Sprintf("%s", s.Value)
}

// IsEmpty checks if square is filled with a letter
func (s Square) IsEmpty() bool {
	return s.Value == Tile{}
}

// SetSquareUsed indicates that multipliers have been applied to provided coordinate
// prevents duplicate multiplier applications
func (b *Board) SetSquareUsed(coordinate Coordinate) {
	b[coordinate.x][coordinate.y].Used = true
}

/* Board Structure
[TW] [__] [__] [DL] [__] [__] [__] [TW] [__] [__] [__] [DL] [__] [__] [TW]
[__] [DW] [__] [__] [__] [TL] [__] [__] [__] [TL] [__] [__] [__] [DW] [__]
[__] [__] [DW] [__] [__] [__] [DL] [__] [DL] [__] [__] [__] [DW] [__] [__]
[DL] [__] [__] [DW] [__] [__] [__] [DL] [__] [__] [__] [DW] [__] [__] [DL]
[__] [__] [__] [__] [DW] [__] [__] [__] [__] [__] [DW] [__] [__] [__] [__]
[__] [TL] [__] [__] [__] [TL] [__] [__] [__] [TL] [__] [__] [__] [TL] [__]
[__] [__] [DL] [__] [__] [__] [DL] [__] [DL] [__] [__] [__] [DL] [__] [__]
[TW] [__] [__] [DL] [__] [__] [__] [DW] [__] [__] [__] [DL] [__] [__] [TW]
[__] [__] [DL] [__] [__] [__] [DL] [__] [DL] [__] [__] [__] [DL] [__] [__]
[__] [TL] [__] [__] [__] [TL] [__] [__] [__] [TL] [__] [__] [__] [TL] [__]
[__] [__] [__] [__] [DW] [__] [__] [__] [__] [__] [DW] [__] [__] [__] [__]
[DL] [__] [__] [DW] [__] [__] [__] [DL] [__] [__] [__] [DW] [__] [__] [DL]
[__] [__] [DW] [__] [__] [__] [DL] [__] [DL] [__] [__] [__] [DW] [__] [__]
[__] [DW] [__] [__] [__] [TL] [__] [__] [__] [TL] [__] [__] [__] [DW] [__]
[TW] [__] [__] [DL] [__] [__] [__] [TW] [__] [__] [__] [DL] [__] [__] [TW]
*/
