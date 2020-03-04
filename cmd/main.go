package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	scrabble "github.com/calebice/scrabble/pkg"
)

func main() {
	reader := bufio.NewReader(os.Stdin) //create new reader, assuming bufio imported
	var game *scrabble.Game
	var err error

	database, err := sql.Open("sqlite3", "./game.db")
	if err != nil {
		panic(err)
	}

	gameDB := scrabble.NewDB(database)
	err = gameDB.InitDB()
	if err != nil {
		panic(err)
	}

	fmt.Print("Load game? (enter id if saved): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSuffix(input, "\n")
	if i, err := strconv.Atoi(input); err == nil {
		game, err = gameDB.GetGameByID(i)
		if err != nil {
			panic(err)
		}
	} else {
		game = instantiateNewGame(reader, gameDB)
	}

	runControlLoop(reader, game)

}

func instantiateNewGame(reader *bufio.Reader, gameDB scrabble.GameDB) *scrabble.Game {
	fmt.Print("Please enter number of players: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSuffix(input, "\n")

	playerCount, err := strconv.Atoi(input)
	if err != nil {
		panic(err)
	}

	var players []string
	for i := 0; i < playerCount; i++ {
		fmt.Printf("Please enter Player %v's name: ", i+1)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSuffix(input, "\n")

		players = append(players, input)
	}

	return scrabble.NewGame(players)
}

func runControlLoop(reader *bufio.Reader, game *scrabble.Game) {
	for {
		fmt.Printf("%s: %v\n%s\n", game.CurrentPlayer().Name,
			game.CurrentPlayer().Score(), game.CurrentPlayer().Tiles())
		fmt.Println(game.GetBoard())

		fmt.Print("Please enter move: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSuffix(input, "\n")

		err := game.ApplyTurn(input)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println()

		fmt.Println("Tiles Remaining: ", len(game.Tiles.GetTiles()))
		fmt.Println("Scores")
		fmt.Println("--------------------")
		for _, p := range game.GetPlayers() {
			fmt.Printf("%s: %v\n", p.Name, p.Score())
		}
		fmt.Println("--------------------")

		if len(game.CurrentPlayer().Tiles()) == 0 {
			winner := game.End()
			fmt.Printf("Winning player: %s with %v points", winner.Name, winner.Score())
			fmt.Println("Stats")
			fmt.Println("--------------------")
			for _, p := range game.GetPlayers() {
				fmt.Printf("%s: %v points, highest scoring word: %s %v points", p.Name, p.Score(), p.HighestWord(), p.HighestScore())
			}
			return
		}

		// TODO store state of game using db layer
		// err = storeState(game)
		// if err != nil {
		// 	panic(err)
		// }
	}
}

func loadState(path string) (*scrabble.Game, error) {
	var board scrabble.Board
	var players []scrabble.Player
	var turn scrabble.Turn
	var tiles scrabble.Tiles

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(file).Decode(&board)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(file).Decode(&tiles)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(file).Decode(&players)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(file).Decode(&turn)
	if err != nil {
		return nil, err
	}

	game := scrabble.LoadFromState(board, tiles, players, turn)
	return &game, nil
}
