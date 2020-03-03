package scrabble

import (
	"fmt"
	"math/rand"
)

const shuffleLoop = 20

type Tiles struct {
	remaining []Tile
}

type Tile struct {
	Letter  string
	value   int
	IsBlank bool
}

func (t Tile) String() string {
	if t.value == 10 {
		return fmt.Sprintf("[%s%v]", t.Letter, t.value)
	}
	return fmt.Sprintf("[%s%v]", t.Letter, t.value)
}

// InitializeTiles sets up the tile bag and then shuffles the entries
func InitializeTiles() Tiles {
	tiles := Tiles{}
	tiles.initializeTiles()
	tiles.shuffle()
	return tiles
}

func (t *Tiles) initializeTiles() {
	for letter, count := range MapLetterToCount {
		for i := 0; i < count; i++ {
			t.remaining = append(t.remaining, getTile(letter))
		}
	}
}

func (t *Tiles) shuffle() {
	for rep := 0; rep < shuffleLoop; rep++ {
		rand.Shuffle(len(t.remaining), func(i, j int) {
			t.remaining[i], t.remaining[j] = t.remaining[j], t.remaining[i]
		})
	}
}

// GetTiles returns all of the remaining tiles
func (t *Tiles) GetTiles() []Tile {
	return t.remaining
}

// Draw pulls the requested number of tiles, or the remaining tiles
func (t *Tiles) Draw(numDraw int) []Tile {
	t.shuffle()
	if numDraw > len(t.remaining) {
		numDraw = len(t.remaining)
	}
	tiles := t.remaining[0:numDraw]
	t.remaining = t.remaining[numDraw:]
	return tiles
}

// Return indicates a player is putting tiles back into the bag
func (t *Tiles) Return(tiles []Tile) {
	t.remaining = append(t.remaining, tiles...)
}

func getTile(letter string) Tile {
	return Tile{
		Letter: letter,
		value:  MapLetterToValue[letter],
	}
}

// MapLetterToCount maps the string to number of each tile for the board
var MapLetterToCount = map[string]int{
	"A": 9,
	"B": 2,
	"C": 2,
	"D": 4,
	"E": 12,
	"F": 2,
	"G": 3,
	"H": 2,
	"I": 9,
	"J": 1,
	"K": 1,
	"L": 4,
	"M": 2,
	"N": 6,
	"O": 8,
	"P": 2,
	"Q": 1,
	"R": 6,
	"S": 4,
	"T": 6,
	"U": 4,
	"V": 2,
	"W": 2,
	"X": 1,
	"Y": 2,
	"Z": 1,
	"_": 2,
}

// MapLetterToValue defines the letter to value pairings for the game
var MapLetterToValue = map[string]int{
	"A": 1,
	"B": 3,
	"C": 3,
	"D": 2,
	"E": 1,
	"F": 4,
	"G": 2,
	"H": 4,
	"I": 1,
	"J": 8,
	"K": 5,
	"L": 1,
	"M": 3,
	"N": 1,
	"O": 1,
	"P": 3,
	"Q": 10,
	"R": 1,
	"S": 1,
	"T": 1,
	"U": 1,
	"V": 4,
	"W": 4,
	"X": 8,
	"Y": 4,
	"Z": 10,
	"_": 0,
}

/*
2 blank tiles (scoring 0 points)
1 point: E ×12, A ×9, I ×9, O ×8, N ×6, R ×6, T ×6, L ×4, S ×4, U ×4
2 points: D ×4, G ×3
3 points: B ×2, C ×2, M ×2, P ×2
4 points: F ×2, H ×2, V ×2, W ×2, Y ×2
5 points: K ×1
8 points: J ×1, X ×1
10 points: Q ×1, Z ×1
*/
