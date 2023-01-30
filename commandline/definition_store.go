package commandline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DefinitionStore struct {
	DefinitionDirectory string
	DefinitionFiles     []string
	Definitions         []DefinitionData
}

const DefinitionsDirectory = "definitions"

func (s *DefinitionStore) Names() ([]string, error) {
	definitionFiles, err := s.discoverDefinitions()
	if err != nil {
		return nil, err
	}
	return s.definitionNames(definitionFiles), nil
}

func (s *DefinitionStore) Read(name string) (*DefinitionData, error) {
	if s.Definitions != nil {
		for _, definition := range s.Definitions {
			if name == definition.Name {
				return &definition, nil
			}
		}
	}

	definitionFiles, err := s.discoverDefinitions()
	if err != nil {
		return nil, err
	}

	for _, path := range definitionFiles {
		if name == s.definitionName(path) {
			definition, err := s.readDefinition(path)
			if definition != nil {
				s.Definitions = append(s.Definitions, *definition)
			}
			return definition, err
		}
	}
	return nil, nil
}

func (s *DefinitionStore) discoverDefinitions() ([]string, error) {
	if s.DefinitionFiles != nil {
		return s.DefinitionFiles, nil
	}

	definitionsDirectory, err := s.definitionsPath()
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(definitionsDirectory)
	if err != nil {
		return nil, fmt.Errorf("Error reading definition files from folder '%s': %v", definitionsDirectory, err)
	}
	definitionFiles := []string{}
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, "yaml") || strings.HasSuffix(filename, "yml") || strings.HasSuffix(filename, "json") {
			path := filepath.Join(definitionsDirectory, file.Name())
			definitionFiles = append(definitionFiles, path)
		}
	}
	s.DefinitionFiles = definitionFiles
	return definitionFiles, nil
}

func (s DefinitionStore) definitionsPath() (string, error) {
	if s.DefinitionDirectory != "" {
		return s.DefinitionDirectory, nil
	}
	currentDirectory, err := os.Executable()
	definitionsDirectory := filepath.Join(filepath.Dir(currentDirectory), DefinitionsDirectory)
	if err != nil {
		return "", fmt.Errorf("Error reading definition files from folder '%s': %v", definitionsDirectory, err)
	}
	return definitionsDirectory, nil
}

func (s DefinitionStore) definitionName(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

func (s DefinitionStore) definitionNames(paths []string) []string {
	names := []string{}
	for _, path := range paths {
		names = append(names, s.definitionName(path))
	}
	return names
}

func (s DefinitionStore) readDefinition(path string) (*DefinitionData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error reading definition file '%s': %v", path, err)
	}
	name := s.definitionName(path)
	return NewDefinitionData(name, data), nil
}
