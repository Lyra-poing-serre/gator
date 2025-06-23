package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func getConfigFilePath() (string, error) {
	home_dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home_dir, configFileName), nil
}

func Read() (Config, error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		fmt.Print(err)
		return Config{}, err
	}
	conf := Config{}
	if err = json.Unmarshal(data, &conf); err != nil {
		fmt.Print(err)
		return Config{}, err
	}
	return conf, nil
}

func write(conf Config) error {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	data, err := json.Marshal(conf)
	if err != nil {
		return err
	}
	err = os.WriteFile(configFilePath, data, 0666)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) SetUser(user string) error {
	if user == "" {
		return errors.New("empty user")
	}
	c.CurrentUserName = user
	err := write(*c)
	if err != nil {
		return err
	}
	return nil
}
