### Concept find parent graph/sub-graph of a node


```go
package cmd

import (
	"fmt"

	"github.com/awalterschulze/gographviz"
)

func etc() {
	// Create a new graph.
	graph := gographviz.NewGraph()
	graph.SetName("G")
	graph.SetDir(true) // Directed graph

	// Add nodes to the main graph.
	graph.AddNode("G", "A", nil)
	graph.AddNode("G", "B", nil)

	// Add a subgraph.
	subgraph1 := "cluster_0"
	graph.AddSubGraph("G", subgraph1, map[string]string{"label": `"Subgraph 1"`})
	graph.AddNode(subgraph1, "C", nil)
	graph.AddNode(subgraph1, "D", nil)

	// Add a subgraph within the first subgraph.
	subgraph2 := "cluster_1_2"
	graph.AddSubGraph("cluster_0", subgraph2, map[string]string{"label": `"Subgraph 1.1"`})
	graph.AddNode(subgraph2, "E", nil)
	graph.AddNode(subgraph2, "F", nil)

	// Add edges.
	graph.AddEdge("A", "B", true, nil)
	graph.AddEdge("B", "C", true, nil)
	graph.AddEdge("C", "E", true, nil)
	graph.AddEdge("E", "F", true, nil)
	graph.AddEdge("D", "F", true, nil)

	// Render the graph to a DOT format string.
	output := graph.String()
	fmt.Println(output)

	//printRelations(graph)

	FindNodeParent("E", graph)
}

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

func FindNodeParent(nodeName string, graph *gographviz.Graph) {
	relations := graph.Relations
	parents, ok := relations.ChildToParents[nodeName]

	if !ok {
		fmt.Printf("No parents found for node %s \n", nodeName) // might need to be an error and return
	}

	for parent := range parents {
		fmt.Printf("Parents of node %s is %s \n", nodeName, parent)
	}
	//fmt.Println(nodeName)
}

```

### DFS traversal


```go
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
	fmt.Println("Visited:", node)

	// Get all the edges starting from the current node.
	for _, edge := range graph.Edges.SrcToDsts[node] {
		for _, dst := range edge {
			if !visited[dst.Src] {
				dfsHelper(graph, dst.Src, visited)
			}
		}
	}
}
```

### BFS traversal


```go
// BfsTraversal performs a breadth-first search traversal on the graph.
func BfsTraversal(graph *gographviz.Graph) {
	visited := make(map[string]bool)
	queue := []string{}

	// Start BFS from each node that hasn't been visited yet
	for _, node := range graph.Nodes.Nodes {
		if !visited[node.Name] {
			queue = append(queue, node.Name)
			visited[node.Name] = true

			// Process the queue
			for len(queue) > 0 {
				// Dequeue the next node
				currentNode := queue[0]
				queue = queue[1:]

				// Visit the node
				fmt.Println("Visited:", currentNode)

				// Enqueue all adjacent nodes
				for _, edge := range graph.Edges.SrcToDsts[currentNode] {
					for _, dst := range edge {
						if !visited[dst.Src] {
							queue = append(queue, dst.Src)
							visited[dst.Src] = true
						}
					}
				}
			}
		}
	}
}
```

### Reverse DFS traversal


```golang
// ReverseDfsTraversal performs a depth-first search traversal on the graph in reverse direction.
func ReverseDfsTraversal(graph *gographviz.Graph) {
	visited := make(map[string]bool)
	for _, node := range graph.Nodes.Nodes {
		if !visited[node.Name] {
			reverseDfsHelper(graph, node.Name, visited)
		}
	}
}

// reverseDfsHelper is a recursive helper function for reverse DFS traversal.
func reverseDfsHelper(graph *gographviz.Graph, node string, visited map[string]bool) {
	// Mark the current node as visited.
	visited[node] = true
	fmt.Println("Visited:", node)

	// Get all the edges ending at the current node (reverse direction).
	for _, edges := range graph.Edges.DstToSrcs[node] {
		for _, src := range edges {
			if !visited[src.Src] {
				reverseDfsHelper(graph, src.Src, visited)
			}
		}
	}
}

```


### BFS traversal


```go
// findRootNode identifies the node with no incoming edges.
func findRootNode(graph *gographviz.Graph) string {
	inDegree := make(map[string]int)

	// Initialize in-degree for all nodes
	for _, node := range graph.Nodes.Nodes {
		inDegree[node.Name] = 0
	}

	// Calculate in-degree for each node
	for _, edgeList := range graph.Edges.Edges {
		inDegree[edgeList.Dst]++
	}

	// Find the node with in-degree 0
	for node, degree := range inDegree {
		if degree == 0 {
			return node
		}
	}

	return ""
}

// BFS performs a breadth-first search on the graph starting from the given node.
func BFS(graph *gographviz.Graph, startNode string) {
	visited := make(map[string]bool)
	queue := []string{startNode}

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

		// Print the current node
		fmt.Println("Visited:", currentNode)

		// Enqueue all adjacent nodes that have not been visited
		for _, edgeList := range graph.Edges.SrcToDsts[currentNode] {
			for _, edge := range edgeList {
				if !visited[edge.Dst] {
					queue = append(queue, edge.Dst)
				}
			}
		}
	}
}
```