package plugin

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/utils/directories"
	"github.com/UiPath/uipathcli/utils/network"
	"github.com/UiPath/uipathcli/utils/resiliency"
	"github.com/UiPath/uipathcli/utils/visualization"
)

const directoryPermissions = 0700

type ModuleManager struct {
	Logger log.Logger
}

func (m ModuleManager) Offline() (string, error) {
	status1, err := m.offline(UipCliCrossPlatform)
	if err != nil {
		return m.overallStatus(status1), err
	}
	status2, err := m.offline(UipCliWindows)
	if err != nil {
		return m.overallStatus(status1, status2), err
	}
	status3, err := m.offline(DotNet8)
	return m.overallStatus(status1, status2, status3), err
}

func (m ModuleManager) overallStatus(status ...string) string {
	return "\n" + strings.Join(status, "\n")
}

func (m ModuleManager) succeededStatus(definition ModuleDefinition) string {
	return "[ succeeded ] " + definition.Name + ": successfully downloaded from " + definition.Url
}

func (m ModuleManager) failedStatus(definition ModuleDefinition, err error) string {
	return "[ failed    ] " + definition.Name + ": download failed with error " + err.Error()
}

func (m ModuleManager) skippedStatus(definition ModuleDefinition, path string) string {
	return "[ skipped   ] " + definition.Name + ": already present at " + path
}

func (m ModuleManager) offline(name string) (string, error) {
	definition := m.getDefinition(name)
	existing := m.findLocalModule(definition)
	if existing != "" {
		return m.skippedStatus(definition, existing), nil
	}

	directory, err := directories.OfflineModules()
	if err != nil {
		return m.failedStatus(definition, err), err
	}
	path := filepath.Join(directory, definition.ArchiveName)

	progressBar := visualization.NewProgressBar(m.Logger)
	defer progressBar.Remove()
	progressBar.UpdatePercentage("downloading...", 0)
	err = m.download(definition, path, progressBar)
	if err != nil {
		return m.failedStatus(definition, err), err
	}
	return m.succeededStatus(definition), nil
}

func (m ModuleManager) Get(name string) (string, error) {
	definition := m.getDefinition(name)
	path := ""
	err := resiliency.Retry(func(attempt int) error {
		var err error
		path, err = m.getModule(definition)
		if err != nil {
			return resiliency.Retryable(err)
		}
		return nil
	})
	return path, err
}

func (m ModuleManager) getDefinition(name string) ModuleDefinition {
	definition, ok := AvailableModules[name]
	if !ok {
		panic("Could not find module: " + name)
	}
	return definition
}

func (m ModuleManager) findLocalModule(module ModuleDefinition) string {
	directory, err := directories.OfflineModules()
	if err != nil {
		return ""
	}
	path := filepath.Join(directory, module.ArchiveName)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

func (m ModuleManager) getModule(definition ModuleDefinition) (string, error) {
	moduleDirectory, err := m.moduleDirectory(definition)
	if err != nil {
		return "", fmt.Errorf("Could not download %s: %w", definition.Name, err)
	}
	path := filepath.Join(moduleDirectory, definition.Executable)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	tmpModuleDirectory := m.createTmpFolder(moduleDirectory, directoryPermissions)

	archivePath := m.findLocalModule(definition)

	nextStep := "extracting... "
	if archivePath == "" {
		nextStep = "downloading..."
	}
	progressBar := visualization.NewProgressBar(m.Logger)
	defer progressBar.Remove()
	progressBar.UpdatePercentage(nextStep, 0)

	if archivePath == "" {
		archivePath = filepath.Join(tmpModuleDirectory, definition.ArchiveName)
		defer func() { _ = os.Remove(archivePath) }()
		err = m.download(definition, archivePath, progressBar)
		if err != nil {
			return "", err
		}
	}

	archive := newArchive(definition.ArchiveType)
	err = archive.Extract(archivePath, tmpModuleDirectory, directoryPermissions)
	if err != nil {
		return "", fmt.Errorf("Could not extract %s archive: %w", definition.Name, err)
	}
	err = m.rename(tmpModuleDirectory, moduleDirectory)
	if err != nil {
		return "", fmt.Errorf("Could not install %s: %w", definition.Name, err)
	}
	return path, nil
}

func (m ModuleManager) rename(source string, target string) error {
	return resiliency.RetryN(10, func(attempt int) error {
		err := os.Rename(source, target)
		if err != nil {
			return resiliency.Retryable(err)
		}
		return nil
	})
}

func (m ModuleManager) download(definition ModuleDefinition, destination string, progressBar *visualization.ProgressBar) error {
	out, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("Could not download %s: %w", definition.Name, err)
	}
	defer func() { _ = out.Close() }()

	request := network.NewHttpGetRequest(definition.Url, nil, http.Header{})
	clientSettings := network.NewHttpClientSettings(false, "", map[string]string{}, 0, 1, false)
	client := network.NewHttpClient(nil, *clientSettings)
	response, err := client.Send(request)
	if err != nil {
		return fmt.Errorf("Could not download %s: %w", definition.Name, err)
	}
	defer func() { _ = response.Body.Close() }()
	downloadReader := m.progressReader("downloading...", "installing... ", response.Body, response.ContentLength, progressBar)
	_, err = io.Copy(out, downloadReader)
	if err != nil {
		return fmt.Errorf("Could not download %s: %w", definition.Name, err)
	}
	return nil
}

func (m ModuleManager) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *visualization.ProgressBar) io.Reader {
	return visualization.NewProgressReader(reader, func(progress visualization.Progress) {
		displayText := text
		if progress.Completed {
			displayText = completedText
		}
		progressBar.UpdateProgress(displayText, progress.BytesRead, length, progress.BytesPerSecond)
	})
}

func (m ModuleManager) moduleDirectory(definition ModuleDefinition) (string, error) {
	modulesDirectory, err := directories.Modules()
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(definition.Url))
	subdirectory := fmt.Sprintf("%s-%x", definition.Name, hash)
	return filepath.Join(modulesDirectory, subdirectory), nil
}

func (m ModuleManager) createTmpFolder(baseDirectory string, permissions os.FileMode) string {
	tmp := baseDirectory + "-" + m.randomFolderName()
	_ = os.MkdirAll(tmp, permissions)
	return tmp
}

func (m ModuleManager) randomFolderName() string {
	value, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	return value.String()
}

func NewModuleManager(logger log.Logger) *ModuleManager {
	return &ModuleManager{logger}
}
