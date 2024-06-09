package graph

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/CiucurDaniel/terraview/internal/config"
	"github.com/CiucurDaniel/terraview/internal/tfstatereader"
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

	// Get the absolute path to the terraform executable
	terraformPath, err := exec.LookPath("terraform")
	if err != nil {
		return nil, fmt.Errorf("failed to find terraform executable in PATH: %v", err)
	}

	// Execute "terraform graph" command in the specified directory
	cmd := exec.Command(terraformPath, "graph")
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

// PrepareGraphForPrinting is a facade function for preparing the graph for printing.
// It obtains the graph data, adds image labels to nodes, and returns the modified graph.
func PrepareGraphForPrinting(dirPath string, cfg *config.Config, handler *tfstatereader.TFStateHandler) (*gographviz.Graph, error) {
	// Obtain the graph
	graph, err := ObtainGraph(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain graph data: %v", err)
	}

	SetGraphGlobalImagePath(graph, GlobalImagePath)
	SetGraphAttrs(graph)
	CreateSubgraphsForGrouppingNodes(graph)
	fmt.Println("Done with grouping nodes")
	AddImageLabel(graph)
	PositionNodeLabelTo(graph, NODE_LABEL_LOCATION)
	PositionGraphLabelTo(graph, GRAPH_LABEL_LOCATION)
	SetGraphFontsize(graph, 26.0, 20.0)
	AddMarginToNodes(graph, 1.5)
	SetSubgraphMargins(graph, CalculateMaxDepth(graph), 10)

	fmt.Println("Starting to add AddImportantAttributesToLabels")
	err = AddImportantAttributesToLabels(graph, cfg, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to add important attributes to labels: %v", err)
	}
	fmt.Println("Ending adding AddImportantAttributesToLabels")

	return graph, nil
}

// SetGraphGlobalImagePath sets the global image path for the graph.
func SetGraphGlobalImagePath(graph *gographviz.Graph, path string) {
	graph.Attrs.Add("imagepath", fmt.Sprintf(`"%s"`, path))
}

