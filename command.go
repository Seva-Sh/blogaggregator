package main

import (
	"context"
	"errors"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if ok {
		err := handler(s, cmd)
		return err
	} else {
		return errors.New("command is not available")
	}
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f

}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("username is required")
	}

	ctx := context.Background()
	_, err := s.db.GetUser(ctx, cmd.args[0])
	if err != nil {
		fmt.Println("User does not exist")
		os.Exit(1)
	}

	err = s.cfg.SetUser(cmd.args[0])
	if err != nil {

		return err
	}
	fmt.Println("Successfully added the username")

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("username is required")
	}

	ctx := context.Background()
	createdUser, err := s.db.CreateUser(ctx, database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	})
	if err != nil {
		errorMessage := err.Error()
		if strings.Contains(errorMessage, "duplicate") {
			fmt.Println("Duplicate user name")
			os.Exit(1)
		}
		return err
	}

	err = s.cfg.SetUser(cmd.args[0])
	if err != nil {
		return err
	}
	fmt.Println("Successfully created user")
	fmt.Println(createdUser.ID)
	fmt.Println(createdUser.CreatedAt)
	fmt.Println(createdUser.UpdatedAt)
	fmt.Println(createdUser.Name)

	return nil
}

func handlerReset(s *state, cmd command) error {
	ctx := context.Background()
	err := s.db.Reset(ctx)
	if err != nil {
		fmt.Println("Error deleting users: ", err)
		return err
	}

	fmt.Println("Successfully deleted users")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	ctx := context.Background()
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		fmt.Println("Error getting users")
		return err
	}

	for _, user := range users {
		if s.cfg.CurrentUserName == user.Name {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

	return nil
}

// goose postgres postgres://postgres:postgres@localhost:5432/gator down
