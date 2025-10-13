package main

import (
	"fmt"
	"os"

	"github.com/slajuwomi/gator/internal/config"
)

func main() {
	var newState config.State
	var commands config.Commands
	newConfig := config.Read()
	newState.Cfg = &newConfig
	commands.AllCommands = make(map[string]func(*config.State, config.Command) error)
	
	commands.Register("login", config.HandlerLogin)
	if len(os.Args) < 2 {
		fmt.Println("need at least two arguments")
		os.Exit(1)
	}
	var newCommand config.Command
	newCommand.CommandName = os.Args[1]
	newCommand.Arguments = os.Args[2:]
	err := commands.Run(&newState, newCommand)
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		os.Exit(1)
	}
}
