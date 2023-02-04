package utils

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/UiPath/uipathcli/log"
)

type ProgressBar struct {
	Logger         log.Logger
	renderedLength int
}

func (b *ProgressBar) Update(text string, current int64, total int64, bytesPerSecond int64) {
	b.Logger.LogError("\r")
	length := b.render(text, current, total, bytesPerSecond)
	left := b.renderedLength - length
	if left > 0 {
		b.Logger.LogError(strings.Repeat(" ", left))
	}
	b.renderedLength = length
}

func (b *ProgressBar) Remove() {
	if b.renderedLength > 0 {
		clear := fmt.Sprintf("\r%s\r", strings.Repeat(" ", b.renderedLength))
		b.Logger.LogError(clear)
	}
}

func (b ProgressBar) render(text string, currentBytes int64, totalBytes int64, bytesPerSecond int64) int {
	percent := math.Min(float64(currentBytes)/float64(totalBytes)*100.0, 100.0)
	barCount := int(percent / 5.0)
	bar := strings.Repeat("█", barCount) + strings.Repeat(" ", 20-barCount)
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
	b.Logger.LogError(output)
	return utf8.RuneCountInString(output)
}

func (b ProgressBar) formatBytes(count int64) (string, string) {
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

func (b ProgressBar) formatBytesInUnit(count int64, unit string) (string, string) {
	if unit == "B" {
		return fmt.Sprintf("%d", count), unit
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
