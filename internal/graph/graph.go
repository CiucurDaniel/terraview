package graph

import (
	"bytes"
	"fmt"
	"github.com/awalterschulze/gographviz"
	"os/exec"
	"path/filepath"
	"strings"
)

// ObtainGraph invokes "terraform graph" command in the specified directory
// and returns the graph data as a string.
func ObtainGraph(dirPath string) (string, error) {
	// Get the absolute path of the directory
	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for directory: %v", err)
	}

	// Execute "terraform graph" command in the specified directory
	cmd := exec.Command("terraform", "graph")
	cmd.Dir = absDirPath // Set the command's working directory

	// Create a buffer to store command output
	var out bytes.Buffer
	cmd.Stdout = &out

	// Run the command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error running 'terraform graph' command in directory %s: %v", absDirPath, err)
	}

	// Return the graph data as a string
	return out.String(), nil
}

// TODO: Needs fixing, currently it break the formatting because the added label is not correct
// AddImageLabel appends an image label to every node in the graph where the value
// is the name of the node plus the ".jpeg" suffix. If a node already has a label,
// the image label is appended while preserving the existing label.
func AddImageLabel(graph *gographviz.Graph) {
	for _, node := range graph.Nodes.Nodes {
		// Get the current label of the node
		label := node.Attrs["label"]

		// Remove existing quotes, if any
		label = strings.Trim(label, `"`)

		// Add ".jpeg" suffix to the current label
		imageLabel := fmt.Sprintf("%s.jpeg", label)

		// Check if the node already has a label
		if label == "" {
			// If no label exists, set the image label as the node's label
			//node.Attrs["label"] = fmt.Sprintf(`"%s"`, imageLabel)
			continue
		} else {
			// If a label exists, append the image label without modifying the existing label
			node.Attrs["label"] = fmt.Sprintf(`"%s"`, label)
			node.Attrs["imageLabel"] = fmt.Sprintf(`"%s"`, imageLabel)
		}
	}
}

// PrepareGraphForPrinting is a facade function for preparing the graph for printing.
// It obtains the graph data, adds image labels to nodes, and returns the modified graph.
func PrepareGraphForPrinting(dirPath string) (*gographviz.Graph, error) {

	graphData, err := ObtainGraph(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain graph data: %v", err)
	}

	graph, err := gographviz.Read([]byte(graphData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse graph data: %v", err)
	}

	// Add image labels to nodes
	AddImageLabel(graph)

	return graph, nil
}
