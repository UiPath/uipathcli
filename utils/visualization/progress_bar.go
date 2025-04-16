package visualization

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/UiPath/uipathcli/log"
)

// The ProgressBar helps rendering a text-based progress indicator on the command-line.
// It uses the standard error output interface of the logger for writing progress.
type ProgressBar struct {
	logger         log.Logger
	renderedLength int
}

func (b *ProgressBar) UpdateSteps(text string, current int, total int) {
	b.logger.LogError("\r")
	percent := 0.1
	if total > 0 {
		percent = float64(current) / float64(total) * 100
	}
	steps := fmt.Sprintf(" (%d/%d)", current, total)
	length := b.renderTick(text, percent, steps)
	left := b.renderedLength - length
	if left > 0 {
		b.logger.LogError(strings.Repeat(" ", left))
	}
	b.renderedLength = length
}

func (b *ProgressBar) UpdatePercentage(text string, percent float64) {
	b.logger.LogError("\r")
	length := b.renderTick(text, percent, "")
	left := b.renderedLength - length
	if left > 0 {
		b.logger.LogError(strings.Repeat(" ", left))
	}
	b.renderedLength = length
}

func (b *ProgressBar) UpdateProgress(text string, current int64, total int64, bytesPerSecond int64) {
	b.logger.LogError("\r")
	length := b.renderProgress(text, current, total, bytesPerSecond)
	left := b.renderedLength - length
	if left > 0 {
		b.logger.LogError(strings.Repeat(" ", left))
	}
	b.renderedLength = length
}

func (b *ProgressBar) Remove() {
	if b.renderedLength > 0 {
		clearScreen := fmt.Sprintf("\r%s\r", strings.Repeat(" ", b.renderedLength))
		b.logger.LogError(clearScreen)
	}
}

func (b *ProgressBar) renderTick(text string, percent float64, info string) int {
	bar := b.createBar(percent)
	output := fmt.Sprintf("%s      |%s|%s",
		text,
		bar,
		info)
	b.logger.LogError(output)
	return utf8.RuneCountInString(output)
}

func (b *ProgressBar) renderProgress(text string, currentBytes int64, totalBytes int64, bytesPerSecond int64) int {
	percent := math.Min(float64(currentBytes)/float64(totalBytes)*100.0, 100.0)
	bar := b.createBar(percent)
	totalBytesFormatted, unit := b.formatBytes(totalBytes)
	currentBytesFormatted, unit := b.formatBytesInUnit(currentBytes, unit)
	bytesPerSecondFormatted, bytesPerSecondUnit := b.formatBytes(bytesPerSecond)
	output := fmt.Sprintf("%s %3d%% |%s| (%s/%s %s, %s %s/s)",
		text,
		int(percent),
		bar,
		currentBytesFormatted,
		totalBytesFormatted,
		unit,
		bytesPerSecondFormatted,
		bytesPerSecondUnit)
	b.logger.LogError(output)
	return utf8.RuneCountInString(output)
}

func (b *ProgressBar) createBar(percent float64) string {
	barCount := int(percent / 5.0)
	if barCount > 20 {
		barCount = 20
	}
	return strings.Repeat("█", barCount) + strings.Repeat(" ", 20-barCount)
}

func (b *ProgressBar) formatBytes(count int64) (string, string) {
	if count < 1000 {
		return b.formatBytesInUnit(count, "B")
	}
	if count < 1000*1000 {
		return b.formatBytesInUnit(count, "kB")
	}
	if count < 1000*1000*1000 {
		return b.formatBytesInUnit(count, "MB")
	}
	return b.formatBytesInUnit(count, "GB")
}

func (b *ProgressBar) formatBytesInUnit(count int64, unit string) (string, string) {
	if unit == "B" {
		return strconv.FormatInt(count, 10), unit
	}
	if unit == "kB" {
		return fmt.Sprintf("%0.1f", float64(count)/1_000), unit
	}
	if unit == "MB" {
		return fmt.Sprintf("%0.1f", float64(count)/1_000_000), unit
	}
	return fmt.Sprintf("%0.1f", float64(count)/1_000_000_000), unit
}

func NewProgressBar(logger log.Logger) *ProgressBar {
	return &ProgressBar{logger, 0}
}
