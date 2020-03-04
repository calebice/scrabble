package scrabble

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
)

// Constants of the game
const (
	HandSize = 7
	Size     = 15
	dictPath = "data/dictionary.txt"
	BINGO    = 50
)

var (
	// CENTER represents middle of board
	CENTER = Coordinate{7, 7}
)

// Game represents an active state of a game of scrabble
type Game struct {
	id         int64
	board      Board
	players    []Player
	Tiles      Tiles
	Dictionary Dictionary
	Turn       Turn
	Turns      []Turn
}

// Turn represents a unit of action driving the game
type Turn struct {
	number int
	input  string
	player Player
}

// GetBoard returns the game board
func (game Game) GetBoard() Board {
	return game.board
}

// GetPlayers returns the list of players
func (game Game) GetPlayers() []Player {
	return game.players
}

// SetPlayerState takes a changed player condition and updates
// Search using name of player
// TODO consider either mapping names --> players
// OR mapping ids to players and have players be
// map[int]Player as struct in Game
func (game *Game) SetPlayerState(player Player) {
	for i, p := range game.players {
		if player.Name == p.Name {
			game.players[i] = player
			break
		}
	}
}

// SetBoard updates the game board to match current board
func (game *Game) SetBoard(board Board) {
	game.board = board
}

// Draw represents the action of pulling tiles out of the bag
func (game *Game) Draw(num int) []Tile {
	return game.Tiles.Draw(num)
}

// NewGame begins a new game of scrabble
// Instantiates the tiles
func NewGame(names []string) *Game {
	tiles := InitializeTiles()
	board := NewBoard()
	game := Game{
		board:   board,
		players: []Player{},
		Tiles:   tiles,
	}

	game.AddPlayers(names)
	dict, err := LoadDictionary(dictPath)
	if err != nil {
		panic(err)
	}
	game.Dictionary = dict
	game.Turn = Turn{
		player: game.players[0],
		number: 1,
	}

	return &game
}

// LoadFromState loads a pre-existing game
func LoadFromState(board Board, tiles Tiles, players []Player, turn Turn) Game {

	dict, err := LoadDictionary(dictPath)
	if err != nil {
		panic(err)
	}

	return Game{
		board:      board,
		players:    players,
		Tiles:      tiles,
		Turn:       turn,
		Dictionary: dict,
	}
}

