package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
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
	callback      func(*config, map[string]api.Pokemon, string) error
	configuration *config
}

var supportedCommands map[string]cliCommand

func commandExit(configuration *config, pokedex map[string]api.Pokemon, userInput string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(configuration *config, pokedex map[string]api.Pokemon, userInput string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	for cmdName, cmd := range supportedCommands {
		fmt.Printf("%s: %s\n", cmdName, cmd.description)
	}
	return nil
}

func commandMap(configuration *config, pokedex map[string]api.Pokemon, userInput string) error {
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

func commandMapB(configuration *config, pokedex map[string]api.Pokemon, userInput string) error {
	if configuration.Previous == nil || *configuration.Previous == "" {
		fmt.Println("you're on the first page")
		start := "https://pokeapi.co/api/v2/location-area"
		configuration.Next = &start
		commandMap(configuration, pokedex, userInput)
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

func explore(configuration *config, pokedex map[string]api.Pokemon, location string) error {
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
		return fmt.Errorf("No Location [%s] Found, Please enter a valid location\n", location)
	}
	for _, pokemon := range exploreLocation.PokemonEncounters {
		fmt.Printf("- %s\n", pokemon.Pokemon.Name)
	}
	return nil
}

func catch(configuration *config, pokedex map[string]api.Pokemon, pokemon string) error {
	if pokemon == "" {
		fmt.Println("Please provide a Pokemon to try to catch")
		return nil
	}
	res, err := http.Get("https://pokeapi.co/api/v2/pokemon/" + pokemon)
	if err != nil {
		return err
	}

	var catchPokemon api.Pokemon
	if err := json.NewDecoder(res.Body).Decode(&catchPokemon); err != nil {
		return fmt.Errorf("No Pokemon [%s] Found, Please enter a valid Pokemon\n", pokemon)
	}
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon)
	// Attempt Catching logic
	attempt := rand.Intn(catchPokemon.BaseExperience)
	if attempt > catchPokemon.BaseExperience/2 {
		fmt.Printf("%s was caught!\n", pokemon)
		pokedex[pokemon] = catchPokemon
	} else {
		fmt.Printf("%s escaped!\n", pokemon)
	}
	return nil
}

func inspect(configuration *config, pokedex map[string]api.Pokemon, pokemon string) error {
	userPokemon, ok := pokedex[pokemon]
	if !ok {
		return fmt.Errorf("you have not caught that pokemon")
	}
	fmt.Printf("Name: %s\n", userPokemon.Name)
	fmt.Printf("Height: %d\n", userPokemon.Height)
	fmt.Printf("Weight: %d\n", userPokemon.Weight)
	fmt.Println("Stats:")
	for _, stat := range userPokemon.Stats {
		fmt.Printf("\t-%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Println("Types:")
	for _, pokemonType := range userPokemon.Types {
		fmt.Printf("\t- %s\n", pokemonType.Type.Name)
	}
	return nil
}

func pokedex(configuration *config, pokedex map[string]api.Pokemon, pokemon string) error {
	if len(pokedex) == 0 {
		return fmt.Errorf("You have not caught any Pokemon, Go catch some!")
	}
	fmt.Println("Your Pokedex")
	for name := range pokedex {
		fmt.Printf("\t-%s\n", name)
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
		"catch": {
			name:        "catch",
			description: "Attempt to capture a given pokemon",
			callback:    catch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect given pokemon",
			callback:    inspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Display the pokemon you've caught",
			callback:    pokedex,
		},
	}

	scanner := bufio.NewScanner(os.Stdin)
	locationAreaURL := "https://pokeapi.co/api/v2/location-area"
	config := config{
		Next:     &locationAreaURL,
		Previous: nil,
	}
	pokedex := map[string]api.Pokemon{}

	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()
		cleaned := strings.Fields(strings.ToLower(input))
		command := cleaned[0]
		userInput := ""
		if len(cleaned) > 1 {
			userInput = cleaned[1]
		}
		cmd, ok := supportedCommands[command]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}
		err := cmd.callback(&config, pokedex, userInput)
		if err != nil {
			fmt.Println(err)
		}
	}
}
