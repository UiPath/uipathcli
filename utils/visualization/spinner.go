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
	logger         log.Logger
	prefix         string
	renderedLength int
	cancel         chan struct{}
}

const spinnerCharacterSequence = "|/-\\"

func (s *Spinner) start() {
	for {
		for _, c := range spinnerCharacterSequence {
			select {
			case <-s.cancel:
				clear := strings.Repeat(" ", s.renderedLength)
				s.logger.LogError("\r" + clear + "\r")
				return
			default:
				s.logger.LogError(fmt.Sprintf("\r%s%c", s.prefix, c))
				s.renderedLength = len(s.prefix) + 1
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (s *Spinner) Close() {
	s.cancel <- struct{}{}
	close(s.cancel)
}

func NewSpinner(logger log.Logger, prefix string) *Spinner {
	spinner := Spinner{
		logger:         logger,
		prefix:         prefix,
		renderedLength: 0,
		cancel:         make(chan struct{}),
	}
	go spinner.start()
	return &spinner
}
