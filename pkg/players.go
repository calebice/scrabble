package scrabble

type PlayerRequest struct {
	Name         string
	UsePlainText bool
}

// Player represents an active participant
type Player struct {
	id           int64
	pStateID     int64
	tiles        []Tile
	score        int
	Name         string
	nextID       int64
	highestScore int
	highestWord  string
	UsePlainText bool
	//TODO add metadata
}

// Update adds a players score from the round to their total
// Also updates the players tiles using placed tiles
func (p *Player) Update(addScore int, place []TilePlacement) {
	p.score += addScore
	for _, pl := range place {
		for i, t := range p.tiles {
			if pl.Tile == t {
				// remove found tile from hand
				if i == len(p.tiles) {
					p.tiles = p.tiles[:i]
				} else {
					p.tiles = append(p.tiles[:i], p.tiles[i+1:]...)
				}
			}
		}
	}
}

// Tiles returns a users tiles for accessing
func (p Player) Tiles() []Tile {
	return p.tiles
}

// Score returns a players score
func (p Player) Score() int {
	return p.score
}

// HighestWord returns the maximum scored word a player has played
func (p Player) HighestWord() string {
	return p.highestWord
}

// HighestScore returns the highest score a player has hit in one turn
func (p Player) HighestScore() int {
	return p.highestScore
}
