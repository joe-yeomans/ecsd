package views

import (
	"fmt"
	"strings"

	"github.com/joe-yeomans/ecsd/styles"

	"github.com/charmbracelet/lipgloss"
)

// Sparkline characters from low to high (Unicode block elements)
var sparklineChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// RenderSparkline renders a sparkline from a slice of values.
// Values are expected to be percentages (0-100).
// Width is the number of characters to render.
func RenderSparkline(values []float64, width int) string {
	if len(values) == 0 {
		return strings.Repeat("─", width)
	}

	// If we have more values than width, sample evenly
	if len(values) > width {
		values = sampleValues(values, width)
	}

	// Pad with empty space if we have fewer values
	padding := ""
	if len(values) < width {
		padding = strings.Repeat(" ", width-len(values))
	}

	var result strings.Builder
	for _, v := range values {
		idx := int(v / 100.0 * float64(len(sparklineChars)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparklineChars) {
			idx = len(sparklineChars) - 1
		}
		result.WriteRune(sparklineChars[idx])
	}

	return padding + result.String()
}

// SparklineWithValue renders a sparkline followed by the current value.
// Returns format: "▃▅▇█▆▄ 45%"
func SparklineWithValue(values []float64, width int, currentVal float64) string {
	// Reserve 5 chars for value display (e.g., " 100%")
	sparkWidth := width - 5
	if sparkWidth < 3 {
		sparkWidth = 3
	}

	spark := RenderSparkline(values, sparkWidth)
	valStr := fmt.Sprintf("%3.0f%%", currentVal)

	return spark + " " + valStr
}

// SparklineWithValueStyled renders a colored sparkline based on value.
func SparklineWithValueStyled(values []float64, width int, currentVal float64, s styles.Styles) string {
	content := SparklineWithValue(values, width, currentVal)

	// Color based on utilization level
	var style lipgloss.Style
	switch {
	case currentVal >= 90:
		style = lipgloss.NewStyle().Foreground(styles.ColorDanger)
	case currentVal >= 70:
		style = lipgloss.NewStyle().Foreground(styles.ColorWarning)
	default:
		style = lipgloss.NewStyle().Foreground(styles.ColorSuccess)
	}

	return style.Render(content)
}

// sampleValues evenly samples values to fit within the target width.
func sampleValues(values []float64, targetLen int) []float64 {
	if len(values) <= targetLen {
		return values
	}

	result := make([]float64, targetLen)
	step := float64(len(values)) / float64(targetLen)

	for i := 0; i < targetLen; i++ {
		idx := int(float64(i) * step)
		if idx >= len(values) {
			idx = len(values) - 1
		}
		result[i] = values[idx]
	}

	return result
}

// ExtractCPUValues extracts CPU utilization values from task metrics.
func ExtractCPUValues(metrics []float64) []float64 {
	return metrics
}

// GetLatestValue returns the most recent value from a slice, or 0 if empty.
func GetLatestValue(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return values[len(values)-1]
}
