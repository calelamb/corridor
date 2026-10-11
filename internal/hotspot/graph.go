package hotspot

import (
	"container/heap"
	"errors"
	"fmt"
	"math"
)

// RoadSegment represents a road segment in the network.
type RoadSegment struct {
	ID      string
	LengthM float64
}

// Edge represents an undirected connection between road segments.
type Edge struct {
	To     string
	Weight float64
}

// RoadGraph represents a graph of connected road segments.
type RoadGraph struct {
	segments map[string]*RoadSegment
	order    []string // deterministic segment insertion order
	adj      map[string][]Edge
}

// NewRoadGraph initializes an empty RoadGraph.
func NewRoadGraph() *RoadGraph {
	return &RoadGraph{
		segments: make(map[string]*RoadSegment),
		order:    make([]string, 0),
		adj:      make(map[string][]Edge),
	}
}

// AddSegment adds a road segment to the graph.
func (g *RoadGraph) AddSegment(id string, lengthM float64) error {
	if id == "" {
		return errors.New("segment ID cannot be empty")
	}
	if lengthM <= 0 {
		return errors.New("segment length must be positive")
	}
	if _, exists := g.segments[id]; exists {
		// Update length if already present
		g.segments[id].LengthM = lengthM
		return nil
	}
	seg := &RoadSegment{ID: id, LengthM: lengthM}
	g.segments[id] = seg
	g.order = append(g.order, id)
	return nil
}

// AddEdge adds an undirected edge between two road segments with a specified distance in meters.
func (g *RoadGraph) AddEdge(fromID, toID string, distanceM float64) error {
	if fromID == "" || toID == "" {
		return errors.New("segment IDs cannot be empty")
	}
	if fromID == toID {
		return errors.New("cannot add self-loop edge")
	}
	if distanceM < 0 {
		return errors.New("edge distance must be non-negative")
	}
	if _, ok := g.segments[fromID]; !ok {
		return fmt.Errorf("segment %q does not exist in graph", fromID)
	}
	if _, ok := g.segments[toID]; !ok {
		return fmt.Errorf("segment %q does not exist in graph", toID)
	}

	for _, e := range g.adj[fromID] {
		if e.To == toID {
			return nil
		}
	}

	g.adj[fromID] = append(g.adj[fromID], Edge{To: toID, Weight: distanceM})
	g.adj[toID] = append(g.adj[toID], Edge{To: fromID, Weight: distanceM})
	return nil
}

// AddConnection connects two segments using the average of their lengths as distance.
func (g *RoadGraph) AddConnection(fromID, toID string) error {
	segA, okA := g.segments[fromID]
	segB, okB := g.segments[toID]
	if !okA || !okB {
		return errors.New("both segments must exist in graph")
	}
	dist := (segA.LengthM + segB.LengthM) / 2.0
	return g.AddEdge(fromID, toID, dist)
}

// Segments returns all road segments in deterministic order.
func (g *RoadGraph) Segments() []RoadSegment {
	res := make([]RoadSegment, len(g.order))
	for i, id := range g.order {
		res[i] = *g.segments[id]
	}
	return res
}

// Segment retrieves a road segment by ID.
func (g *RoadGraph) Segment(id string) (*RoadSegment, bool) {
	seg, ok := g.segments[id]
	if !ok {
		return nil, false
	}
	return seg, true
}

// Priority queue item for Dijkstra
type pqItem struct {
	id    string
	dist  float64
	index int
}

type priorityQueue []*pqItem

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].dist < pq[j].dist }
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *priorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*pqItem)
	item.index = n
	*pq = append(*pq, item)
}
func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

// ShortestDistances computes shortest network path distances from sourceID to all reachable segments.
// If maxDistance > 0, exploration stops for nodes further than maxDistance.
func (g *RoadGraph) ShortestDistances(sourceID string, maxDistance float64) (map[string]float64, error) {
	if _, ok := g.segments[sourceID]; !ok {
		return nil, fmt.Errorf("source segment %q not found", sourceID)
	}

	dist := make(map[string]float64)
	dist[sourceID] = 0.0

	pq := &priorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &pqItem{id: sourceID, dist: 0.0})

	for pq.Len() > 0 {
		curr := heap.Pop(pq).(*pqItem)
		u := curr.id
		d := curr.dist

		if d > dist[u] {
			continue
		}
		if maxDistance > 0 && d > maxDistance {
			continue
		}

		for _, edge := range g.adj[u] {
			v := edge.To
			alt := d + edge.Weight
			if maxDistance > 0 && alt > maxDistance {
				continue
			}
			prevDist, exists := dist[v]
			if !exists || alt < prevDist {
				dist[v] = alt
				heap.Push(pq, &pqItem{id: v, dist: alt})
			}
		}
	}

	return dist, nil
}

// ShortestDistance returns the shortest network distance between two segments, or error if unreachable.
func (g *RoadGraph) ShortestDistance(fromID, toID string) (float64, error) {
	if fromID == toID {
		if _, ok := g.segments[fromID]; ok {
			return 0.0, nil
		}
		return math.Inf(1), fmt.Errorf("segment %q not found", fromID)
	}
	distances, err := g.ShortestDistances(fromID, 0)
	if err != nil {
		return math.Inf(1), err
	}
	d, ok := distances[toID]
	if !ok {
		return math.Inf(1), fmt.Errorf("segment %q unreachable from %q", toID, fromID)
	}
	return d, nil
}
