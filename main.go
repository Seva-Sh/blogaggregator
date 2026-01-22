package main

import (
	"fmt"
	"gator/internal/config"
	"os"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Error reading the config:", err)
		return
	}

	s := &state{cfg: &cfg}
	c := &commands{
		handlers: make(map[string]func(*state, command) error),
	}

	c.register("login", handlerLogin)

	args := os.Args

	if len(args) < 2 {
		fmt.Println("command name is required")
		os.Exit(1)
	}
	loginCommand := command{
		name: args[1],
		args: args[2:],
	}

	err = c.run(s, loginCommand)
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
