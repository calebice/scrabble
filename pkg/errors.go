package scrabble

import "fmt"

// ErrNoValidWordsFound represents when a user places a single tile but it forms no words
var ErrNoValidWordsFound = fmt.Errorf("no valid words found in tile placements")

// ErrTileNotInHand represents an attempt at using a tile that a player does not have
var ErrTileNotInHand = fmt.Errorf("Tile requested for action but not found")

// ErrNotEnoughTilesForSwap when a player requests a swap that is greater than total tiles remaining
var ErrNotEnoughTilesForSwap = fmt.Errorf("Could not perform swap, not enough tiles remaining")

// ErrWordDisconnected represents a word placement that is not connected
var ErrWordDisconnected = fmt.Errorf("Word placement invalid, gap between letters found")

// ErrNoConnectionPoint represents a word placement that is not against existing tiles
var ErrNoConnectionPoint = fmt.Errorf("Word placement invalid, placed word not connected to existing tiles")

// ErrInvalidPlacement represents an failed attempt to place a tile non-linearly
var ErrInvalidPlacement = fmt.Errorf("Word placement invalid, must place only horizontal or vertically")

// ErrRepeatedPlacement represents a placement where two tiles are attempted to be played in same spot
var ErrRepeatedPlacement = fmt.Errorf("Word placement invalid, played tiles in the same position")

// ErrInvalidSpace indicates an invalid tile placement
var ErrInvalidSpace = fmt.Errorf("Provided space is illegal. Must be in range of [a,%s], [0,%v]", string(toRune(Size-1)), Size-1)

// ErrInvalidStart starting turn requires tile be placed in center of board
var ErrInvalidStart = fmt.Errorf("Starting move must touch center tile")

// ErrInvalidAction is when a user attempts to perform an illegal operation
var ErrInvalidAction = fmt.Errorf("Invalid action requested: allowed [swap, place]")

// Errors related to database interactions
var (
	ErrCouldNotUpdatePlayerState = fmt.Errorf("player state update called, could not update")
	ErrCouldNotUpdateGame        = fmt.Errorf("game update called, could not update")
	ErrInsertPlayerState         = fmt.Errorf("could not insert player state")
	ErrInsertTurnFailed          = fmt.Errorf("could not insert turn")
)

// ErrSpaceOccupied represents an error for an already occupied coordinate on the board
type ErrSpaceOccupied struct {
	Location Coordinate
}

func (e ErrSpaceOccupied) Error() string {
	return fmt.Sprintf("Could not place tile: Space [%v, %v] already occupied", e.Location.x, e.Location.y)
}

// ErrInvalidWords represents a set of words that are created by the move that are invalid
type ErrInvalidWords struct {
	failedWords []string
}

func (e ErrInvalidWords) Error() string {
	return fmt.Sprintf("Invalid words: %v", e.failedWords)
}
