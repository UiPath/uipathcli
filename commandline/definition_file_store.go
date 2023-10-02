package commandline

import (
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// DefinitionFileStore discovers the definition files from disk searching for
// the definitions/ folder and returns the data for a particular definition file.
type DefinitionFileStore struct {
	directory   string
	embedded    embed.FS
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

	for _, fileName := range definitionFiles {
		if name == s.definitionName(fileName) {
			data, err := s.readDefinitionData(version, fileName)
			if err != nil {
				return nil, err
			}
			definition := NewDefinitionData(name, version, data)
			return definition, err
		}
	}
	return nil, nil
}

func (s *DefinitionFileStore) discoverDefinitions(version string) ([]string, error) {
	if s.files != nil {
		return s.files, nil
	}

	definitionFiles := map[string]string{}

	embeddedFiles := s.discoverDefinitionsEmbedded()
	for _, fileName := range embeddedFiles {
		definitionFiles[fileName] = fileName
	}
	directoryFiles := s.discoverDefinitionsDirectory(version)
	for _, fileName := range directoryFiles {
		definitionFiles[fileName] = fileName
	}

	if len(definitionFiles) == 0 {
		return nil, fmt.Errorf("Could not find definition files in folder '%s'", s.definitionsPath(version))
	}

	result := []string{}
	for _, path := range definitionFiles {
		result = append(result, path)
	}
	sort.Strings(result)
	s.files = result
	return result, nil
}

func (s DefinitionFileStore) discoverDefinitionsEmbedded() []string {
	definitionFiles := []string{}
	embeddedDir, err := s.embedded.ReadDir(DefinitionsDirectory)
	if err == nil {
		for _, file := range embeddedDir {
			definitionFiles = append(definitionFiles, file.Name())
		}
	}
	return definitionFiles
}

func (s DefinitionFileStore) discoverDefinitionsDirectory(version string) []string {
	definitionFiles := []string{}
	definitionsDirectory := s.definitionsPath(version)
	files, err := os.ReadDir(definitionsDirectory)
	if err == nil {
		for _, file := range files {
			filename := file.Name()
			if strings.HasSuffix(filename, "yaml") || strings.HasSuffix(filename, "yml") || strings.HasSuffix(filename, "json") {
				definitionFiles = append(definitionFiles, filename)
			}
		}
	}
	return definitionFiles
}

func (s DefinitionFileStore) definitionsPath(version string) string {
	if s.directory != "" {
		return s.directory
	}
	currentDirectory, err := os.Executable()
	if err != nil {
		return filepath.Join(DefinitionsDirectory, version)
	}
	return filepath.Join(filepath.Dir(currentDirectory), DefinitionsDirectory, version)
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

func (s DefinitionFileStore) readDefinitionData(version string, fileName string) ([]byte, error) {
	definitionsFilePath := filepath.Join(s.definitionsPath(version), fileName)
	data, err := os.ReadFile(definitionsFilePath)
	if err != nil {
		embeddedFilePath := path.Join(DefinitionsDirectory, fileName)
		data, err = s.embedded.ReadFile(embeddedFilePath)
	}
	if err != nil {
		return nil, fmt.Errorf("Error reading definition file '%s': %w", fileName, err)
	}
	return data, nil
}

func NewDefinitionFileStore(directory string, embedded embed.FS) *DefinitionFileStore {
	return &DefinitionFileStore{
		directory: directory,
		embedded:  embedded,
	}
}

func NewDefinitionFileStoreWithData(data []DefinitionData) *DefinitionFileStore {
	return &DefinitionFileStore{
		definitions: data,
	}
}
