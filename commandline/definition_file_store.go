package commandline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefinitionFileStore discovers the definition files from disk searching for
// the definitions/ folder and returns the data for a particular definition file.
type DefinitionFileStore struct {
	directory   string
	files       []string
	definitions []DefinitionData
}

const DefinitionsDirectory = "definitions"

func (s *DefinitionFileStore) Names(version string) ([]string, error) {
	if s.definitions != nil {
		names := []string{}
		for _, definition := range s.definitions {
			if version == definition.Version {
				names = append(names, definition.Name)
			}
		}
		return names, nil
	}

	definitionFiles, err := s.discoverDefinitions(version)
	if err != nil {
		return nil, err
	}
	return s.definitionNames(definitionFiles), nil
}

func (s *DefinitionFileStore) Read(name string, version string) (*DefinitionData, error) {
	if s.definitions != nil {
		for _, definition := range s.definitions {
			if name == definition.Name && version == definition.Version {
				return &definition, nil
			}
		}
	}

	definitionFiles, err := s.discoverDefinitions(version)
	if err != nil {
		return nil, err
	}

	for _, path := range definitionFiles {
		if name == s.definitionName(path) {
			data, err := s.readDefinitionData(path)
			if err != nil {
				return nil, err
			}
			definition := NewDefinitionData(name, version, data)
			s.definitions = append(s.definitions, *definition)
			return definition, err
		}
	}
	return nil, nil
}

func (s *DefinitionFileStore) discoverDefinitions(version string) ([]string, error) {
	if s.files != nil {
		return s.files, nil
	}

	definitionsDirectory, err := s.definitionsPath(version)
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(definitionsDirectory)
	if err != nil && os.IsNotExist(err) && version != "" {
		return nil, fmt.Errorf("Could not find definition files for version '%s' in folder '%s'", version, definitionsDirectory)
	}
	if err != nil {
		return nil, fmt.Errorf("Error reading definition files from folder '%s': %w", definitionsDirectory, err)
	}
	definitionFiles := []string{}
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, "yaml") || strings.HasSuffix(filename, "yml") || strings.HasSuffix(filename, "json") {
			path := filepath.Join(definitionsDirectory, file.Name())
			definitionFiles = append(definitionFiles, path)
		}
	}
	s.files = definitionFiles
	return definitionFiles, nil
}

func (s DefinitionFileStore) definitionsPath(version string) (string, error) {
	if s.directory != "" {
		return s.directory, nil
	}
	homeDir, err := os.UserHomeDir()
	if err == nil {
		definitionsDirectory := filepath.Join(homeDir, ".uipath", "definitions")
		_, err := os.Stat(definitionsDirectory)
		if err == nil {
			return definitionsDirectory, nil
		}
	}

	currentDirectory, err := os.Executable()
	definitionsDirectory := filepath.Join(filepath.Dir(currentDirectory), DefinitionsDirectory, version)
	if err != nil {
		return "", fmt.Errorf("Error reading definition files from folder '%s': %w", definitionsDirectory, err)
	}
	return definitionsDirectory, nil
}

func (s DefinitionFileStore) definitionName(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func (s DefinitionFileStore) definitionNames(paths []string) []string {
	names := []string{}
	for _, path := range paths {
		names = append(names, s.definitionName(path))
	}
	return names
}

func (s DefinitionFileStore) readDefinitionData(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error reading definition file '%s': %w", path, err)
	}
	return data, nil
}

func NewDefinitionFileStore(directory string) *DefinitionFileStore {
	return &DefinitionFileStore{
		directory: directory,
	}
}

func NewDefinitionFileStoreWithData(data []DefinitionData) *DefinitionFileStore {
	return &DefinitionFileStore{
		definitions: data,
	}
}
