package scrabble

import (
	"fmt"
	"strconv"
	"strings"
)

func parseTiles(tokens []string) []Tile {
	var tiles []Tile
	for _, t := range tokens {
		letter := strings.ToUpper(t)
		tiles = append(tiles, getTile(letter))
	}

	return tiles
}

// parses input tokens into a list of tiles to be placed
// received in format of `a(1,a)`
// special case of _a(1,a) indicates usage of blank tile
func parseTilePlacements(tokens []string) ([]TilePlacement, error) {
	var tilePlacements []TilePlacement
	for _, t := range tokens {
		t = strings.Trim(t, ")")
		spl := strings.Split(t, "(")
		if len(spl) != 2 {
			return nil, ErrTileFormat
		}
		letter := strings.ToUpper(spl[0])
		coord := strings.Split(spl[1], ",")
		if len(spl) != 2 {
			return nil, ErrTileFormat
		}
		rawX := coord[0]
		rawY := coord[1]

		x, err := strconv.Atoi(rawX)
		if err != nil {
			return nil, err
		}

		runeY := []rune(rawY)
		y := toInt(runeY[0])
		if y > 15 || y < 1 {
			return nil, ErrInvalidIndex
		}

		if strings.HasPrefix(letter, "_") {
			tilePlacements = append(tilePlacements, TilePlacement{
				Location: Coordinate{x, y},
				Tile:     Tile{Letter: letter, value: 0, IsBlank: true},
			})
		} else {
			tilePlacements = append(tilePlacements, TilePlacement{
				Location: Coordinate{x - 1, y - 1},
				Tile:     getTile(letter),
			})
		}
	}

	return tilePlacements, nil
}

func flipDirection(dir string) string {
	if dir == "vertical" {
		return "horizontal"
	}
	return "vertical"
}

func toInt(char rune) int {
	return int(char - 'a' + 1)
}

func toRune(i int) rune {
	return rune('a' - 1 + i)
}

// Errors in handling parsing an error in user input
var (
	ErrTileFormat   = fmt.Errorf("Could not parse tile input")
	ErrInvalidIndex = fmt.Errorf("Attempting to address invalid index")
)
