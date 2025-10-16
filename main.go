package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/slajuwomi/gator/internal/config"
	"github.com/slajuwomi/gator/internal/database"

	_ "github.com/lib/pq"
)

// postgres://stephen@localhost:5432/gator
// goose postgres postgres://stephen@localhost:5432/gator up
// goose postgres postgres://stephen@localhost:5432/gator down
func main() {
	var newState config.State
	var commands config.Commands
	newConfig := config.Read()
	newState.Cfg = &newConfig
	commands.AllCommands = make(map[string]func(*config.State, config.Command) error)
	dbURL := newConfig.DbUrl
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	newState.Db = dbQueries
	commands.RegisterNewCommand("login", config.HandlerLogin)
	commands.RegisterNewCommand("register", config.HandlerRegister)
	commands.RegisterNewCommand("reset", config.HandleReset)
	commands.RegisterNewCommand("users", config.HandleGetAllUsers)
	commands.RegisterNewCommand("agg", config.HandleAgg)
	commands.RegisterNewCommand("addfeed", config.MiddlewareLoggedIn(config.HandleAddFeed))
	commands.RegisterNewCommand("feeds", config.HandleGetAllFeeds)
	commands.RegisterNewCommand("follow", config.MiddlewareLoggedIn(config.HandleFeedFollow))
	commands.RegisterNewCommand("following", config.MiddlewareLoggedIn(config.HandleFollowing))
	commands.RegisterNewCommand("unfollow", config.MiddlewareLoggedIn(config.HandleUnfollow))
	commands.RegisterNewCommand("browse", config.MiddlewareLoggedIn(config.HandleBrowse))
	if len(os.Args) < 2 {
		fmt.Println("need at least two arguments")
		os.Exit(1)
	}
	var newCommand config.Command
	newCommand.CommandName = os.Args[1]
	newCommand.Arguments = os.Args[2:]
	err = commands.Run(&newState, newCommand)
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		os.Exit(1)
	}

}
