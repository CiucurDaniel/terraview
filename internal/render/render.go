package render

import (
	"bytes"
	"fmt"
	"github.com/awalterschulze/gographviz"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var DPI = 96

// SaveGraphAs saves the given graph in the specified format.
func SaveGraphAs(graph *gographviz.Graph, baseName string, format string) error {
	// Render the graph to DOT format
	dot := graph.String()

	// Handle the special case where format is "DOT"
	if format == "dot" {
		fmt.Println(dot)
		return nil
	}

	// Ensure the format is supported by Graphviz
	supportedFormats := map[string]bool{
		"png": true,
		"jpg": true,
		"svg": true,
		"pdf": true,
		"dot": true,
	}
	if !supportedFormats[format] {
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Generate the filename with a timestamp
	timestamp := time.Now().Format("20060102_150405")
	filePath := fmt.Sprintf("%s_%s.%s", baseName, timestamp, format)

	// Ensure the output directory exists
	outputDir := filepath.Dir(filePath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %v", err)
	}

	// Convert DOT content to the specified format using Graphviz command-line tool
	cmd := exec.Command("dot", fmt.Sprintf("-T%s", format), fmt.Sprintf("-Gdpi=%d", DPI), "-o", filePath)
	cmd.Stdin = bytes.NewBufferString(dot) // Pass the DOT content as standard input

	// Capture standard output and standard error
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error converting DOT to %s: %v, output: %s", format, err, string(output))
	}

	return nil
}
