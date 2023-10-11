package main

import (
	"fmt"

	"github.com/emicklei/dot"
)

func main() {
	//fmt.Println("Landscape CLI")

	di := dot.NewGraph(dot.Directed)
	outside := di.Node("Outside")

	// A
	clusterA := di.Subgraph("Cluster A", dot.ClusterOption{})
	insideOne := clusterA.Node("one").Attr("image", "assets/azure/icons/storage/screenshot.png")
	insideTwo := clusterA.Node("two")

	// B
	clusterB := di.Subgraph("Cluster B", dot.ClusterOption{})
	insideThree := clusterB.Node("three")
	insideFour := clusterB.Node("four")

	// edges
	outside.Edge(insideFour).Edge(insideOne).Edge(insideTwo).Edge(insideThree).Edge(outside)

	fmt.Println(di.String())
}
