package main

import (
	"database/sql"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Error reading the config:", err)
		return
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		fmt.Println("Error opening the database:", err)
		return
	}
	dbQueries := database.New(db)

	s := &state{cfg: &cfg, db: dbQueries}
	cmd := &commands{
		handlers: make(map[string]func(*state, command) error),
	}

	cmd.register("login", handlerLogin)
	cmd.register("register", handlerRegister)
	cmd.register("reset", handlerReset)
	cmd.register("users", handlerUsers)
	cmd.register("agg", handlerAgg)
	cmd.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmd.register("feeds", handlerFeeds)
	cmd.register("follow", middlewareLoggedIn(handlerFollow))
	cmd.register("following", middlewareLoggedIn(handlerFollowing))
	cmd.register("unfollow", middlewareLoggedIn(handlerUnfollow))

	args := os.Args

	if len(args) < 2 {
		fmt.Println("command name is required")
		os.Exit(1)
	}
	loginCommand := command{
		name: args[1],
		args: args[2:],
	}

	err = cmd.run(s, loginCommand)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// if err := cfg.SetUser("seva"); err != nil {
	// 	fmt.Println("Error setting user:", err)
	// 	return
	// }

	// cfg, err = config.Read()
	// if err != nil {
	// 	fmt.Print("Error reading the config file")
	// 	return
	// }

	// // Pretify JSON config printout
	// b, err := json.MarshalIndent(cfg, "", "  ")
	// if err != nil {
	// 	fmt.Println("error marshalling config:", err)
	// 	return
	// }
	// fmt.Println(string(b))
}