// Player represents an active participant
type Player struct {
	id           int64
	tiles        []Tile
	score        int
	Name         string
	next         *Player
	highestScore int
	highestWord  string
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

// AddPlayers instantiates players into the game
func (game *Game) AddPlayers(names []string) {
	for _, name := range names {
		game.players = append(game.players, Player{
			Name:  name,
			tiles: game.Draw(HandSize),
		})
	}

	// shuffle players for who goes first (and ordering)
	for i := 0; i < shuffleLoop; i++ {
		rand.Shuffle(len(game.players), func(i, j int) {
			game.players[i], game.players[j] = game.players[j], game.players[i]
		})
	}

	// loops through players and links them to who is next
	for i := 0; i < len(game.players); i++ {
		game.players[i].next = &game.players[(i+1)%len(game.players)]
	}
}

// CurrentPlayer returns the active player
func (game *Game) CurrentPlayer() Player {
	return game.Turn.player
}

// NextPlayer points to the next person to go
func (game *Game) NextPlayer() Player {
	return *game.Turn.player.next
}

// SetNextTurn increments the turn counter and changes players
func (game *Game) SetNextTurn(action string) {
	game.Turn.input = action
	game.Turns = append(game.Turns, game.Turn)
	game.Turn.number++
	game.Turn.player = game.NextPlayer()
}

// End enters the final scoring of the game
func (game *Game) End() Player {
	for _, p := range game.players {
		for _, t := range p.Tiles() {
			p.score = p.score - t.value
		}
	}

	return game.HighestScore()
}

// HighestScore finds the player who has the highest current score
func (game *Game) HighestScore() Player {
	var highest Player
	for _, p := range game.players {
		if p.score > highest.score {
			highest = p
		}
	}
	return highest
}

// CheckWord validates the input string
func (game *Game) CheckWord(word string) bool {
	return game.Dictionary.Words[word]
}

// ApplyTurn parses user input and
func (game *Game) ApplyTurn(input string) error {
	var err error
	var placements []TilePlacement

	tokens := strings.Split(input, " ")
	if len(tokens) == 0 {
		return ErrInvalidAction
	}

	switch tokens[0] {
	case "swap":
		// Format of `swap a b c d`
		tokens = tokens[1:]
		tiles := parseTiles(tokens)
		err = game.SwapTiles(tiles)

	case "place":
		// Format of `place a(1,a) b(2,a)`
		tilePlacements := tokens[1:]
		placements, err = parseTilePlacements(tilePlacements)
		if err != nil {
			return err
		}
		err = game.PlaceTiles(placements)
	default:
		return ErrInvalidAction
	}
	if err != nil {
		return err
	}

	game.SetNextTurn(input)

	return nil
}

// SwapTiles is a move a player can execute that puts tiles from their hands back into bag
// first validates enough tiles are remaining
// then validates the
func (game *Game) SwapTiles(tiles []Tile) error {
	player := game.CurrentPlayer()
	if len(game.Tiles.remaining) < len(tiles) {
		return ErrNotEnoughTilesForSwap
	}

	var swapTiles []Tile

	for _, tile := range tiles {
		var found bool
		for i, heldTile := range player.tiles {
			if tile == heldTile {
				// add to reshuffle into tiles (also removing from hand)
				swapTiles = append(swapTiles, heldTile)
				if i == len(player.tiles) {
					player.tiles = player.tiles[:i]
				} else {
					player.tiles = append(player.tiles[:i], player.tiles[i+1:]...)
				}
				found = true
				break
			}
		}
		if !found {
			return ErrTileNotInHand
		}
	}

	player.tiles = append(player.tiles, game.Draw(len(tiles))...)
	game.Tiles.Return(swapTiles)

	game.SetPlayerState(player)

	return nil
}

// PlaceTiles denotes an attempt to play a word
func (game *Game) PlaceTiles(place []TilePlacement) error {
	player := game.CurrentPlayer()

	if game.Turn.number == 1 {
		if !touchesCenter(place) {
			return ErrInvalidStart
		}
	}

	err := validateHand(player, place)
	if err != nil {
		return err
	}

	board := game.GetBoard()
	var words []Word
	direction, start, err := validateTiles(&board, place)
	if err != nil {
		return err
	}

	// Find new word that is being played linearly
	word, _ := FindWord(board, direction, start)
	compareWord := word.String()

	if len(word.String()) > 1 {
		words = append(words, word)
	}

	// iterate across TilePlacements to find Additional words
	// for horizontal moves this will be all adjacent vertical connections to letters
	for _, t := range place {
		word, found := FindWord(board, flipDirection(direction), t.Location)
		if found {
			words = append(words, word)
		}
	}

	if len(words) == 0 {
		return ErrNoValidWordsFound
	}

	var scoreTotal int
	var failedWords []string
	for _, word := range words {
		w := word.String()
		if game.CheckWord(w) {
			scoreTotal += word.ScoreWord()
		} else {
			failedWords = append(failedWords, w)
		}
	}

	if len(failedWords) > 0 {
		return ErrInvalidWords{failedWords}
	}

	if scoreTotal > player.HighestScore() {
		player.highestScore = scoreTotal
		player.highestWord = compareWord
	}

	for _, p := range place {
		board.SetSquareUsed(p.Location)
	}
	if len(place) == HandSize {
		scoreTotal += BINGO
	}

	player.Update(scoreTotal, place)
	player.tiles = append(player.tiles, game.Draw(len(place))...)

	game.SetPlayerState(player)
	game.SetBoard(board)
	return nil
}

// FindWord takes direction and starting index and finds connected Word
// For horizontal
func FindWord(board Board, direction string, coord Coordinate) (Word, bool) {
	var word Word
	x, y := coord.x, coord.y
	word.direction = direction
	word.Squares = append(word.Squares, board[x][y])

	switch direction {
	case "horizontal":
		for i := y + 1; i < Size; i++ {
			if board[x][i].IsEmpty() {
				break
			}
			word.Squares = append(word.Squares, board[x][i])
		}
		for i := y - 1; i > 0; i-- {
			if board[x][i].IsEmpty() {
				break
			}
			word.Squares = append(word.Squares, board[x][i])
		}
	case "vertical":
		for i := x + 1; i < Size; i++ {
			if board[i][y].IsEmpty() {
				break
			}
			word.Squares = append(word.Squares, board[i][y])
		}
		for i := x - 1; i > 0; i-- {
			if board[i][y].IsEmpty() {
				break
			}
			word.Squares = append(word.Squares, board[i][y])
		}
	}

	// Case of only having the single letter present
	if len(word.Squares) == 1 {
		return word, false
	}
	sort.Sort(word)
	return word, true
}

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
			total += s.Value.value
		} else if mult, ok := letterMult[s.Multiplier]; ok {
			total += mult * s.Value.value
		} else {
			total += s.Value.value
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

// TilePlacement represents a request to place a tile at a particular grid coordinate
type TilePlacement struct {
	Location Coordinate
	Tile     Tile
}

// Coordinate represents grid location on board
type Coordinate struct {
	x int
	y int
}

func (c Coordinate) String() string {
	return fmt.Sprintf("(%v,%v)", c.x, c.y)
}

func validateHand(player Player, place []TilePlacement) error {
	for _, t := range place {
		var found bool
		for _, p := range player.tiles {
			if t.Tile == p {
				found = true
				break
			}
			// handle special case of blank character
			if t.Tile.IsBlank && p.Letter == "_" {
				found = true
				break
			}
		}
		if !found {
			return ErrTileNotInHand
		}
	}
	return nil
}

// validateTiles takes an incumbent board and requested tile placements
// validates that all letters are placed in the same direction (vertical/horizontal)
// validates that the range of letters is connected
// TODO refactor this to break out into some kind of sub function that handles
// different options better func(dir string, start, finish int) error
// returns direction, start, end, and error
func validateTiles(board *Board, place []TilePlacement) (direction string, start Coordinate, err error) {
	// verify all tiles are in bounds
	// verify all tiles are in the same horizontal or vertical direction
	var lastX, lastY int
	var empty Tile
	// represent the earliest and latest x,y coordinates to verify board completeness
	var firstX, finalX, firstY, finalY int

	// This check allows for single placements to be calculated
	if len(place) == 1 {
		direction = "horizontal"
	}

	for i, t := range place {
		x, y := t.Location.x, t.Location.y

		// Verify requested coordinates are in range
		if x < 0 || x >= Size || y < 0 || y >= Size {
			err = ErrInvalidSpace
			return
		}
		// Verify placement is not already filled
		if board[x][y].Value != empty {
			err = ErrSpaceOccupied{
				Location: t.Location,
			}
			return
		}

		// CASES:
		// 0: set initial x, y value for later comparison (completeness of word)
		// 1: second iteration, determines if vertical, horizontal, or invalid
		// 2: direction is set, validate against this
		switch i {
		case 0:
			firstX, firstY = x, y
			finalX, finalY = x, y
		case 1:
			switch {
			// Indicates word needs to be vertically placed
			case x == lastX:
				direction = "horizontal"
				if y < firstY {
					firstY = y
				}
				if y > finalY {
					finalY = y
				}
			// Indicates word needs to be horizontally placed
			case y == lastY:
				direction = "vertical"
			default:
				err = ErrInvalidPlacement
				return
			}
		default:
			switch direction {
			case "horizontal":
				if x != lastX {
					err = ErrInvalidPlacement
					return
				}
				if y < firstY {
					firstY = y
				}
				if y > finalY {
					finalY = y
				}
			case "vertical":
				if y != lastY {
					err = ErrInvalidPlacement
					return
				}
				if x < firstX {
					firstX = x
				}
				if x > finalX {
					finalX = x
				}
			}
		}
		board[x][y].Value = t.Tile
		lastX, lastY = x, y
	}

	// Validate starting and ending letters connected
	switch direction {
	case "horizontal":
		for y := firstY; y <= finalY; y++ {
			if board[lastX][y].IsEmpty() {
				err = ErrWordDisconnected
				return
			}
		}
		start = Coordinate{lastX, firstY}
	case "vertical":
		for x := firstX; x <= finalX; x++ {
			if board[x][lastY].IsEmpty() {
				err = ErrWordDisconnected
				return
			}
		}
		start = Coordinate{firstX, lastY}
	}
	return
}

func touchesCenter(place []TilePlacement) bool {
	for _, p := range place {
		if p.Location == CENTER {
			return true
		}
	}
	return false
}

// ErrNoValidWordsFound represents when a user places a single tile but it forms no words
var ErrNoValidWordsFound = fmt.Errorf("no valid words found in tile placements")

// ErrTileNotInHand represents an attempt at using a tile that a player does not have
var ErrTileNotInHand = fmt.Errorf("Tile requested for action but not found")

// ErrNotEnoughTilesForSwap when a player requests a swap that is greater than total tiles remaining
var ErrNotEnoughTilesForSwap = fmt.Errorf("Could not perform swap, not enough tiles remaining")

// ErrWordDisconnected represents a word placement that is not connected
var ErrWordDisconnected = fmt.Errorf("Word placement invalid, gap between letters found")

// ErrInvalidPlacement represents an failed attempt to place a tile non-linearly
var ErrInvalidPlacement = fmt.Errorf("Word placement invalid, must place only horizontal or vertically")

// ErrInvalidSpace indicates an invalid tile placement
var ErrInvalidSpace = fmt.Errorf("Provided space is illegal. Must be in range of [0,%v], [0,%v]", Size-1, Size-1)

// ErrInvalidStart starting turn requires tile be placed in center of board
var ErrInvalidStart = fmt.Errorf("Starting move must touch center tile")

// ErrInvalidAction is when a user attempts to perform an illegal operation
var ErrInvalidAction = fmt.Errorf("Invalid action requested: allowed [swap, place]")

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
