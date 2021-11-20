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
	score  int
	player Player
}

// Result represents a struct response for a requested turn
type Result struct {
	Words   []Word
	Score   int
	Swapped int
	Action  string
}

func (r Result) String() string {
	switch r.Action {
	case "swap":
		return fmt.Sprintf("successfully swapped %v tiles", r.Swapped)
	case "place":
		return fmt.Sprintf("successfully placed %v for %v points", r.Words, r.Score)
	}
	return "no action implemented"
}

// GetBoard returns the game board
func (game Game) GetBoard() Board {
	return game.board
}

// GetPlayers returns the list of players
func (game Game) GetPlayers() []Player {
	return game.players
}

// GetID returns the id of the game
func (game Game) GetID() int64 {
	return game.id
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
func NewGame(playerReq []PlayerRequest, gameDB *GameDB) *Game {
	tiles := InitializeTiles()
	board := NewBoard()
	game := Game{
		board:   board,
		players: []Player{},
		Tiles:   tiles,
	}

	err := game.AddPlayers(playerReq, gameDB)
	if err != nil {
		panic(err)
	}
	dict, err := LoadDictionary(dictPath)
	if err != nil {
		panic(err)
	}
	game.Dictionary = dict

	err = gameDB.UpsertGame(&game)
	if err != nil {
		panic(err)
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

// AddPlayers instantiates players into the game
func (game *Game) AddPlayers(playerRequests []PlayerRequest, gameDB *GameDB) error {
	for _, p := range playerRequests {
		player := Player{
			Name:         p.Name,
			UsePlainText: p.UsePlainText,
			tiles:        game.Draw(HandSize),
		}
		err := gameDB.InsertPlayer(&player)
		if err != nil {
			return err
		}
		game.players = append(game.players, player)
	}

	// shuffle players for who goes first (and ordering)
	for i := 0; i < shuffleLoop; i++ {
		rand.Shuffle(len(game.players), func(i, j int) {
			game.players[i], game.players[j] = game.players[j], game.players[i]
		})
	}

	// loops through players and links them to who is next
	for i := 0; i < len(game.players); i++ {
		game.players[i].nextID = game.players[(i+1)%len(game.players)].id
	}
	return nil
}

// CurrentPlayer returns the active player
func (game *Game) CurrentPlayer() Player {
	return game.Turn.player
}

// NextPlayer points to the next person to go
func (game *Game) NextPlayer() Player {
	for _, p := range game.players {
		if p.id == game.Turn.player.nextID {
			return p
		}
	}
	panic("could not find requisite player")
}

// SetNextTurn increments the turn counter and changes players
func (game *Game) SetNextTurn() {
	game.Turn = Turn{
		number: game.Turn.number + 1,
		player: game.NextPlayer(),
	}
}

// End enters the final scoring of the game
func (game *Game) End() Player {
	for _, p := range game.players {
		for _, t := range p.Tiles() {
			p.score = p.score - t.Value
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
func (game *Game) ApplyTurn(input string, gameDB *GameDB) (Result, error) {
	var err error
	var placements []TilePlacement
	var score int
	var words []Word
	var result Result

	tokens := strings.Split(input, " ")
	if len(tokens) == 0 {
		return Result{}, ErrInvalidAction
	}

	result.Action = tokens[0]
	tokens = tokens[1:]
	switch result.Action {
	case "swap":
		// Format of `swap a b c d`
		tiles := parseTiles(tokens)
		err = game.SwapTiles(tiles)
		result.Swapped = len(tiles)

	case "place":
		// Format of `place a(1,a) b(2,a)`
		placements, err = parseTilePlacements(tokens)
		if err != nil {
			return Result{}, err
		}
		words, score, err = game.PlaceTiles(placements)
		result.Words = words
		result.Score = score
	default:
		return Result{}, ErrInvalidAction
	}
	if err != nil {
		return Result{}, err
	}
	game.Turn.input = input
	game.Turn.score = score
	game.Turns = append(game.Turns, game.Turn)

	err = gameDB.SaveState(game)
	if err != nil {
		return Result{}, err
	}
	game.SetNextTurn()

	return result, nil
}

// SwapTiles is a move a player can execute that puts tiles from their hands back into bag
// first validates enough tiles are remaining
// then validates the
func (game *Game) SwapTiles(tiles []Tile) error {
	player := game.CurrentPlayer()
	if len(game.Tiles.Remaining) < len(tiles) {
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
func (game *Game) PlaceTiles(place []TilePlacement) ([]Word, int, error) {
	player := game.CurrentPlayer()

	if game.Turn.number == 1 {
		if !touchesCenter(place) {
			return nil, 0, ErrInvalidStart
		}
	}

	err := validateHand(player, place)
	if err != nil {
		return nil, 0, err
	}

	board := game.GetBoard()
	var words []Word
	direction, start, err := validateTiles(&board, place)
	if err != nil {
		return nil, 0, err
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
		return nil, 0, ErrNoValidWordsFound
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
		return nil, 0, ErrInvalidWords{failedWords}
	}

	for _, p := range place {
		board.SetSquareUsed(p.Location)
	}
	if len(place) == HandSize {
		scoreTotal += BINGO
	}

	if scoreTotal > player.HighestScore() {
		player.highestScore = scoreTotal
		player.highestWord = compareWord
	}

	player.Update(scoreTotal, place)
	player.tiles = append(player.tiles, game.Draw(len(place))...)

	game.SetPlayerState(player)
	game.SetBoard(board)
	return words, scoreTotal, nil
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
