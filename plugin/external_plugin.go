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
	"github.com/UiPath/uipathcli/utils"
)

const pluginDirectoryPermissions = 0700

type ExternalPlugin struct {
	Logger log.Logger
}

func (p ExternalPlugin) GetTool(name string, url string, executable string) (string, error) {
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

	progressBar := utils.NewProgressBar(p.Logger)
	defer progressBar.Remove()
	progressBar.UpdatePercentage("downloading...", 0)
	zipArchivePath := filepath.Join(tmpPluginDirectory, name)
	err = p.download(name, url, zipArchivePath, progressBar)
	if err != nil {
		return "", err
	}
	err = newZipArchive().Extract(zipArchivePath, tmpPluginDirectory, pluginDirectoryPermissions)
	if err != nil {
		return "", fmt.Errorf("Could not extract %s archive: %v", name, err)
	}
	os.Remove(zipArchivePath)
	err = os.Rename(tmpPluginDirectory, pluginDirectory)
	if err != nil {
		return "", fmt.Errorf("Could not install %s: %v", name, err)
	}
	return path, nil
}

func (p ExternalPlugin) download(name string, url string, destination string, progressBar *utils.ProgressBar) error {
	out, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("Could not download %s: %v", name, err)
	}
	defer out.Close()

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("Could not download %s: %v", name, err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("Could not download %s: %v", name, err)
	}
	downloadReader := p.progressReader("downloading...", "installing... ", response.Body, response.ContentLength, progressBar)
	_, err = io.Copy(out, downloadReader)
	if err != nil {
		return fmt.Errorf("Could not download %s: %v", name, err)
	}
	return nil
}

func (p ExternalPlugin) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *utils.ProgressBar) io.Reader {
	progressReader := utils.NewProgressReader(reader, func(progress utils.Progress) {
		displayText := text
		if progress.Completed {
			displayText = completedText
		}
		progressBar.UpdateProgress(displayText, progress.BytesRead, length, progress.BytesPerSecond)
	})
	return progressReader
}

func (p ExternalPlugin) pluginDirectory(name string, url string) (string, error) {
	pluginDirectory, err := utils.Directories{}.Plugin()
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
