package graph

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/awalterschulze/gographviz"
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
func ObtainGraph(dirPath string) (*gographviz.Graph, error) {
	// Get the absolute path of the directory
	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for directory: %v", err)
	}

	// Execute "terraform graph" command in the specified directory
	cmd := exec.Command("terraform", "graph")
	cmd.Dir = absDirPath // Set the command's working directory

	// Create a buffer to store command output
	var out bytes.Buffer
	cmd.Stdout = &out

	// Run the command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error running 'terraform graph' command in directory %s: %v", absDirPath, err)
	}

	// Get the graph data as a string
	graphData := out.String()

	// Parse string into AST
	graphAst, err := gographviz.ParseString(graphData)
	if err != nil {
		return nil, fmt.Errorf("error parsing Terraform graph data: %v", err)
	}

	// Create a new graph object
	graph := gographviz.NewGraph()

	// Analyze and populate the graph object
	err = gographviz.Analyse(graphAst, graph)
	if err != nil {
		return nil, fmt.Errorf("error analyzing Terraform graph data: %v", err)
	}

	// Return the parsed graph
	return graph, nil
}

// SetGraphGlobalImagePath sets the global image path for the graph.
func SetGraphGlobalImagePath(graph *gographviz.Graph, path string) {
	graph.Attrs.Add("imagepath", fmt.Sprintf(`"%s"`, path))
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

	graph, err := ObtainGraph(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain graph data: %v", err)
	}

	//SetGraphGlobalImagePath(graph, GlobalImagePath)
	ConvertNodesToSubgraphs(graph)
	//AddImageLabel(graph)
	//PositionLabelTo(graph, LABEL_LOCATION)

	return graph, nil
}

// Function to convert nodes to subgraphs
func ConvertNodesToSubgraphs(graph *gographviz.Graph) error {

	var groupingLabels = []string{"azurerm_virtual_network", "azurerm_resource_group"}
	var counter = 0

	for _, node := range graph.Nodes.Nodes {
		label := node.Attrs["label"]
		label = strings.Trim(label, `"`)
		if IsResourceNode(label) {
			parts := strings.Split(label, ".")
			foundGroupingResource := contains(groupingLabels, parts[0])
			if foundGroupingResource {
				fmt.Println("Found grouping resource: " + parts[0])
				counter++

				// do not add only the resource type, add the full name, so later we can add the node inside this subGraph
				err := graph.AddSubGraph(`"root"`, "cluster_"+label, nil)
				if err != nil {
					fmt.Println("DEBUG: Got an error trying to add subgraph")
				}

			}
		}
	}

	graph.AddNode("cluster_azurerm_resource_group.rg", "A", nil)
	fmt.Println("Subgraphs are:")
	for _, s := range graph.SubGraphs.Sorted() {
		fmt.Println(s.Name)
	}
	fmt.Println("------------------------")

	fmt.Println("About to print the graph")
	fmt.Println(graph.String())
	fmt.Println("------------------------")

	for _, edge := range graph.Edges.Sorted() {
		fmt.Println("Edge: " + edge.Src + "--->" + edge.Dst)
	}
	return nil
}

func contains(arr []string, str string) bool {
	for _, item := range arr {
		if item == str {
			return true
		}
	}
	return false
}
