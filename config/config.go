package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var VCSV1ToV2 = map[string]string{
	"github":    "gh",
	"bitbucket": "bb",
	"gitlab":    "gl",
}

type circleciCliConfig struct {
	token string `yaml:"token"`
}

type circlogState struct {
	Organisation string `yaml:"organisation"`
	Vcs          string `yaml:"vcs"`
}

type CirclogConfig struct {
	Branch  string
	Org     string `yaml:"organisation"`
	Project string
	Token   string
	Vcs     string `yaml:"vcs"`
}

func GetToken() (string, bool, error) {
	token, exists := os.LookupEnv("CIRCLECI_TOKEN")
	if exists {
		return token, true, nil
	}

	token, exists, err := getTokenFromCliConfig()
	if err != nil {
		return "", false, err
	} else if !exists {
		return "", false, nil
	} else {
		return token, true, err
	}
}

func getTokenFromCliConfig() (string, bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", false, err
	}

	circleciCliConfigPath, err := filepath.Abs(fmt.Sprintf("%s/.circleci/cli.yml", homeDir))
	if err != nil {
		return "", false, err
	}

	_, err = os.Stat(circleciCliConfigPath)
	if err != nil {
		return "", false, nil
	} else {
		circleciCliConfigFile, _ := os.ReadFile(circleciCliConfigPath)
		var config circleciCliConfig
		err = yaml.Unmarshal(circleciCliConfigFile, &config)
		if err != nil {
			return "", true, fmt.Errorf("could not parse %s", circleciCliConfigPath)
		} else {
			return config.token, true, nil
		}
	}
}

func ensureConfigDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	circlogConfigDir, err := filepath.Abs(fmt.Sprintf("%s/.config/circlog", homeDir))
	if err != nil {
		return err
	}

	if _, err = os.Stat(circlogConfigDir); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(circlogConfigDir, 0755)
		if err != nil {
			return err
		}
	}

	return err
}

func updateState(newState circlogState) error {
	stateFilePath, err := stateFilePath()
	if err != nil {
		return err
	}

	configYaml, err := yaml.Marshal(&newState)
	if err != nil {
		return err
	}

	return os.WriteFile(stateFilePath, configYaml, 0644)
}

func stateFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	stateFilePath, err := filepath.Abs(fmt.Sprintf("%s/.config/circlog/config.yaml", homeDir))
	if err != nil {
		return "", err
	}

	return stateFilePath, err
}

func readConfigFromState(stateFilePath string) (CirclogConfig, error) {
	var config CirclogConfig

	b, err := os.ReadFile(stateFilePath)

	if err == nil {
		err = yaml.Unmarshal(b, &config)
		if err != nil {
			return CirclogConfig{}, fmt.Errorf("could not parse %s", stateFilePath)
		}
	}

	return config, err
}

func updateConfig(config *CirclogConfig, vcs string, org string) error {
	if vcs != "" {
		if _, ok := VCSV1ToV2[vcs]; ok {
			config.Vcs = vcs
		} else {
			return fmt.Errorf("invalid VCS, valid values are ['github', 'bitbucket', 'gitlab']")
		}
	}

	if org != "" {
		config.Org = org
	}

	return nil
}

func NewConfig(project string, vcs string, org string, branch string) (CirclogConfig, error) {
	err := ensureConfigDir()
	if err != nil {
		return CirclogConfig{}, err
	}

	stateFilePath, err := stateFilePath()
	if err != nil {
		return CirclogConfig{}, err
	}

	config, err := readConfigFromState(stateFilePath)
	if err != nil {
		return CirclogConfig{}, err
	}

	err = updateConfig(&config, vcs, org)
	if err != nil {
		return CirclogConfig{}, err
	}

	err = updateState(circlogState{
		Organisation: config.Org,
		Vcs:          config.Vcs,
	})
	if err != nil {
		return CirclogConfig{}, err
	}

	token, exists, err := GetToken()
	if err != nil {
		return CirclogConfig{}, err
	} else if !exists {
		return CirclogConfig{}, errors.New("could not find token in either 'CIRCLECI_TOKEN' env var or CircleCi Cli config.yml")
	}

	config.Project= project
	config.Token = token
	config.Branch = branch

	return config, nil
}

func (config *CirclogConfig) ProjectSlugV2() string {
	vcs := VCSV1ToV2[config.Vcs]

	return fmt.Sprintf("%s/%s/%s", vcs, config.Org, config.Project)
}

func (config *CirclogConfig) ProjectSlugV1() string {
	return fmt.Sprintf("%s/%s/%s", config.Vcs, config.Org, config.Project)
}
