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

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/utils/directories"
	"github.com/UiPath/uipathcli/utils/network"
	"github.com/UiPath/uipathcli/utils/resiliency"
	"github.com/UiPath/uipathcli/utils/visualization"
)

const localPluginsDirectoryVarName = "UIPATH_PLUGINS_PATH"
const pluginsDirectoryPermissions = 0700

type ExternalPluginManager struct {
	Logger log.Logger
}

func (m ExternalPluginManager) Offline() error {
	err := m.offline(UipCliCrossPlatform)
	if err != nil {
		return err
	}
	err = m.offline(UipCliWindows)
	if err != nil {
		return err
	}
	return m.offline(DotNet8)
}

func (m ExternalPluginManager) offline(name string) error {
	plugin, ok := AvailablePlugins[name]
	if !ok {
		panic(fmt.Sprintf("Could not find external plugin: %s", name))
	}
	directory := m.getLocalPluginsDirectory()
	_ = os.MkdirAll(directory, pluginsDirectoryPermissions)
	path := filepath.Join(directory, plugin.ArchiveName)

	progressBar := visualization.NewProgressBar(m.Logger)
	defer progressBar.Remove()
	progressBar.UpdatePercentage("downloading...", 0)
	return m.download(name, plugin.Url, path, progressBar)
}

func (m ExternalPluginManager) Get(name string) (string, error) {
	plugin, ok := AvailablePlugins[name]
	if !ok {
		panic(fmt.Sprintf("Could not find external plugin: %s", name))
	}

	path := ""
	err := resiliency.Retry(func(attempt int) error {
		var err error
		path, err = m.getPlugin(name, plugin)
		if err != nil {
			return resiliency.Retryable(err)
		}
		return nil
	})
	return path, err
}

func (m ExternalPluginManager) getLocalPluginsDirectory() string {
	localPluginsDirectory := os.Getenv(localPluginsDirectoryVarName)
	if localPluginsDirectory != "" {
		return localPluginsDirectory
	}
	executable, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Join(filepath.Dir(executable), "plugins")
}

func (m ExternalPluginManager) findLocalTool(plugin ExternalPluginDefinition) string {
	directory := m.getLocalPluginsDirectory()
	path := filepath.Join(directory, plugin.ArchiveName)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

func (m ExternalPluginManager) getPlugin(name string, plugin ExternalPluginDefinition) (string, error) {
	pluginsDirectory, err := m.pluginsDirectory(name, plugin.Url)
	if err != nil {
		return "", fmt.Errorf("Could not download %s: %v", name, err)
	}
	path := filepath.Join(pluginsDirectory, plugin.Executable)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	tmpPluginsDirectory := m.createTmpFolder(pluginsDirectory, pluginsDirectoryPermissions)

	archivePath := m.findLocalTool(plugin)

	nextStep := "extracting... "
	if archivePath == "" {
		nextStep = "downloading..."
	}
	progressBar := visualization.NewProgressBar(m.Logger)
	defer progressBar.Remove()
	progressBar.UpdatePercentage(nextStep, 0)

	if archivePath == "" {
		archivePath = filepath.Join(tmpPluginsDirectory, plugin.ArchiveName)
		defer os.Remove(archivePath)
		err = m.download(name, plugin.Url, archivePath, progressBar)
		if err != nil {
			return "", err
		}
	}

	archive := newArchive(plugin.ArchiveType)
	err = archive.Extract(archivePath, tmpPluginsDirectory, pluginsDirectoryPermissions)
	if err != nil {
		return "", fmt.Errorf("Could not extract %s archive: %v", name, err)
	}
	err = m.rename(tmpPluginsDirectory, pluginsDirectory)
	if err != nil {
		return "", fmt.Errorf("Could not install %s: %v", name, err)
	}
	return path, nil
}

func (m ExternalPluginManager) rename(source string, target string) error {
	return resiliency.RetryN(10, func(attempt int) error {
		err := os.Rename(source, target)
		if err != nil {
			return resiliency.Retryable(err)
		}
		return nil
	})
}

func (m ExternalPluginManager) download(name string, url string, destination string, progressBar *visualization.ProgressBar) error {
	out, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("Could not download %s: %v", name, err)
	}
	defer out.Close()

	request := network.NewHttpGetRequest(url, nil, http.Header{})
	clientSettings := network.NewHttpClientSettings(false, "", 0, 1, false)
	client := network.NewHttpClient(nil, *clientSettings)
	response, err := client.Send(request)
	if err != nil {
		return fmt.Errorf("Could not download %s: %v", name, err)
	}
	defer response.Body.Close()
	downloadReader := m.progressReader("downloading...", "installing... ", response.Body, response.ContentLength, progressBar)
	_, err = io.Copy(out, downloadReader)
	if err != nil {
		return fmt.Errorf("Could not download %s: %v", name, err)
	}
	return nil
}

func (m ExternalPluginManager) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *visualization.ProgressBar) io.Reader {
	return visualization.NewProgressReader(reader, func(progress visualization.Progress) {
		displayText := text
		if progress.Completed {
			displayText = completedText
		}
		progressBar.UpdateProgress(displayText, progress.BytesRead, length, progress.BytesPerSecond)
	})
}

func (m ExternalPluginManager) pluginsDirectory(name string, url string) (string, error) {
	pluginDirectory, err := directories.Plugins()
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(url))
	subdirectory := fmt.Sprintf("%s-%x", name, hash)
	return filepath.Join(pluginDirectory, subdirectory), nil
}

func (m ExternalPluginManager) createTmpFolder(baseDirectory string, permissions os.FileMode) string {
	tmp := baseDirectory + "-" + m.randomFolderName()
	_ = os.MkdirAll(tmp, permissions)
	return tmp
}

func (m ExternalPluginManager) randomFolderName() string {
	value, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	return value.String()
}

func NewExternalPluginManager(logger log.Logger) *ExternalPluginManager {
	return &ExternalPluginManager{logger}
}
