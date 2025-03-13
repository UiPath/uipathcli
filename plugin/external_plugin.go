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

const pluginDirectoryPermissions = 0700

type ExternalPlugin struct {
	Logger log.Logger
}

func (p ExternalPlugin) GetTool(name string, url string, archiveType ArchiveType, executable string) (string, error) {
	path := ""
	err := resiliency.Retry(func(attempt int) error {
		var err error
		path, err = p.getTool(name, url, archiveType, executable)
		if err != nil {
			return resiliency.Retryable(err)
		}
		return nil
	})
	return path, err
}

func (p ExternalPlugin) getTool(name string, url string, archiveType ArchiveType, executable string) (string, error) {
	pluginDirectory, err := p.pluginDirectory(name, url)
	if err != nil {
		return "", fmt.Errorf("Could not download %s: %v", name, err)
	}
	path := filepath.Join(pluginDirectory, executable)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	tmpPluginDirectory := pluginDirectory + "-" + p.randomFolderName()
	_ = os.MkdirAll(tmpPluginDirectory, pluginDirectoryPermissions)

	progressBar := visualization.NewProgressBar(p.Logger)
	defer progressBar.Remove()
	progressBar.UpdatePercentage("downloading...", 0)
	archivePath := filepath.Join(tmpPluginDirectory, name)
	err = p.download(name, url, archivePath, progressBar)
	if err != nil {
		return "", err
	}
	archive := newArchive(archiveType)
	err = archive.Extract(archivePath, tmpPluginDirectory, pluginDirectoryPermissions)
	if err != nil {
		return "", fmt.Errorf("Could not extract %s archive: %v", name, err)
	}
	err = os.Remove(archivePath)
	if err != nil {
		return "", fmt.Errorf("Could not remove %s archive: %v", name, err)
	}
	err = p.rename(tmpPluginDirectory, pluginDirectory)
	if err != nil {
		return "", fmt.Errorf("Could not install %s: %v", name, err)
	}
	return path, nil
}

func (p ExternalPlugin) rename(source string, target string) error {
	return resiliency.RetryN(10, func(attempt int) error {
		err := os.Rename(source, target)
		if err != nil {
			return resiliency.Retryable(err)
		}
		return nil
	})
}

func (p ExternalPlugin) download(name string, url string, destination string, progressBar *visualization.ProgressBar) error {
	out, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("Could not download %s: %v", name, err)
	}
	defer out.Close()

	request := network.NewHttpGetRequest(url, http.Header{})
	clientSettings := network.NewHttpClientSettings(false, "", 0, 1, false)
	client := network.NewHttpClient(nil, *clientSettings)
	response, err := client.Send(request)
	if err != nil {
		return fmt.Errorf("Could not download %s: %v", name, err)
	}
	defer response.Body.Close()
	downloadReader := p.progressReader("downloading...", "installing... ", response.Body, response.ContentLength, progressBar)
	_, err = io.Copy(out, downloadReader)
	if err != nil {
		return fmt.Errorf("Could not download %s: %v", name, err)
	}
	return nil
}

func (p ExternalPlugin) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *visualization.ProgressBar) io.Reader {
	return visualization.NewProgressReader(reader, func(progress visualization.Progress) {
		displayText := text
		if progress.Completed {
			displayText = completedText
		}
		progressBar.UpdateProgress(displayText, progress.BytesRead, length, progress.BytesPerSecond)
	})
}

func (p ExternalPlugin) pluginDirectory(name string, url string) (string, error) {
	pluginDirectory, err := directories.Plugin()
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(url))
	subdirectory := fmt.Sprintf("%s-%x", name, hash)
	return filepath.Join(pluginDirectory, subdirectory), nil
}

func (p ExternalPlugin) randomFolderName() string {
	value, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	return value.String()
}

func NewExternalPlugin(logger log.Logger) *ExternalPlugin {
	return &ExternalPlugin{logger}
}
