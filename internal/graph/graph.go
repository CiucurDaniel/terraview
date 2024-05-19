package graph

import (
	"bytes"
	"fmt"
	"github.com/CiucurDaniel/terraview/internal/config"
	"github.com/awalterschulze/gographviz"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	LABEL_LOCATION = "b"

	// GlobalImagePath is the path where the images are stored.
	GlobalImagePath = "internal/icons/azurerm"
)

// KnownProviders is a constant array containing known provider prefixes
var KnownProviders = []string{"azurerm", "aws", "gcp"}

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

// IsResourceNode checks if the label represents a resource node based on the known provider prefixes.
// It returns true if the label starts with any of the known prefixes, otherwise false.
func IsResourceNode(label string) bool {
	for _, provider := range KnownProviders {
		if strings.HasPrefix(label, provider+"_") {
			return true
		}
	}
	return false
}

// SetGraphGlobalImagePath sets the global image path for the graph.
func SetGraphGlobalImagePath(graph *gographviz.Graph, path string) {
	graph.Attrs.Add("imagepath", fmt.Sprintf(`"%s"`, path))
}

// TODO: Needs fixing, currently it break the formatting because the added label is not correct
// AddImageLabel appends an image label to every resource node in the graph where the value
// is constructed based on the label of the node.
func AddImageLabel(graph *gographviz.Graph) {
	for _, node := range graph.Nodes.Nodes {
		// Get the current label of the node
		label := node.Attrs["label"]

		// Remove existing quotes, if any
		label = strings.Trim(label, `"`)

		// Check if the node represents a resource
		if IsResourceNode(label) {
			// Split the label by underscore to extract the relevant part
			parts := strings.Split(label, ".")

			// If the label is empty or doesn't contain a dot, skip the node
			if len(parts) < 2 {
				continue
			}

			// Construct the image label path
			imageLabel := filepath.Join(parts[0] + ".png")

			// Set the image label attribute
			node.Attrs["image"] = fmt.Sprintf(`"%s"`, imageLabel)

			// Set shape to none so the icon is not surrounded by a box
			node.Attrs["shape"] = `"none"`
		}
	}
}

// PositionLabelTo sets the labelloc attribute of every node in the graph to the specified position.
// Valid positions are "t" (top), "c" (center), and "b" (bottom).
func PositionLabelTo(graph *gographviz.Graph, position string) {
	// Check if the specified position is valid
	if position != "t" && position != "c" && position != "b" {
		return // If not valid, do nothing
	}

	// Iterate over every node in the graph
	for _, node := range graph.Nodes.Nodes {
		// Set the labelloc attribute of the node to the specified position
		node.Attrs["labelloc"] = fmt.Sprintf(`"%s"`, position)
	}
}

// SaveGraphAsJPEG saves the given graph as a JPEG image.
func SaveGraphAsJPEG(graph *gographviz.Graph, filePath string) error {
	// Render the graph to DOT format
	dot := graph.String()

	// Create a temporary directory for the DOT file and the output image
	tempDir, err := ioutil.TempDir("", "graphviz")
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up temporary directory

	// Write the DOT string to a temporary file
	tempFilePath := filepath.Join(tempDir, "graph.dot")
	err = ioutil.WriteFile(tempFilePath, []byte(dot), 0644)
	if err != nil {
		return fmt.Errorf("error writing DOT to temporary file: %v", err)
	}

	// Ensure the output file path is absolute
	outputFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("error getting absolute file path: %v", err)
	}

	// Convert DOT file to JPEG using Graphviz command-line tool
	cmd := exec.Command("dot", "-Tjpg", tempFilePath, "-o", outputFilePath)
	cmd.Dir = tempDir // Set the working directory to the temporary directory

	// Capture standard output and standard error
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error converting DOT to JPEG: %v, output: %s", err, string(output))
	}

	return nil
}

// TODO: Create func which puts consecutive identical resources on same rank
// similar to what was done here: https://stackoverflow.com/questions/58832678/how-to-separate-picture-and-label-of-a-node-with-graphviz

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

	SetGraphGlobalImagePath(graph, GlobalImagePath)
	AddImageLabel(graph)
	PositionLabelTo(graph, LABEL_LOCATION)

	return graph, nil
}

// ConvertNodesToSubgraphs converts specified nodes into subgraphs.
func ConvertNodesToSubgraphs(graph *gographviz.Graph) {
	cfg := config.GetConfig()
	if cfg == nil {
		fmt.Println("No configuration loaded.")
		return
	}

	for _, node := range graph.Nodes.Nodes {
		label := node.Attrs["label"]
		label = strings.Trim(label, `"`)
		parts := strings.Split(label, ".")

		if len(parts) < 2 {
			continue
		}

		nodeType := parts[0]

		if contains(cfg.GroupingElements, nodeType) {
			// Create a subgraph for this node
			subgraphName := fmt.Sprintf("cluster_%s", node.Name)
			graph.AddSubGraph("G", subgraphName, nil)

			// Customize the subgraph's appearance
			for _, subgraph := range graph.SubGraphs.SubGraphs {
				if subgraph.Name == subgraphName {
					subgraph.Attrs.Add("style", "rounded")
					subgraph.Attrs.Add("label", fmt.Sprintf("\"%s\"", label))
					break
				}
			}

			// Add the node to the subgraph
			for _, subgraph := range graph.SubGraphs.Sorted() {
				if subgraph.Name == subgraphName {
					// Update subgraph label
					subgraph.Attrs["label"] = fmt.Sprintf("\"%s\"", label)
					// Customize the subgraph's appearance
					subgraph.Attrs.Add("style", "filled")
					subgraph.Attrs.Add("fillcolor", "lightgray")

					// Construct the DOT representation of the node
					nodeDOT := fmt.Sprintf(`"%s" [label="%s"]`, node.Name, node.Attrs["label"])

					// Append the node DOT representation to the subgraph's DOT representation
					subgraph.Attrs["statement"] += "\n" + nodeDOT

					break
				}
			}

			// Update edges to refer to the subgraph
			for _, edge := range graph.Edges.Edges {
				if edge.Src == node.Name {
					edge.Src = subgraphName
				}
				if edge.Dst == node.Name {
					edge.Dst = subgraphName
				}
			}
		}
	}
}

func contains(slice []string, item string) bool {
	for _, elem := range slice {
		if elem == item {
			return true
		}
	}
	return false
}