// SetGraphAttrs func will set:
// - compound = true
// - newrank = true
// - rankdir = "TD
// - (optional) call SetGraphGlobalImagePath
func SetGraphAttrs(graph *gographviz.Graph) {
	graph.Attrs["compound"] = `"true"`
	graph.Attrs["rankdir"] = `"BT"`
	graph.Attrs["newrank"] = `"true"`
	graph.Attrs["nodesep"] = `"4"`
	graph.Attrs["ranksep"] = `"1.5"`

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

// SetGraphFontsize sets the fontsize attribute of every node in the graph to the specified value.
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

// CalculateMaxDepth calculates the maximum depth of nested subgraphs in the graph.
func CalculateMaxDepth(graph *gographviz.Graph) int {
	var maxDepth int
	var dfs func(node string, depth int)
	dfs = func(node string, depth int) {
		if depth > maxDepth {
			maxDepth = depth
		}
		for child := range graph.Relations.ParentToChildren[node] {
			dfs(child, depth+1)
		}
	}
	for _, subgraph := range graph.SubGraphs.SubGraphs {
		dfs(subgraph.Name, 1)
	}
	return maxDepth
}

// SetSubgraphMargins sets the margin attribute for each subgraph based on its depth.
func SetSubgraphMargins(graph *gographviz.Graph, maxDepth, baseMargin int) {
	var setMargins func(node string, depth int)
	setMargins = func(node string, depth int) {
		marginValue := (maxDepth - depth + 1) * baseMargin
		if subgraph, exists := graph.SubGraphs.SubGraphs[node]; exists {
			subgraph.Attrs["margin"] = fmt.Sprintf(`"%d"`, marginValue)
		}
		for child := range graph.Relations.ParentToChildren[node] {
			setMargins(child, depth+1)
		}
	}
	for _, subgraph := range graph.SubGraphs.SubGraphs {
		setMargins(subgraph.Name, 1)
	}
}

// TODO: Create func which puts consecutive identical resources on same rank
// similar to what was done here: https://stackoverflow.com/questions/58832678/how-to-separate-picture-and-label-of-a-node-with-graphviz

// AddImportantAttributesToLabels traverses nodes in the graph, checks if the node is in the important attributes resource list,
// calls GetImportantAttributes for the given node, and adds the result within the label field with newlines in between.
func AddImportantAttributesToLabels(graph *gographviz.Graph, cfg *config.Config, handler *tfstatereader.TFStateHandler) error {
	for _, node := range graph.Nodes.Nodes {
		// Get the current label of the node
		label := node.Attrs["label"]

		// Remove existing quotes, if any
		label = strings.Trim(label, `"`)

		// Check if the node represents a resource
		parts := strings.Split(label, ".")

		// If the label is empty or doesn't contain a dot, skip the node
		if len(parts) < 2 {
			continue
		}

		resourceType := parts[0]
		resourceName := parts[1]

		// Check if the resource type is in the important attributes list
		for _, resConfig := range cfg.ImportantAttributes {
			if resConfig.Name == resourceType {
				// Construct the resource identifier (e.g., azurerm_linux_virtual_machine.vm_1)
				resourceIdentifier := fmt.Sprintf("%s.%s", resourceType, resourceName)

				// Get important attributes for the resource
				importantAttrs, err := handler.GetImportantAttributes(resourceIdentifier)
				if err != nil {
					return fmt.Errorf("failed to get important attributes for %s: %v", resourceIdentifier, err)
				}

				// Join important attributes with newlines
				attrString := strings.Join(importantAttrs, "\\n")

				// Update the label with important attributes
				newLabel := fmt.Sprintf("%s\\n%s", label, attrString)
				node.Attrs["label"] = fmt.Sprintf(`"%s"`, newLabel)
				break
			}
		}
	}

	return nil
}

// CreateSubgraphsForGrouppingNodes creates a subgraph for each node that is a grouping node
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

// SetChildOf function updates the relations in order to make nodeName a child of the given graph/subgraph,
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

// ExpandNodeCreatedWithList expands nodes created with count or for_each in the graph.
func ExpandNodeCreatedWithList(graph *gographviz.Graph, handler *tfstatereader.TFStateHandler) {
	// Store the new nodes and edges to be added
	var newNodes []gographviz.Node
	var newEdges []gographviz.Edge

	// Track relationships for new nodes
	newParentToChildren := make(map[string]map[string]bool)
	newChildToParents := make(map[string]map[string]bool)

	// Iterate over every node in the graph
	for _, node := range graph.Nodes.Nodes {
		// Get the current label of the node
		label := node.Attrs["label"]
		label = strings.Trim(label, `"`)

		// Check if the node was created with a list
		if handler.IsCreatedWithList(label) {
			// Get the list of actual names for the resource
			resourceNames, err := handler.GetListOfNamesForResource(label)
			if err != nil {
				log.Printf("error getting list of names for resource %s: %v", label, err)
				continue
			}

			// Create new nodes and edges based on the list of names
			for _, resourceName := range resourceNames {
				// Create a new node with the same attributes as the original node
				newNodeName := strings.Replace(node.Name, label, resourceName, 1)
				newNodeAttrs := gographviz.Attrs{}
				for k, v := range node.Attrs {
					newNodeAttrs[gographviz.Attr(k)] = v
				}
				newNodeAttrs["label"] = fmt.Sprintf(`"%s"`, resourceName)

				// Add the new node to the list
				newNodes = append(newNodes, gographviz.Node{
					Name:  newNodeName,
					Attrs: newNodeAttrs,
				})

				// Create edges from the new node to all the destinations of the original node
				for _, edgeList := range graph.Edges.SrcToDsts[node.Name] {
					for _, edge := range edgeList {
						newEdges = append(newEdges, gographviz.Edge{
							Src:   newNodeName,
							Dst:   edge.Dst,
							Attrs: edge.Attrs.Copy(),
						})

						// Update new relationships
						if newParentToChildren[newNodeName] == nil {
							newParentToChildren[newNodeName] = make(map[string]bool)
						}
						newParentToChildren[newNodeName][edge.Dst] = true
					}
				}

				// Create edges to the new node from all the sources of the original node
				for _, edgeList := range graph.Edges.DstToSrcs[node.Name] {
					for _, edge := range edgeList {
						newEdges = append(newEdges, gographviz.Edge{
							Src:   edge.Src,
							Dst:   newNodeName,
							Attrs: edge.Attrs.Copy(),
						})

						// Update new relationships
						if newChildToParents[newNodeName] == nil {
							newChildToParents[newNodeName] = make(map[string]bool)
						}
						newChildToParents[newNodeName][edge.Src] = true
					}
				}
			}

			// Remove the original node and its edges
			parentGraph := FindNodeParent(node.Name, graph)
			graph.RemoveNode(parentGraph, node.Name)
		}
	}

	// Add the new nodes and edges to the graph
	for _, newNode := range newNodes {
		// Convert newNode.Attrs to map[string]string
		nodeAttrs := make(map[string]string)
		for k, v := range newNode.Attrs {
			nodeAttrs[string(k)] = v
		}
		graph.AddNode("", newNode.Name, nodeAttrs)
	}
	for _, newEdge := range newEdges {
		// Convert newEdge.Attrs to map[string]string
		edgeAttrs := make(map[string]string)
		for k, v := range newEdge.Attrs {
			edgeAttrs[string(k)] = v
		}
		graph.AddEdge(newEdge.Src, newEdge.Dst, true, edgeAttrs)
	}

	// Update relationships
	for parent, children := range newParentToChildren {
		if _, exists := graph.Relations.ParentToChildren[parent]; !exists {
			graph.Relations.ParentToChildren[parent] = make(map[string]bool)
		}
		for child := range children {
			graph.Relations.ParentToChildren[parent][child] = true
		}
	}

	for child, parents := range newChildToParents {
		if _, exists := graph.Relations.ChildToParents[child]; !exists {
			graph.Relations.ChildToParents[child] = make(map[string]bool)
		}
		for parent := range parents {
			graph.Relations.ChildToParents[child][parent] = true
		}
	}
}
