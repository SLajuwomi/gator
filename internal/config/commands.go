package config

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slajuwomi/gator/internal/database"
)

type State struct {
	Db  *database.Queries
	Cfg *Config
}

type Command struct {
	CommandName string
	Arguments   []string
}

type Commands struct {
	AllCommands map[string]func(*State, Command) error
}

func (c *Commands) Run(s *State, cmd Command) error {
	err := c.AllCommands[cmd.CommandName](s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *Commands) RegisterNewCommand(name string, f func(*State, Command) error) {
	c.AllCommands[name] = f
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Arguments) == 0 {
		return errors.New("username is required")
	}
	_, err := s.Db.GetUser(context.Background(), cmd.Arguments[0])
	if err != nil {
		return fmt.Errorf("user not found in database: %v", err)
	}
	err = s.Cfg.SetUser(cmd.Arguments[0])
	if err != nil {
		return err
	}
	fmt.Println("User has been set!")
	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Arguments) == 0 {
		return errors.New("username is required")
	}
	dbUser, err := s.Db.CreateUser(context.Background(), database.CreateUserParams{uuid.New(), time.Now(), time.Now(), cmd.Arguments[0]})
	if err != nil {
		return fmt.Errorf("could not register: %v", err)
	}
	err = s.Cfg.SetUser(cmd.Arguments[0])
	if err != nil {
		return err
	}
	fmt.Printf("User has been set!\n%+v", dbUser)
	return nil
}

func HandleReset(s *State, cmd Command) error {
	err := s.Db.Clear(context.Background())
	if err != nil {
		return fmt.Errorf("reset failed: %v", err)
	} else {
		fmt.Println("Reset successful!")
	}
	return nil
}
