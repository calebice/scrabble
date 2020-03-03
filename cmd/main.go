package main

import (
	"bufio"
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

	fmt.Print("Attempt load? (y/n): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSuffix(input, "\n")
	switch input {
	case "y", "Y":
		game, err = loadState(dataPath)
		if err != nil {
			panic(err)
		}
	default:
		game = instantiateNewGame(reader)
	}

	runControlLoop(reader, game)

}

func instantiateNewGame(reader *bufio.Reader) *scrabble.Game {
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

		err = storeState(game)
		if err != nil {
			panic(err)
		}
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

func storeState(game *scrabble.Game) error {
	file, err := os.Create(dataPath)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)
	err = encoder.Encode(game.GetBoard())
	if err != nil {
		return err
	}
	err = encoder.Encode(game.Tiles)
	if err != nil {
		return err
	}
	err = encoder.Encode(game.GetPlayers())
	if err != nil {
		return err
	}
	err = encoder.Encode(game.Turn)

	return err
}

const dataPath = "data/state"
