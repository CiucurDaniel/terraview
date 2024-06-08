package render

import (
	"bytes"
	"fmt"
	"github.com/awalterschulze/gographviz"
	"os/exec"
)

var DPI = 96

// SaveGraphAsJPEG saves the given graph as a JPEG image with the specified DPI.
func SaveGraphAsJPEG(graph *gographviz.Graph, filePath string) error {
	// Render the graph to DOT format
	dot := graph.String()

	// Convert DOT content to JPEG using Graphviz command-line tool with specified DPI
	cmd := exec.Command("dot", "-Tjpg", fmt.Sprintf("-Gdpi=%d", DPI), "-o", filePath)
	cmd.Stdin = bytes.NewBufferString(dot) // Pass the DOT content as standard input

	// Capture standard output and standard error
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error converting DOT to JPEG: %v, output: %s", err, string(output))
	}

	return nil
}
