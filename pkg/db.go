package scrabble

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// GameDB represents the internal scrabble state as a db schema
type GameDB struct {
	db *sql.DB
}

const createPlayerTable = `CREATE TABLE if not exists players(
	id INTEGER PRIMARY KEY,
	name TEXT
)`

const createGameTable = `CREATE TABLE if not exists games(
	id INTEGER PRIMARY KEY,
	board BLOB,
	tiles BLOB
)`

const createTurnsTable = `CREATE TABLE if not exists turns(
	id INTEGER PRIMARY KEY,
	number INTEGER,
	input TEXT,
	playerID INTEGER,
	tiles BLOB,
	gameID INTEGER,
	FOREIGN KEY(playerID) REFERENCES players(id),
	FOREIGN KEY(gameID) REFERENCES games(id)
)`

const createScoresTable = `CREATE TABLE if not exists scores(
	id INTEGER PRIMARY KEY,
	gameID INTEGER,
	playerID INTEGER,
	turnID INTEGER
	score INTEGER,
	FOREIGN KEY(turnID) REFERENCES turns(id),
	FOREIGN KEY(playerID) REFERENCES players(id),
	FOREIGN KEY(gameID) REFERENCES games(id)
)`

const createMatchesTable = `CREATE TABLE if not exists matches(
	id INTEGER PRIMARY KEY,
	gameID INTEGER
	playerID INTEGER	
)`

// NewDB instantiates a gameDB
func NewDB(db *sql.DB) GameDB {
	return GameDB{
		db: db,
	}
}

// InitDB starts the db tables if it does not already exist
func (db *GameDB) InitDB() error {
	// create players table
	statement, _ := db.db.Prepare(createPlayerTable)
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

	// create turns table
	statement, _ = db.db.Prepare(createTurnsTable)
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	// create scores table
	statement, _ = db.db.Prepare(createScoresTable)
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	return nil
}

// InsertPlayer adds a new player into the db
func (db *GameDB) InsertPlayer(player Player) (int64, error) {
	statement, _ := db.db.Prepare("INSERT INTO players (name) VALUES (?)")
	result, err := statement.Exec(player.Name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetGameByID joins all of the fields to instantiate a game state
func (db *GameDB) GetGameByID(id int) (*Game, error) {
	// TODO join all necessary tables for this

	// Grab the board, and current tiles
	query := `SELECT id, board, tiles FROM games WHERE id = ?`
	statement, err := db.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	rows, err := statement.Query(id)
	if err != nil {
		return nil, err
	}
	var game Game
	for rows.Next() {
		rows.Scan(&game.id, &game.board, &game.Tiles)
	}

	// Grab the list of players in this game
	matchQuery := `SELECT playerID WHERE gameID = ?`
	statement, err = db.db.Prepare(matchQuery)
	if err != nil {
		return nil, err
	}
	rows, err = statement.Query(id)
	if err != nil {
		return nil, err
	}
	var ids []int64
	for rows.Next() {
		var pID int64
		rows.Scan(&pID)
		ids = append(ids, pID)
	}

	// Grab player information
	playersQuery := `SELECT player.id, name, FROM players WHERE id in (?)`
	statement, err = db.db.Prepare(playersQuery)
	if err != nil {
		return nil, err
	}
	rows, err = statement.Query(ids)
	if err != nil {
		return nil, err
	}
	var players []Player
	for rows.Next() {
		var player Player
		rows.Scan(player.id, player.Name)
		players = append(players, player)
	}

	// TODO
	// Get a list of the turns for the current game
	// Grab players current scores (scores table latest turn)

	game.players = players
	dict, err := LoadDictionary(dictPath)
	if err != nil {
		return nil, err
	}
	game.Dictionary = dict
	return &game, nil
}
