package main

import (
	"context"
	"errors"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"os"
	"strconv"
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

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %v <time_between_reqs>", cmd.name)
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	fmt.Printf("Collecting feeds every %s\n", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)
	defer ticker.Stop()
	scrapeFeeds(s)
	for ; ; <-ticker.C {
		// ignore returned error of log it
		scrapeFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return errors.New("name and url are required")
	}
	ctx := context.Background()

	feed, err := s.db.CreateFeed(ctx, database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID,
	})
	if err != nil {
		fmt.Println("Failed to create a feed")
		return err
	}

	feedFollow, err := s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		fmt.Println("The insert has failed")
		return err
	}

	fmt.Println("Followed feed:", feedFollow.FeedName)
	fmt.Printf("%+v\n", feed)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	ctx := context.Background()
	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		fmt.Println("Error getting feeds")
		return err
	}

	for _, feed := range feeds {
		user, err := s.db.GetUserViaId(ctx, feed.UserID)
		if err != nil {
			fmt.Println("Failed to obtain user")
			return err
		}
		fmt.Println(feed.Name)
		fmt.Println(feed.Url)
		fmt.Println(user.Name)
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return errors.New("url is required")
	}
	ctx := context.Background()
	url := cmd.args[0]
	feed, err := s.db.GetFeedViaUrl(ctx, url)
	if err != nil {
		fmt.Println("Failed to fetch feed")
		return err
	}

	feedFollow, err := s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		fmt.Println("The insert has failed")
		return err
	}

	fmt.Println(feedFollow.FeedName)
	fmt.Println(feedFollow.UserName)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	ctx := context.Background()

	feedFollows, err := s.db.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		fmt.Println("Error getting feeds")
		return err
	}

	fmt.Println("Current User:", user.Name)
	for _, feedFollow := range feedFollows {
		fmt.Println(feedFollow.FeedName)
	}
	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {

	return func(s *state, cmd command) error {
		if s.cfg.CurrentUserName == "" {
			return errors.New("no user logged in yet")
		}
		ctx := context.Background()
		currentUser, err := s.db.GetUser(ctx, s.cfg.CurrentUserName)
		if err != nil {
			fmt.Println("failed to fetch current user")
			return err
		}
		return handler(s, cmd, currentUser)
	}

}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return errors.New("url is required")
	}

	ctx := context.Background()
	feed, err := s.db.GetFeedViaUrl(ctx, cmd.args[0])
	if err != nil {
		fmt.Println("error fetching feed")
		return err
	}
	err = s.db.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		fmt.Println("failed to delete feed follow for the current user")
		return err
	}

	fmt.Println("unfollow successfull")
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := int32(2)
	if len(cmd.args) > 0 {
		n, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %v", err)
		}
		limit = int32(n)
	}
	ctx := context.Background()
	posts, err := s.db.GetPostsForUser(ctx, database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  limit,
	})
	if err != nil {
		fmt.Println("error getting posts")
		return err
	}

	for _, post := range posts {
		fmt.Println(post.Title)
		fmt.Println(post.Url)
		fmt.Println(post.PublishedAt)
	}

	return nil
}

// goose postgres postgres://postgres:postgres@localhost:5432/gator up
