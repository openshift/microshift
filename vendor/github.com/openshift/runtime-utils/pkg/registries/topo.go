package registries

import (
	"container/list"
	"fmt"
	"sort"
)

// topoNode is a node in the tsort graph. In our case, it is a mirror host[:port].
type topoNode = string // An alias, so that the users don't have to explicitly cast data.

// internalTopoNode is the topoGraph representation of a node
// There is exactly one instance of *internalTopoNode for each .public value.
type internalTopoNode struct {
	public topoNode
	used   bool
	ins    map[*internalTopoNode]struct{} // Contains unprocessed predecessors (destructively modified by g.Sorted())
	outs   map[*internalTopoNode]struct{} // Contains all successors (i.e. not modified by g.Sorted())
}

// topoGraph is a directed graph, ideally acyclic, with the ability to return a topologically-sorted sequence, or something close to that.
// The graph is built by calling .AddEdge (using possibly never-before-mentioned topoNode values)
// There isn’t currently a way to add a completely disconnected node (it would be trivial to add, but we don’t need it right now).
type topoGraph struct {
	// This is not all that efficient, the hash map lookups by topoNode == string are likely more costly than any other
	// use of the graph, but the graphs we need to handle are very small and readability is more important.
	nodes map[topoNode]*internalTopoNode
}

// newTopoGraph returns an empty topoGraph.
func newTopoGraph() *topoGraph {
	return &topoGraph{
		nodes: map[topoNode]*internalTopoNode{},
	}
}

// getNode returns g.nodes[node], creating it if it does not exist yet
func (g *topoGraph) getNode(node topoNode) *internalTopoNode {
	res, ok := g.nodes[node]
	if !ok {
		res = &internalTopoNode{
			public: node,
			used:   false,
			ins:    map[*internalTopoNode]struct{}{},
			outs:   map[*internalTopoNode]struct{}{},
		}
		g.nodes[node] = res
	}
	return res
}

// AddEdge adds a directed edge (from, to)
func (g *topoGraph) AddEdge(from, to topoNode) {
	fromTN := g.getNode(from)
	toTN := g.getNode(to)
	// This automatically ignores redundant edges
	fromTN.outs[toTN] = struct{}{}
	toTN.ins[fromTN] = struct{}{}
}

// Sorted returns the nodes of g, sorted to respect the edges in the graph, _if possible_.
func (g *topoGraph) Sorted() ([]topoNode, error) {
	// NOTE: The order of the returned nodes should be fully deterministic.

	allNodes := []*internalTopoNode{}
	for _, n := range g.nodes {
		allNodes = append(allNodes, n)
	}
	// This is O(N log N * string_length), more costly than anything else; but it’s fast enough, and we do want the behavior
	// to be deterministic.
	sort.Slice(allNodes, func(i, j int) bool {
		return allNodes[i].public < allNodes[j].public
	})

	// noIncoming is a queue of nodes with no (unprocessed) predecessors
	noIncoming := list.List{}
	for _, n := range allNodes {
		if len(n.ins) == 0 {
			noIncoming.PushBack(n) // noIncoming is a subsequence of allNodes, so the nodes are added in sorted order
		}
	}

	res := []topoNode{}
	allNodeIndex := 0
	for noIncoming.Len() != 0 || allNodeIndex < len(allNodes) {
		// Find the next node to output
		var node *internalTopoNode
		if noIncoming.Len() != 0 {
			node = noIncoming.Remove(noIncoming.Front()).(*internalTopoNode)
			if node.used {
				return nil, fmt.Errorf("internal error in topoGraph.Sorted: node %#v already used", node.public)
			}
		} else { // There is an (unbroken) loop; break it by using the first unused node in the sorted order
			node = allNodes[allNodeIndex]
			allNodeIndex++
			if node.used {
				continue // Return to the allNodeIndex < len(allNodes) check
			}
		}

		node.used = true
		res = append(res, node.public)

		nextBatch := []*internalTopoNode{}
		for out := range node.outs {
			delete(out.ins, node)
			if len(out.ins) == 0 && !out.used { // out.used may be true if out was used to break a loop.
				nextBatch = append(nextBatch, out)
			}
		}
		sort.Slice(nextBatch, func(i, j int) bool {
			return nextBatch[i].public < nextBatch[j].public
		})
		for _, next := range nextBatch {
			noIncoming.PushBack(next)
		}
	}

	if len(res) != len(g.nodes) {
		return nil, fmt.Errorf("internal error in topoGraph.Sorted: Not all nodes used")
	}
	return res, nil
}
