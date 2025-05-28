package definition

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Parse is responsible for loading the service definitions file (service.toml)
// into a proper Definitions structure.
func Parse() (*Definitions, error) {
	path, err := getServiceTomlPath()
	if err != nil {
		return nil, err
	}

	return ParseFromFile(path)
}

// ParseFromFile is an alternative way of loading a service definitions file
// for outside projects.
func ParseFromFile(path string) (*Definitions, error) {
	defs, err := New()
	if err != nil {
		return nil, err
	}

	if _, err := toml.DecodeFile(path, &defs); err != nil {
		return nil, err
	}

	// Let available the path where we just loaded the file
	defs.path = path
	return defs, nil
}

func getServiceTomlPath() (string, error) {
	path := flag.String("config", "", "Sets the alternative path for 'service.toml' file.")
	flag.Parse()

	if path != nil && *path != "" {
		return *path, nil
	}

	serviceDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(serviceDir, "service.toml"), nil
}

// ParseExternalDefinitions allows loading specific service definitions from its
// file using a custom target. This provides external features (plugins) to load
// their definitions from the same file into their own structures.
func ParseExternalDefinitions(path string, defs interface{}) error {
	if _, err := toml.DecodeFile(path, defs); err != nil {
		return err
	}

	return nil
}
