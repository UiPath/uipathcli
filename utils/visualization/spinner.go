package visualization

import (
	"fmt"
	"strings"
	"time"

	"github.com/UiPath/uipathcli/log"
)

// The Spinner visualization shows a waiting indicator until the operation
// is finshed. Typically used when the time of the operation is unknown
// upfront and there should be a loading indicator until the operation finished.
type Spinner struct {
	logger log.Logger
	prefix string
	cancel chan struct{}
}

const spinnerCharacterSequence = "|/-\\"

func (s Spinner) start() {
	for {
		for _, c := range spinnerCharacterSequence {
			select {
			case <-s.cancel:
				clear := strings.Repeat(" ", len(s.prefix)+1)
				s.logger.LogError("\r" + clear + "\r")
				return
			default:
				s.logger.LogError(fmt.Sprintf("\r%s%c", s.prefix, c))
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (s Spinner) Close() error {
	s.cancel <- struct{}{}
	close(s.cancel)
	return nil
}

func NewSpinner(logger log.Logger, prefix string) *Spinner {
	spinner := Spinner{
		logger: logger,
		prefix: prefix,
		cancel: make(chan struct{}),
	}
	go spinner.start()
	return &spinner
}
