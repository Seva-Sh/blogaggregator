package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func (cfg *Config) SetUser(username string) error {
	cfg.CurrentUserName = username
	filePath, err := getConfigFilePath()
	if err != nil {
		fmt.Print("Error building filepath")
		return err
	}

	// Marshal the Go data into a JSON byte slice
	// Use json.MarshalIndent for pretty-printed JSON
	jsonData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON byte slice to a file
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return err
	}

	fmt.Println("Successfully wrote data to json config")
	return nil
}

func UnmarshalReadFile(data []byte) (Config, error) {
	var cfg Config
	err := json.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func Read() (Config, error) {
	filePath, err := getConfigFilePath()
	if err != nil {
		fmt.Print("Error building filepath")
		return Config{}, err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		// Handle any errors that occurred during file reading
		fmt.Printf("File reading error: %v\n", err)
		return Config{}, err
	}

	cfg, err := UnmarshalReadFile(data)
	if err != nil {
		fmt.Print("Error converting data in bytes into JSON")
		return Config{}, err
	}

	return cfg, nil
}

func getConfigFilePath() (string, error) {
	str, err := os.UserHomeDir()
	return filepath.Join(str, configFileName), err
}
