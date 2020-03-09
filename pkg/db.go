package scrabble

import (
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"
)

// TODO(s):
// - Add transactions for db layer
// - Implement logic to insert results into historical table (upon finishing game)
// - Add support for tracking word usages by player (words table)
// - Add metadata to various tables
// 	 - users: clarifying information/login information to enforce unique players
// 	 - games: add creation dates
// - Implement queries to aggregate data about individual players (max scores/highest word/etc)

// GameDB represents the internal scrabble state as a db schema
type GameDB struct {
	db *sql.DB
}

// users: users of the scrabble game
const createUsersTable = `CREATE TABLE if not exists users(
	id INTEGER PRIMARY KEY,
	name TEXT
)`

// games: holds board and remaining tile information
const createGameTable = `CREATE TABLE if not exists games(
	id INTEGER PRIMARY KEY,
	board BLOB,
	tiles BLOB
)`

// player_states: tracks the score and tiles for a given player in a game
const createPlayerStatesTable = `CREATE TABLE if not exists player_states(
	id INTEGER PRIMARY KEY,
	game_id INTEGER,
	player_id INTEGER,
	next INTEGER,
	score INTEGER,
	tiles BLOB,
	FOREIGN KEY(player_id) REFERENCES users(id),
	FOREIGN KEY(next) REFERENCES player_states(id),
	FOREIGN KEY(game_id) REFERENCES games(id)
)`

// turns: live turn data with applied operation/result
const createTurnsTable = `CREATE TABLE if not exists turns(
	id INTEGER PRIMARY KEY,
	number INTEGER,
	input TEXT,
	score INTEGER,
	gp_id INTEGER,
	next_player INTEGER,
	FOREIGN KEY(gp_id) REFERENCES player_states(id)
	FOREIGN KEY(next_player) REFERENCES player_states(id)
)`

// historical: the historical game data for a given player (links to games played)
const createHistoricalTable = `CREATE TABLE if not exists historical(
	id INTEGER PRIMARY KEY,
	score INTEGER,
	max_single INTEGER,
	max_word TEXT,
	won  BOOLEAN,
	gp_id INTEGER,
	FOREIGN KEY(gp_id) REFERENCES player_states(id)
)`

// words_played represents the historical data around words that have been played
// includes reference back to user playing word
// TODO actually use this
const createWordsTable = `CREATE TABLE if not exists words_played(
	id INTEGER PRIMARY KEY,	
	word TEXT,
	userID INTEGER,
	score INTEGER,
	FOREIGN KEY(userID) REFERENCES users(id)
)`

// NewDB instantiates a gameDB
func NewDB(db *sql.DB) *GameDB {
	return &GameDB{
		db: db,
	}
}

// InitDB starts the db tables if it does not already exist
func (db *GameDB) InitDB() error {
	// create users table
	statement, _ := db.db.Prepare(createUsersTable)
	_, err := statement.Exec()
	if err != nil {
		return err
	}

	// create games table
	statement, err = db.db.Prepare(createGameTable)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	// create game users table
	statement, _ = db.db.Prepare(createPlayerStatesTable)
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	// create turns table
	statement, _ = db.db.Prepare(createTurnsTable)
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	// create historical games table
	statement, _ = db.db.Prepare(createHistoricalTable)
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	// create words table
	statement, _ = db.db.Prepare(createWordsTable)
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	return nil
}

