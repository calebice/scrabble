package main

// TODO(s)
// - Add better logic around control loop (error handling/retry)
// - Implement http server package to interact with multiple games at one time
// - refactor game logic to be more easily packaged into simple http requests
// - connect options to actual logic to perform requested function
// - add end state of game logic implemntation

import (
	"bufio"
	"database/sql"
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

	fmt.Println(listOptions())

	for game == nil {
		action := getAction(reader)

		switch action {
		case "new":
			game = instantiateNewGame(reader, gameDB)
		case "load":
			game, err = loadGameInput(reader, gameDB)
		default:
			panic(fmt.Sprintf("Requested action not implemented: %q", action))
		}
		if err != nil {
			fmt.Printf("Could not perform requested action: %v", err)
		}
	}

	runControlLoop(reader, game, gameDB)
}

func getAction(reader *bufio.Reader) string {
	fmt.Println(listOptions())

	fmt.Print("Please enter requested action: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSuffix(input, "\n")
	input = strings.TrimSpace(input)
	if _, ok := optionsMap[input]; ok {
		return input
	}
	fmt.Println("Invalid action requested: ", input)
	return getAction(reader)
}

func listOptions() string {
	str := fmt.Sprintf("Available Options\n--------------------------\n")
	for opt, info := range optionsMap {
		str = fmt.Sprintf("%s%s: %s\n", str, opt, info)
	}
	return str
}

func instantiateNewGame(reader *bufio.Reader, gameDB *scrabble.GameDB) *scrabble.Game {
	fmt.Print("Please enter number of players [1-4]: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSuffix(input, "\n")

	playerCount, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("Invalid input for players")
		instantiateNewGame(reader, gameDB)
	}

	if playerCount < 1 || playerCount > 4 {
		fmt.Println("Invalid number of players")
		instantiateNewGame(reader, gameDB)
	}

	var players []scrabble.PlayerRequest
	for i := 0; i < playerCount; i++ {
		var playerReq scrabble.PlayerRequest

		fmt.Printf("Please enter Player %v's name: ", i+1)
		name, _ := reader.ReadString('\n')
		name = strings.TrimSuffix(name, "\n")
		playerReq.Name = name

		fmt.Printf("Use plaintext? (y/n): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSuffix(input, "\n")
		switch input {
		case "y", "Y":
			playerReq.UsePlainText = true
		default:
			playerReq.UsePlainText = false
		}
		players = append(players, playerReq)
	}

	for _, p := range players {
		fmt.Printf("%+v\n", p)
	}

	return scrabble.NewGame(players, gameDB)
}

func runControlLoop(reader *bufio.Reader, game *scrabble.Game, gameDB *scrabble.GameDB) {
	fmt.Printf("Current game id: %v\n\n", game.GetID())

	for {
		current := game.CurrentPlayer()
		fmt.Printf("%s: %v\n%s\n", current.Name,
			current.Score(), current.Tiles())
		fmt.Println(game.GetBoard().FormatPrint(current.UsePlainText))

		fmt.Print("Please enter move: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSuffix(input, "\n")

		result, err := game.ApplyTurn(input, gameDB)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("%s: %s", current.Name, result.String())

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
	}
}

var optionsMap = map[string]string{
	"list":   "list all current games",
	"new":    "create a new game",
	"load":   "load a game using game id",
	"delete": "delete a game using id",
	"stats":  "display stats for a given player",
}

func loadGameInput(reader *bufio.Reader, gameDB *scrabble.GameDB) (*scrabble.Game, error) {
	fmt.Print("Please enter game ID: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSuffix(input, "\n")

	i, err := strconv.Atoi(input)
	if err != nil {
		fmt.Printf("Invalid game id %q. Please enter a valid id\n", input)
		loadGameInput(reader, gameDB)
	}
	return gameDB.GetGameByID(i)
}
