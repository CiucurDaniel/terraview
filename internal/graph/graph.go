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
	fmt.Println("EDGES BEFORE")
	printEdges(graph)
	fmt.Println("..........")
	ExpandNodeCreatedWithList(graph, handler)
	CleanUpEdges(graph)
	BetaCreateSubgraphsForGroupingNodes(graph)
	AddImageLabel(graph)
	PositionNodeLabelTo(graph, NODE_LABEL_LOCATION)
	PositionGraphLabelTo(graph, GRAPH_LABEL_LOCATION)
	SetGraphFontsize(graph, 26.0, 20.0)
	AddMarginToNodes(graph, 1.5)
	SetSubgraphMargins(graph, CalculateMaxDepth(graph), 10)

	err = AddImportantAttributesToLabels(graph, cfg, handler)
	if err != nil {
		return nil, fmt.Errorf("failed to add important attributes to labels: %v", err)
	}

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
				// fmt.Println("DEBUG: Will get important attributes for " + resourceIdentifier)
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
		// fmt.Printf("Parents of node %s is %s \n", nodeName, parent)
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

// printEdges prints all edges in the graph.
func printEdges(graph *gographviz.Graph) {
	for _, edge := range graph.Edges.Sorted() {
		fmt.Printf("Edge: %s -> %s\n", edge.Src, edge.Dst)
	}
}

func ExpandNodeCreatedWithList(graph *gographviz.Graph, handler *tfstatereader.TFStateHandler) {
	visited := make(map[string]bool)

	var dfs func(node string)
	dfs = func(node string) {
		if visited[node] {
			return
		}
		visited[node] = true

		// Get the current label of the node
		label := graph.Nodes.Lookup[node].Attrs["label"]
		label = strings.Trim(label, `"`)

		// Check if the node was created with a list (count or for_each)
		if handler.IsCreatedWithList(label) {
			// Get the list of actual names for the resource
			resourceNames, err := handler.GetListOfNamesForResource(label)
			if err != nil {
				log.Printf("error getting list of names for resource %s: %v", label, err)
				return
			}

			// Get the parent graph of the original node
			parentGraph := FindNodeParent(node, graph)

			// Create new nodes and edges based on the list of names
			for _, resourceName := range resourceNames {
				// Create a new node with the same attributes as the original node
				newNodeName := strings.Replace(node, label, resourceName, 1)
				newNodeAttrs := gographviz.Attrs{}
				for k, v := range graph.Nodes.Lookup[node].Attrs {
					newNodeAttrs[gographviz.Attr(k)] = v
				}
				newNodeAttrs["label"] = fmt.Sprintf(`"%s"`, resourceName)

				// Add the new node to the graph
				graph.AddNode(parentGraph, newNodeName, attrsToMap(newNodeAttrs))

				// Create edges from the new node to all the destinations of the original node
				for _, edgeList := range graph.Edges.SrcToDsts[node] {
					for _, edge := range edgeList {
						newEdge := gographviz.Edge{
							Src:   newNodeName,
							Dst:   edge.Dst,
							Attrs: edge.Attrs.Copy(),
						}
						graph.AddEdge(newEdge.Src, newEdge.Dst, true, attrsToMap(newEdge.Attrs))
					}
				}

				// Create edges to the new node from all the sources of the original node
				for _, edgeList := range graph.Edges.DstToSrcs[node] {
					for _, edge := range edgeList {
						newEdge := gographviz.Edge{
							Src:   edge.Src,
							Dst:   newNodeName,
							Attrs: edge.Attrs.Copy(),
						}
						graph.AddEdge(newEdge.Src, newEdge.Dst, true, attrsToMap(newEdge.Attrs))
					}
				}
			}

			// Remove the original node and its edges
			graph.RemoveNode(parentGraph, node)
		}

		// Visit all the children nodes
		for _, edgeList := range graph.Edges.SrcToDsts[node] {
			for _, edge := range edgeList {
				dfs(edge.Dst)
			}
		}
	}

	// Start DFS from all top-level nodes
	for _, node := range graph.Nodes.Nodes {
		if !visited[node.Name] {
			dfs(node.Name)
		}
	}
}

func attrsToMap(attrs gographviz.Attrs) map[string]string {
	result := make(map[string]string)
	for k, v := range attrs {
		result[string(k)] = v
	}
	return result
}