// GetGameByID joins all of the fields to instantiate a game state
// joins data from games, turns, users and player_states tables
func (db *GameDB) GetGameByID(id int) (*Game, error) {
	// get the board, and current tiles
	query := `
	SELECT id, board, tiles FROM games WHERE id = ?`
	statement, err := db.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	rows, err := statement.Query(id)
	if err != nil {
		return nil, err
	}

	var boardBytes []byte
	var tileBytes []byte

	var game Game
	for rows.Next() {
		rows.Scan(&game.id, &boardBytes, &tileBytes)
	}
	err = json.Unmarshal(boardBytes, &game.board)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(tileBytes, &game.Tiles)
	if err != nil {
		return nil, err
	}

	// Retrieves all relevent player information
	// name, score, tiles, next player
	// join tables linking user_id to player_states.player_id
	playersQuery := `
		SELECT users.id, player_states.id, users.name, player_states.score, player_states.tiles, player_states.next
		FROM users JOIN player_states ON users.id = player_states.player_id
		WHERE player_states.game_id = ?`
	statement, err = db.db.Prepare(playersQuery)
	if err != nil {
		return nil, err
	}
	rows, err = statement.Query(id)
	if err != nil {
		return nil, err
	}
	var players []Player
	for rows.Next() {
		var player Player
		var tileBytes []byte

		rows.Scan(&player.id, &player.pStateID, &player.Name, &player.score, &tileBytes, &player.nextID)
		err = json.Unmarshal(tileBytes, &player.tiles)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	game.players = players

	// Load in dictionary from disk
	dict, err := LoadDictionary(dictPath)
	if err != nil {
		return nil, err
	}
	game.Dictionary = dict

	// Load turn data
	turnsQuery := `
	SELECT number, input, turns.score, next_player
	FROM turns JOIN player_states ON turns.gp_id = player_states.id
	WHERE player_states.game_id = ?`
	statement, err = db.db.Prepare(turnsQuery)
	if err != nil {
		return nil, err
	}
	rows, err = statement.Query(id)
	if err != nil {
		return nil, err
	}
	var turns []Turn
	var maxNum int
	var nextID int64
	for rows.Next() {
		var turn Turn
		var next int64
		rows.Scan(&turn.number, &turn.input, &turn.score, &next)
		turns = append(turns, turn)
		if maxNum < turn.number {
			maxNum = turn.number
			nextID = next
		}
	}

	// find current player and create a fill in turn for game state
	var current Turn
	// if game was started but no moves made, set initial player
	if len(turns) == 0 {
		current = Turn{
			number: maxNum + 1,
			player: players[0],
		}
	} else {
		// find and set desired player
		player := findPlayer(players, nextID)
		current = Turn{
			number: maxNum + 1,
			player: *player,
		}
	}

	game.Turns = turns
	game.Turn = current

	return &game, nil
}

// UpsertGame updates a game if it exists, creates a new one if not
func (db *GameDB) UpsertGame(game *Game) error {
	if game.id != 0 {
		err := db.updateGame(game)
		if err != nil {
			return err
		}
	}
	gameQuery := `INSERT INTO games (board, tiles) VALUES(?, ?)`
	boardJSON, err := json.Marshal(game.board)
	if err != nil {
		return err
	}

	tilesJSON, err := json.Marshal(game.Tiles)
	if err != nil {
		return err
	}

	statement, _ := db.db.Prepare(gameQuery)
	result, err := statement.Exec(boardJSON, tilesJSON)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	game.id = id

	err = db.insertPlayerState(game)
	if err != nil {
		return err
	}
	game.Turn = Turn{
		player: game.players[0],
		number: 1,
	}
	return nil
}

func (db *GameDB) insertPlayerState(game *Game) error {
	playerStateQuery := `INSERT INTO player_states (game_id, player_id, next, score, tiles) VALUES (?, ?, ?, ?, ?)`
	statement, err := db.db.Prepare(playerStateQuery)
	if err != nil {
		return err
	}
	for i, p := range game.players {
		tilesJSON, err := json.Marshal(p.tiles)
		if err != nil {
			return err
		}

		result, err := statement.Exec(game.id, p.id, p.nextID, p.score, tilesJSON)
		if err != nil {
			return err
		}
		gpID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		if gpID == 0 {
			return ErrInsertPlayerState
		}
		game.players[i].pStateID = gpID
	}
	return nil
}

// InsertTurn inputs the executed turn
func (db *GameDB) InsertTurn(turn Turn) error {
	insertQuery := `INSERT INTO turns (gp_id, number, input, score, next_player)
	VALUES (?, ?, ?, ?, ?)
	`

	statement, _ := db.db.Prepare(insertQuery)
	result, err := statement.Exec(
		turn.player.pStateID,
		turn.number,
		turn.input,
		turn.score,
		turn.player.nextID)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	if id == 0 {
		return ErrInsertTurnFailed
	}

	return nil
}

// SaveState will run once per turn and it will execute a series of
// db requests to modify player states, the turn and then the board
func (db *GameDB) SaveState(game *Game) error {

	// ranges across all players and updates current score/tiles
	for _, p := range game.players {
		err := db.updatePlayerState(game, p)
		if err != nil {
			return err
		}
	}

	// Add current turn as an entry
	err := db.InsertTurn(game.Turn)
	if err != nil {
		return err
	}

	// update the game
	err = db.updateGame(game)
	if err != nil {
		return err
	}

	return nil
}

func (db *GameDB) updateGame(game *Game) error {
	boardJSON, err := json.Marshal(game.board)
	if err != nil {
		return err
	}
	tilesJSON, err := json.Marshal(game.Tiles)
	if err != nil {
		return err
	}

	insertQuery := `
	UPDATE games 
	SET board = ?, tiles = ?
	WHERE id = ?`
	statement, _ := db.db.Prepare(insertQuery)
	result, err := statement.Exec(boardJSON, tilesJSON, game.id)
	if err != nil {
		return err
	}
	id, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if id == 0 {
		return ErrCouldNotUpdateGame
	}
	return nil
}

func (db *GameDB) updatePlayerState(game *Game, player Player) error {

	updateQuery := `
	UPDATE player_states
	SET score = ?, tiles = ?
	WHERE id = ?`

	tilesJSON, err := json.Marshal(player.tiles)
	if err != nil {
		return err
	}

	statement, _ := db.db.Prepare(updateQuery)
	result, err := statement.Exec(player.score, tilesJSON, player.pStateID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCouldNotUpdatePlayerState
	}
	return nil
}

// InsertPlayer adds a new player into the db, or returns id of existing player
func (db *GameDB) InsertPlayer(player *Player) error {

	// check for existing user
	existingPlayer, err := db.getUserByName(player.Name)
	if err != nil {
		return err
	}
	if existingPlayer != nil {
		player.id = existingPlayer.id
		return nil
	}

	statement, _ := db.db.Prepare("INSERT INTO users (name) VALUES (?)")
	result, err := statement.Exec(player.Name)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	player.id = id
	return nil
}

func (db *GameDB) getUserByName(name string) (*Player, error) {
	var player Player

	getQuery := `SELECT id, name FROM users WHERE name = ?`
	statement, err := db.db.Prepare(getQuery)
	if err != nil {
		return nil, err
	}
	rows, err := statement.Query(name)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		rows.Scan(&player.id, &player.Name)
	}
	if player.id == 0 {
		return nil, nil
	}

	return &player, nil
}
