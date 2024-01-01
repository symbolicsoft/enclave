// SPDX-FileCopyrightText: © 2024 Nadim Kobeissi <nadim@symbolic.software>
// SPDX-License-Identifier: GPL-2.0-only

package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func EnsurePath() string {
	configPath := ""
	switch runtime.GOOS {
	case "windows":
		localAppDataPath, _ := os.UserCacheDir()
		configPath = filepath.Join(localAppDataPath, "Enclave")
	case "linux":
		homePath, _ := os.UserHomeDir()
		configPath = filepath.Join(homePath, ".config", "enclave")
	case "darwin":
		homePath, _ := os.UserHomeDir()
		configPath = filepath.Join(homePath, "Library", "Application Support", "Enclave")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		os.MkdirAll(configPath, 0o700)
	}
	return filepath.Join(configPath, "keys")
}

func ConfigFileExists() error {
	configFilePath := EnsurePath()
	var err error
	if _, err = os.Stat(configFilePath); os.IsNotExist(err) {
		return err
	}
	_, err = os.ReadFile(configFilePath)
	return err
}

func Read() ([]string, error) {
	var configFileBytes []byte
	configFilePath := EnsurePath()
	configFileBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return []string{}, err
	}
	configFileStrings := strings.Split(string(configFileBytes), "\n")
	return configFileStrings, nil
}

func Write(configFileStrings []string) error {
	configFilePath := EnsurePath()
	return os.WriteFile(configFilePath, []byte(strings.Join(configFileStrings, "\n")), 0o600)
}

func Delete() {
	configFilePath := EnsurePath()
	if ConfigFileExists() == nil {
		os.Remove(configFilePath)
	}
}
