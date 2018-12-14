// Package dmap implements a Brogue-style Dijkstra Map data structure
// based on the map it is given. For more information on this data
// structure, read the article on RogueBasin:
// http://www.roguebasin.com/index.php?title=The_Incredible_Power_of_Dijkstra_Maps
package dmap

import (
	"bytes"
	"fmt"
	"math"
)

// Point is a representation of a point on a map
type Point interface {
	GetXY() (int, int)
}

// Map is a map to calculate dmaps with. All methods should be linear
// time, as the algorithm will call them a lot. Also your map should
// be statically sized; if the map is dynamically sized, a new dmap
// will need to be created for it whenever it changes size.
type Map interface {
	SizeX() int
	SizeY() int
	IsPassable(int, int) bool
	OOB(int, int) bool
}

// Rank is the rank of a tile - lower is closer to the target
type Rank uint16

// RankMax is the highest possible rank. It takes a value a little
// below its implementations maximum to prevent overflow
const RankMax = math.MaxUint16 - 10

// DijkstraMap is a representation of a Brogue-style 'Dijkstra'
// map. To reach a target, an AI should try to minimize the rank of
// the tile it's standing on (targets have a value of zero)
type DijkstraMap struct {
	Points       [][]Rank
	M            Map
	NeigbourFunc func(d *DijkstraMap, x, y int) []WeightedPoint
}

// WeightedPoint is a Point that also has a rank
type WeightedPoint struct {
	X   int
	Y   int
	Val Rank
}

// GetXY implements the Point interface
func (d *WeightedPoint) GetXY() (int, int) {
	return d.X, d.Y
}

// BlankDMap creates a blank Dijkstra map to be used with the map passed to it
func BlankDMap(m Map, neigbourfunc func(d *DijkstraMap, x, y int) []WeightedPoint) *DijkstraMap {
	ret := make([][]Rank, m.SizeX())
	for i := range ret {
		ret[i] = make([]Rank, m.SizeY())
		for j := range ret[i] {
			ret[i][j] = RankMax
		}
	}
	return &DijkstraMap{ret, m, neigbourfunc}
}

// ManhattanNeighbours returns the neighbours of the block x, y to the
// north, south, east, and west
func ManhattanNeighbours(d *DijkstraMap, x, y int) []WeightedPoint {
	return []WeightedPoint{
		d.GetValPoint(x+1, y),
		d.GetValPoint(x-1, y),
		d.GetValPoint(x, y-1),
		d.GetValPoint(x, y+1),
	}
}

// DiagonalNeighbours returns the neighbours of the block x, y to the
// north, south, east, west, NE, SE, NW, and SW
func DiagonalNeighbours(d *DijkstraMap, x, y int) []WeightedPoint {
	return []WeightedPoint{
		d.GetValPoint(x+1, y),
		d.GetValPoint(x-1, y),
		d.GetValPoint(x, y-1),
		d.GetValPoint(x, y+1),
		d.GetValPoint(x+1, y+1),
		d.GetValPoint(x+1, y-1),
		d.GetValPoint(x-1, y+1),
		d.GetValPoint(x-1, y-1),
	}
}

// Calc calculates the Dijkstra map with points given as targets. You
// need to blank the map before using this method. It's recommended to
// use this one initially, but to use a Recalc instead for subsequent
// moves, since Recalc, unlike BlankDMap, doesn't allocate memory.
func (d *DijkstraMap) Calc(points ...Point) {
	for _, point := range points {
		x, y := point.GetXY()
		d.Points[x][y] = 0
	}
	mademutation := true
	for mademutation {
		mademutation = false
		for x := range d.Points {
			for y := range d.Points[x] {
				if d.M.IsPassable(x, y) {
					ln := d.LowestNeighbour(x, y).Val
					if d.Points[x][y] > ln+1 {
						d.Points[x][y] = ln + 1
						mademutation = true
					}
				}
				x1, y1 := (d.M.SizeX()-1)-x, (d.M.SizeY()-1)-y
				if d.M.IsPassable(x1, y1) {
					ln := d.LowestNeighbour(x1, y1).Val
					if d.Points[x1][y1] > ln+1 {
						d.Points[x1][y1] = ln + 1
						mademutation = true
					}
				}
			}
		}
	}
}

// Recalc recalculates the Dijkstra map with points given as
// targets. It's essentially equivalent to a blank followed by a calc,
// but should be a bit faster because it doesn't reallocate the
// memory. As per the note for DijkstraMap, don't use this method if
// your map is dynamically sized; you'll just have to use BlankDMap
// and Calc as if creating a new dmap every update.
func (d *DijkstraMap) Recalc(points ...Point) {
	for i := range d.Points {
		for j := range d.Points[i] {
			d.Points[i][j] = RankMax
		}
	}
	d.Calc(points...)
}

// GetValPoint gets the weighted point at X, Y of the Dijkstra
// map. Points that are out of bounds count as maximum rank (so
// shouldn't be targeted)
func (d *DijkstraMap) GetValPoint(x, y int) WeightedPoint {
	if d.M.OOB(x, y) {
		return WeightedPoint{x, y, RankMax}
	}
	return WeightedPoint{x, y, d.Points[x][y]}
}

// LowestNeighbour returns the neighbour of the point at x, y with the
// lowest rank.
func (d *DijkstraMap) LowestNeighbour(x, y int) WeightedPoint {
	vals := d.NeigbourFunc(d, x, y)
	var lv Rank = RankMax
	ret := vals[0]
	for _, val := range vals {
		if val.Val < lv {
			lv = val.Val
			ret = val
		}
	}
	return ret
}

// String returns a string representation of a Dijkstra Map
func (d *DijkstraMap) String() string {
	buf := bytes.Buffer{}
	for x := range d.Points {
		for y := range d.Points[x] {
			buf.WriteString(fmt.Sprintf("%6d", d.Points[x][y]))
			buf.WriteString(", ")
		}
		buf.WriteRune('\n')
	}
	return buf.String()
}
