package config

import (
	"errors"
	"fmt"
)

type State struct {
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

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.AllCommands[name] = f
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Arguments) == 0 {
		return errors.New("username is required")
	}
	err := s.Cfg.SetUser(cmd.Arguments[0])
	if err != nil {
		return err 
	}
	fmt.Println("User has been set!")
	return nil
}
