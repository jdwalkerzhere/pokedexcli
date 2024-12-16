package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"pokedexcli/api"
	"strings"
)

type config struct {
	Next     *string
	Previous *string
}

type cliCommand struct {
	name          string
	description   string
	callback      func(*config) error
	configuration *config
}

var supportedCommands map[string]cliCommand

func commandExit(configuration *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(configuration *config) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	for cmdName, cmd := range supportedCommands {
		fmt.Printf("%s: %s\n", cmdName, cmd.description)
	}
	return nil
}

func commandMap(configuration *config) error {
	res, err := http.Get(*configuration.Next)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var response api.Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return err
	}
	for _, location := range response.Results {
		fmt.Println(location.Name)
	}
	configuration.Next = &response.Next
	configuration.Previous = &response.Prev
	return nil
}

func commandMapB(configuration *config) error {
	if configuration.Previous == nil || *configuration.Previous == "" {
		fmt.Println("you're on the first page")
		start := "https://pokeapi.co/api/v2/location-area"
		configuration.Previous = &start
	}
	res, err := http.Get(*configuration.Previous)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var response api.Response
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return err
	}
	for _, location := range response.Results {
		fmt.Println(location.Name)
	}
	configuration.Next = &response.Next
	configuration.Previous = &response.Prev
	return nil
}

func main() {
	supportedCommands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Display the next 20 areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Display the previous 20 areas",
			callback:    commandMapB,
		},
	}

	scanner := bufio.NewScanner(os.Stdin)
	locationAreaURL := "https://pokeapi.co/api/v2/location-area"
	config := config{
		Next:     &locationAreaURL,
		Previous: nil,
	}

	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()
		cleaned := strings.Fields(strings.ToLower(input))
		command := cleaned[0]
		cmd, ok := supportedCommands[command]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		err := cmd.callback(&config)
		if err != nil {
			fmt.Println(err)
		}
	}
}