func CleanUpEdges(graph *gographviz.Graph) {
	visited := make(map[string]bool)

	var dfs func(node string)
	dfs = func(node string) {
		if visited[node] {
			return
		}
		visited[node] = true

		// Get the current label of the node
		label := graph.Nodes.Lookup[node].Attrs["label"]
		label = strings.Trim(label, `"`)

		// Check if the node is a list node
		var currentIndex string
		if strings.Contains(label, "[") && strings.Contains(label, "]") {
			currentIndex = label[strings.Index(label, "[")+1 : strings.Index(label, "]")]
		}

		// Remove edges based on indices or keys
		var edgesToRemove []*gographviz.Edge
		for _, edgeList := range graph.Edges.SrcToDsts[node] {
			for _, edge := range edgeList {
				dstLabel := graph.Nodes.Lookup[edge.Dst].Attrs["label"]
				dstLabel = strings.Trim(dstLabel, `"`)

				var dstIndex string
				if strings.Contains(dstLabel, "[") && strings.Contains(dstLabel, "]") {
					dstIndex = dstLabel[strings.Index(dstLabel, "[")+1 : strings.Index(dstLabel, "]")]
				}

				if currentIndex != "" && dstIndex != "" && currentIndex != dstIndex {
					edgesToRemove = append(edgesToRemove, edge)
				}
			}
		}

		// Actually remove the edges
		for _, edge := range edgesToRemove {
			removeEdgeFromGraph(graph, edge)
		}

		// Visit all the children nodes
		for _, edgeList := range graph.Edges.SrcToDsts[node] {
			for _, edge := range edgeList {
				dfs(edge.Dst)
			}
		}
	}

	// Start DFS from all top-level nodes
	for _, node := range graph.Nodes.Nodes {
		if !visited[node.Name] {
			dfs(node.Name)
		}
	}
}

func removeEdgeFromGraph(graph *gographviz.Graph, edge *gographviz.Edge) {
	// Remove the edge from SrcToDsts
	srcEdges := graph.Edges.SrcToDsts[edge.Src][edge.Dst]
	for i, e := range srcEdges {
		if e.Dst == edge.Dst && e.Src == edge.Src {
			graph.Edges.SrcToDsts[edge.Src][edge.Dst] = append(srcEdges[:i], srcEdges[i+1:]...)
			break
		}
	}

	// Remove the edge from DstToSrcs
	dstEdges := graph.Edges.DstToSrcs[edge.Dst][edge.Src]
	for i, e := range dstEdges {
		if e.Dst == edge.Dst && e.Src == edge.Src {
			graph.Edges.DstToSrcs[edge.Dst][edge.Src] = append(dstEdges[:i], dstEdges[i+1:]...)
			break
		}
	}

	// Remove the edge from the main edges list
	for i, e := range graph.Edges.Edges {
		if e.Dst == edge.Dst && e.Src == edge.Src {
			graph.Edges.Edges = append(graph.Edges.Edges[:i], graph.Edges.Edges[i+1:]...)
			break
		}
	}
}

// findRootNode identifies the node with no outgoing edges.
func findRootNode(graph *gographviz.Graph) string {
	outDegree := make(map[string]int)

	// Initialize out-degree for all nodes
	for _, node := range graph.Nodes.Nodes {
		outDegree[node.Name] = 0
	}

	// Calculate out-degree for each node
	for _, edge := range graph.Edges.Edges {
		outDegree[edge.Src]++
	}

	// Find the node with out-degree 0
	for node, degree := range outDegree {
		if degree == 0 {
			return node
		}
	}

	return ""
}

// BFS performs a breadth-first search on the graph starting from the given node and returns the list of visited nodes.
func BFS(graph *gographviz.Graph, startNode string) []string {
	visited := make(map[string]bool)
	queue := []string{startNode}
	var visitedNodes []string

	for len(queue) > 0 {
		// Dequeue a node from the front of the queue
		currentNode := queue[0]
		queue = queue[1:]

		// If the node has already been visited, skip it
		if visited[currentNode] {
			continue
		}

		// Mark the node as visited
		visited[currentNode] = true

		// Add the current node to the visited list
		visitedNodes = append(visitedNodes, currentNode)

		// Enqueue all parent nodes that have not been visited
		for _, edge := range graph.Edges.DstToSrcs[currentNode] {
			for _, e := range edge {
				if !visited[e.Src] {
					queue = append(queue, e.Src)
				}
			}
		}
	}

	return visitedNodes
}

func BetaCreateSubgraphsForGroupingNodes(graph *gographviz.Graph) {

	nodes := BFS(graph, findRootNode(graph))

	var groupingLabels = []string{"azurerm_virtual_network", "azurerm_resource_group", "azurerm_subnet"} // TODO: Use global config for this

	for _, node := range nodes {
		// Node is "azurerm_linux_virtual_machine.vm"
		// Needed part is azurerm_linux_virtual_machine

		cleanNodeName := strings.Trim(node, `"`)
		if foundGroupingResource := contains(groupingLabels, strings.Split(cleanNodeName, ".")[0]); foundGroupingResource {

			// 1. Prepare cluster name for node

			clusterName := fmt.Sprintf(`"%s"`, "cluster_"+cleanNodeName)
			parentGraph := FindNodeParent(node, graph)

			// 2. Create the SubGraph
			err := graph.AddSubGraph(parentGraph, clusterName, map[string]string{"label": clusterName})
			if err != nil {
				fmt.Println("ERROR: Got an error trying to add subgraph")
			}

			// 3. Add all reaching nodes as children of the new SubGraph
			for _, reachingNode := range findAllReachingNodes(node, graph) {
				SetChildOf(clusterName, reachingNode, graph)
			}

		}
	}

}
