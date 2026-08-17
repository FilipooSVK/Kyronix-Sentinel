package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Load loads configuration from file.
//
// If file is missing,
// default configuration is returned.
func Load(
	path string,
) (Config, error) {

	config := Default()

	data, err := os.ReadFile(path)

	if err != nil {

		if os.IsNotExist(err) {
			return config, nil
		}

		return config, err
	}

	err = yaml.Unmarshal(
		data,
		&config,
	)

	if err != nil {
		return config, err
	}

	return config, nil
}
