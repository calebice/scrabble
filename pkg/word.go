package scrabble

import "fmt"

// Word represents sequence of squares that makes up a scrabble word
type Word struct {
	direction string
	Squares   []Square
}

// String parses collection of squares into a dictionary formatted word
func (w Word) String() string {
	var word string
	for _, s := range w.Squares {
		word = fmt.Sprintf("%s%s", word, s.Value.Letter)
	}
	return word
}

// ScoreWord calculates the value of the word
func (w Word) ScoreWord() int {
	var total int
	wordMultiplier := 1
	for _, s := range w.Squares {
		if s.Used {
			total += s.Value.Value
		} else if mult, ok := letterMult[s.Multiplier]; ok {
			total += mult * s.Value.Value
		} else {
			total += s.Value.Value
			wordMultiplier *= wordMult[s.Multiplier]
		}
	}
	total = total * wordMultiplier
	return total
}

// Sorting implementation to allow organizing all the letters correctly in a word
func (w Word) Len() int      { return len(w.Squares) }
func (w Word) Swap(i, j int) { w.Squares[i], w.Squares[j] = w.Squares[j], w.Squares[i] }
func (w Word) Less(i, j int) bool {
	if w.direction == "vertical" {
		return w.Squares[i].Coordinate.x < w.Squares[j].Coordinate.x
	}
	return w.Squares[i].Coordinate.y < w.Squares[j].Coordinate.y
}
