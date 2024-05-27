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
	NODE_LABEL_LOCATION  = "b"
	GRAPH_LABEL_LOCATION = "b"

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
		// WARNING: this will always be thrown if user didn't run terraform init prior to invoking our code
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

// This will set:
// - compound = true
// - newrank = true
// - rankdir = "TD
// - (optional) call SetGraphGlobalImagePath
func SetGraphAttrs(graph *gographviz.Graph) {
	graph.Attrs["compound"] = `"true"`
	graph.Attrs["rankdir"] = `"BT"`
	graph.Attrs["newrank"] = `"true"`
	graph.Attrs["nodesep"] = `"4"`
	graph.Attrs["ranksep"] = `"1"`

	// TODO: For each subgraph set labelloc="b";
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

// PositionNodeLabelTo sets the labelloc attribute of every node in the graph to the specified position.
// Valid positions are "t" (top), "c" (center), and "b" (bottom).
func PositionNodeLabelTo(graph *gographviz.Graph, position string) {
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

// PositionGraphLabelTo sets the labelloc attribute of every graph/subgraph in the graph to the specified position.
// Valid positions are "t" (top), "c" (center), and "b" (bottom).
func PositionGraphLabelTo(graph *gographviz.Graph, position string) {
	// Check if the specified position is valid
	if position != "t" && position != "c" && position != "b" {
		return // If not valid, do nothing
	}

	// Iterate over every node in the graph
	for _, graph := range graph.SubGraphs.SubGraphs {
		// Set the labelloc attribute of the node to the specified position
		graph.Attrs["labelloc"] = fmt.Sprintf(`"%s"`, position)
	}
}

// AddMarginToNodes sets the margin attribute of every node in the graph to the specified value.
func AddMarginToNodes(graph *gographviz.Graph, value float32) {
	// Iterate over every node in the graph
	for _, node := range graph.Nodes.Nodes {
		// Set the labelloc attribute of the node to the specified position
		node.Attrs["margin"] = fmt.Sprintf(`"%.2f"`, value)
	}
}

// AddMarginToNodes sets the margin attribute of every node in the graph to the specified value.
func SetGraphFontsize(graph *gographviz.Graph, graphValue, nodeValue float32) {
	// Iterate over every node in the graph
	for _, node := range graph.Nodes.Nodes {
		node.Attrs["fontsize"] = fmt.Sprintf(`"%.1f"`, nodeValue)
	}

	// Iterate over every graph/subgraph in the graph
	for _, graph := range graph.SubGraphs.SubGraphs {
		graph.Attrs["fontsize"] = fmt.Sprintf(`"%.1f"`, graphValue)
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

	SetGraphGlobalImagePath(graph, GlobalImagePath)
	SetGraphAttrs(graph)
	CreateSubgraphsForGrouppingNodes(graph)
	AddImageLabel(graph)
	PositionNodeLabelTo(graph, NODE_LABEL_LOCATION)
	PositionGraphLabelTo(graph, GRAPH_LABEL_LOCATION)
	SetGraphFontsize(graph, 26.0, 20.0)
	AddMarginToNodes(graph, 1.5)

	return graph, nil
}

// Function creates a subgraph for each node that is a groupping node
func CreateSubgraphsForGrouppingNodes(graph *gographviz.Graph) error {
	DfsTraversal(graph)
	return nil
}

// DfsTraversal performs a depth-first search traversal on the graph.
func DfsTraversal(graph *gographviz.Graph) {
	visited := make(map[string]bool)
	for _, node := range graph.Nodes.Nodes {
		if !visited[node.Name] {
			dfsHelper(graph, node.Name, visited)
		}
	}
}

// dfsHelper is a recursive helper function for DFS traversal.
func dfsHelper(graph *gographviz.Graph, node string, visited map[string]bool) {
	// Mark the current node as visited.
	visited[node] = true
	//fmt.Println("Visited:", node)

	// Add subgraph if the node is a groupping resource
	label := strings.Trim(graph.Nodes.Lookup[node].Attrs["label"], `"`)

	// TODO: Use global config for this
	var groupingLabels = []string{"azurerm_virtual_network", "azurerm_resource_group", "azurerm_subnet"}

	if IsResourceNode(label) {
		// label="azurerm_linux_virtual_machine.vm"
		// azurerm_linux_virtual_machine is the part we need
		if foundGroupingResource := contains(groupingLabels, strings.Split(label, ".")[0]); foundGroupingResource {
			fmt.Println("Visited a groupping node: " + node)

			// special case because root graph name doesn't have quotes
			parentNode := FindNodeParent(node, graph)
			if parentNode != "G" {
				parentNode = strings.Trim(parentNode, `"`)
				parentNode = fmt.Sprintf(`"%s"`, parentNode)
			}
			// INFO: do not add only the resource type, add the full name, so later we can add the node inside this subGraph
			err := graph.AddSubGraph(parentNode, fmt.Sprintf(`"%s"`, "cluster_"+label), map[string]string{"label": fmt.Sprintf(`"%s"`, label)})
			if err != nil {
				fmt.Println("ERROR: Got an error trying to add subgraph")
			}

			SetChildOf(fmt.Sprintf(`"%s"`, "cluster_"+label), node, graph)

			rn := findAllReachingNodes(node, graph)

			// Step 2
			for _, subGraph := range graph.SubGraphs.Sorted() {
				n := strings.TrimLeft(strings.Trim(subGraph.Name, `"`), "cluster_")
				n = `"` + n + `"`
				fmt.Println("checking if i have an edge between " + n + " and " + node)
				if CheckEdgeExistence(n, node, graph) {
					fmt.Println("--- graph " + n + " has to be child of the current found graph " + node)
					fmt.Println("Setting " + fmt.Sprintf(`"%s"`, "cluster_"+n) + " as child subgraph of " + fmt.Sprintf(`"%s"`, "cluster_"+node))
					SetChildOf(fmt.Sprintf(`"%s"`, "cluster_"+strings.Trim(node, `"`)), fmt.Sprintf(`"%s"`, "cluster_"+strings.Trim(n, `"`)), graph)
				}
			}

			// Step 1
			for _, reachingNode := range rn {
				if reachingNodeParent := FindNodeParent(reachingNode, graph); reachingNodeParent == "G" {
					fmt.Println("reaching node " + reachingNode + " will be set as child  of " + fmt.Sprintf(`"%s"`, "cluster_"+label))
					SetChildOf(fmt.Sprintf(`"%s"`, "cluster_"+label), reachingNode, graph)
				}
			}

		} else {
			fmt.Println("Visited:", node)
		}
	}

	// Get all the edges starting from the current node.
	for _, edge := range graph.Edges.SrcToDsts[node] {
		for _, dst := range edge {
			if !visited[dst.Dst] {
				dfsHelper(graph, dst.Dst, visited)
			}
		}
	}
}

// change to be foundGroupingResource
// receive a node and check the label
func contains(arr []string, str string) bool {
	for _, item := range arr {
		if item == str {
			return true
		}
	}
	return false
}

func FindNodeParent(nodeName string, graph *gographviz.Graph) string {
	relations := graph.Relations
	parents, ok := relations.ChildToParents[nodeName]
	var nodeParent string

	if !ok {
		fmt.Printf("No parents found for node %s \n", nodeName) // might need to be an error and return
	}

	for parent := range parents {
		fmt.Printf("Parents of node %s is %s \n", nodeName, parent)
		nodeParent = parent
	}

	return nodeParent

}

// Function updates the relations in order to make nodeName a child of the given graph/subgraph,
// nodeName can also be a subgraph
func SetChildOf(graphName string, nodeName string, graph *gographviz.Graph) {
	relations := graph.Relations

	// Initialize parent-to-children and child-to-parents maps if they don't exist
	if _, exists := relations.ParentToChildren[graphName]; !exists {
		relations.ParentToChildren[graphName] = make(map[string]bool)
	}
	if _, exists := relations.ChildToParents[nodeName]; !exists {
		relations.ChildToParents[nodeName] = make(map[string]bool)
	}

	// Update relations
	relations.ParentToChildren[graphName][nodeName] = true
	relations.ChildToParents[nodeName][graphName] = true

	// Optionally, you may want to remove the node from any other parents to ensure it is only a child of the specified graphName
	for parent, children := range relations.ParentToChildren {
		if parent != graphName {
			delete(children, nodeName)
		}
	}
	for child, parents := range relations.ChildToParents {
		if child == nodeName && len(parents) > 1 {
			for parent := range parents {
				if parent != graphName {
					delete(parents, parent)
				}
			}
		}
	}
}

// CheckEdgeExistence checks if there is an edge from node1 to node2 in the graph.
func CheckEdgeExistence(node1, node2 string, graph *gographviz.Graph) bool {
	// Check if node1 has edges directed towards node2
	if edges, exists := graph.Edges.SrcToDsts[node1]; exists {
		for _, edgeList := range edges {
			for _, edge := range edgeList {
				if edge.Dst == node2 {
					return true
				}
			}
		}
	}
	return false
}

// findAllReachingNodes performs a reverse DFS to find all nodes that can reach the given node.
func findAllReachingNodes(targetNode string, graph *gographviz.Graph) []string {
	visited := make(map[string]bool)
	var result []string

	var reverseDfs func(string)
	reverseDfs = func(n string) {
		if visited[n] {
			return
		}
		visited[n] = true
		result = append(result, n)

		if edges, ok := graph.Edges.DstToSrcs[n]; ok {
			for src := range edges {
				reverseDfs(src)
			}
		}
	}

	reverseDfs(targetNode)
	return result
}

// findAllReachableNodes performs a DFS to find all reachable nodes from the given node.
func findAllReachableNodes(startNode string, graph *gographviz.Graph) []string {
	visited := make(map[string]bool)
	var result []string

	var dfs func(string)
	dfs = func(n string) {
		if visited[n] {
			return
		}
		visited[n] = true
		result = append(result, n)

		if edges, ok := graph.Edges.SrcToDsts[n]; ok {
			for dst := range edges {
				dfs(dst)
			}
		}
	}

	dfs(startNode)
	return result
}

// printRelations nicely prints the Relations struct from gographviz.Graph
func printRelations(graph *gographviz.Graph) {
	relations := graph.Relations

	fmt.Println("ParentToChildren relationships:")
	for parent, children := range relations.ParentToChildren {
		fmt.Printf("Parent: %s", parent)
		for child := range children {
			fmt.Printf("  Child: %s ;", child)
		}
		fmt.Println()
	}

	fmt.Println("\nChildToParents relationships:")
	for child, parents := range relations.ChildToParents {
		fmt.Printf("Child: %s", child)
		for parent := range parents {
			fmt.Printf("  Parent: %s ;", parent)
		}
		fmt.Println()
	}
}
