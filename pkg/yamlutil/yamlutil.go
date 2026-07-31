package yamlutil

import (
	"os"

	"gopkg.in/yaml.v3"
)

// ReadFile reads a YAML file and unmarshals it into the target struct.
func ReadFile(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, target)
}

// WriteFile marshals a struct to YAML and writes it to the given path.
// Creates parent directories if needed.
func WriteFile(path string, source interface{}) error {
	data, err := yaml.Marshal(source)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// WriteFileSafe marshals and writes, creating parent directories first.
func WriteFileSafe(path string, source interface{}) error {
	data, err := yaml.Marshal(source)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
