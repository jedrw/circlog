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

type circleCiCliConfig struct {
	Token string `yaml:"token"`
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

	token, exists, err := getTokenFromCircleCiCliConfig()
	if err != nil {
		return "", false, err
	} else if !exists {
		return "", false, nil
	} else {
		return token, true, err
	}
}

func getTokenFromCircleCiCliConfig() (string, bool, error) {
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
	}

	circleciCliConfigBytes, err := os.ReadFile(circleciCliConfigPath)
	if err != nil {
		return "", false, err
	}

	var config circleCiCliConfig
	err = yaml.Unmarshal(circleciCliConfigBytes, &config)
	if err != nil {
		return "", true, fmt.Errorf("could not parse %s", circleciCliConfigPath)
	} else if config.Token == "" {
		return "", false, err
	} else {
		return config.Token, true, err
	}
}

func ensureConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	circlogConfigDir, err := filepath.Abs(fmt.Sprintf("%s/.config/circlog", homeDir))
	if err != nil {
		return circlogConfigDir, err
	}

	err = os.MkdirAll(circlogConfigDir, 0755)
	if err != nil {
		return circlogConfigDir, err
	}

	return circlogConfigDir, err
}

func ensureStateFile() (string, error) {
	circlogConfigDir, err := ensureConfigDir()
	if err != nil {
		return "", err
	}

	circlogConfigFile, err := filepath.Abs(fmt.Sprintf("%s/config.yaml", circlogConfigDir))
	if err != nil {
		return circlogConfigFile, err
	}

	if _, err = os.Stat(circlogConfigFile); errors.Is(err, os.ErrNotExist) {
		_, err = os.Create(circlogConfigFile)
		if err != nil {
			return circlogConfigFile, err
		}
	}

	return circlogConfigFile, err
}

func updateState(stateFilePath string, newState circlogState) error {
	configYaml, err := yaml.Marshal(&newState)
	if err != nil {
		return err
	}

	return os.WriteFile(stateFilePath, configYaml, 0644)
}

func readConfigFromState(stateFilePath string) (CirclogConfig, error) {
	var config CirclogConfig

	b, err := os.ReadFile(stateFilePath)
	if err != nil {
		return CirclogConfig{}, err
	}

	err = yaml.Unmarshal(b, &config)
	if err != nil {
		return CirclogConfig{}, fmt.Errorf("could not parse %s", stateFilePath)
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
	circlogStateFile, err := ensureStateFile()
	if err != nil {
		return CirclogConfig{}, err
	}

	config, err := readConfigFromState(circlogStateFile)
	if err != nil {
		return config, err
	}

	if vcs != "" || org != "" {
		err := updateConfig(&config, vcs, org)
		if err != nil {
			return config, err
		}

		err = updateState(circlogStateFile, circlogState{
			Organisation: config.Org,
			Vcs:          config.Vcs,
		})
		if err != nil {
			return config, err
		}
	}

	if config.Org == "" {
		return config, fmt.Errorf("organisation is not set")
	}

	if config.Vcs == "" {
		return config, fmt.Errorf("vcs is not set")
	}

	token, exists, err := GetToken()
	if err != nil {
		return config, err
	} else if !exists {
		return config, errors.New("could not find token in either 'CIRCLECI_TOKEN' env var or CircleCi Cli cli.yml")
	}

	config.Project = project
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
