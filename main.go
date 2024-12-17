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
	callback      func(*config, string) error
	configuration *config
}

var supportedCommands map[string]cliCommand

func commandExit(configuration *config, location string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(configuration *config, location string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	for cmdName, cmd := range supportedCommands {
		fmt.Printf("%s: %s\n", cmdName, cmd.description)
	}
	return nil
}

func commandMap(configuration *config, location string) error {
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

func commandMapB(configuration *config, location string) error {
	if configuration.Previous == nil || *configuration.Previous == "" {
		fmt.Println("you're on the first page")
		start := "https://pokeapi.co/api/v2/location-area"
		configuration.Next = &start
		commandMap(configuration, location)
		return nil
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

func explore(configuration *config, location string) error {
	if location == "" {
		fmt.Println("Please provide a location to explore")
		return nil
	}
	res, err := http.Get("https://pokeapi.co/api/v2/location-area/" + location)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var exploreLocation api.LocationArea
	if err := json.NewDecoder(res.Body).Decode(&exploreLocation); err != nil {
		return fmt.Errorf("No Location [%s] Found, Please enter a valid location", location)
	}
	for _, pokemon := range exploreLocation.PokemonEncounters {
		fmt.Printf("- %s\n", pokemon.Pokemon.Name)
	}
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
		"explore": {
			name:        "explore",
			description: "Displays what pokemon appear in a given location",
			callback:    explore,
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
		location := ""
		if len(cleaned) > 1 {
			location = cleaned[1]
		}
		cmd, ok := supportedCommands[command]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		err := cmd.callback(&config, location)
		if err != nil {
			fmt.Println(err)
		}
	}
}
