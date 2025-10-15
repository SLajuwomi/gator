package config

import (
	"context"
	"fmt"

	"github.com/slajuwomi/gator/internal/database"
)

func MiddlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		currentUser, err := s.Db.GetUser(context.Background(), s.Cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("failed getting current user from database: %v", err)
		}
		return handler(s, cmd, currentUser)
	}
}