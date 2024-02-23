package main

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type Direction byte

const (
	UP    Direction = 'u'
	DOWN  Direction = 'd'
	LEFT  Direction = 'l'
	RIGHT Direction = 'r'

	OutRangeChar         = 0
	BlankChar       byte = ' '
	WallChar        byte = '#'
	BoxChar         byte = 'X'
	TargetChar      byte = 'O'
	BoxOnTargetChar byte = '@'
	PlayerChar      byte = '*'
)

func direct(s *SokobanState) []Direction { return []Direction{UP, DOWN, LEFT, RIGHT} }
func process(s *SokobanState) (states []*SokobanState) {
	if !s.alive() {
		return nil
	}

	for _, d := range direct(s) {
		if state := s.move(d); state != nil {
			states = append(states, state.refix())
		}
	}

	var targetDistances, boxDistances = make([]float64, len(states)), make([]float64, len(states))
	for i, s := range states {
		targetDistances[i], boxDistances[i] = s.targetDistance(), s.boxDistance()
	}
	sort.Slice(states, func(i, j int) bool {
		if targetDistances[i] == targetDistances[j] {
			return boxDistances[i] < boxDistances[j]
		}
		return targetDistances[i] < targetDistances[j]
	})

	return states
}

func countDistance(from *pos, to ...*pos) (d float64) {
	for _, t := range to {
		d += from.distance(t)
	}
	return d
}

// ############## state ############

var _ State = new(SokobanState)

type SokobanState struct {
	// static
	size    *pos // bottom-right coordinates
	wall    [][]byte
	targets []*pos

	boxes  []*pos
	player *pos

	by Direction

	ttl int
	key string
}

func (s *SokobanState) Preprocess() error {
	if s == nil {
		return fmt.Errorf("nil state")
	}
	if len(s.boxes) != len(s.targets) {
		return fmt.Errorf("targets num must equal to box num: %d != %d", len(s.targets), len(s.boxes))
	}
	sort.Slice(s.targets, func(i, j int) bool {
		return s.targets[i].x*(s.size.y+1)+s.targets[i].y < s.targets[j].x*(s.size.y+1)+s.targets[j].y
	})
	s.refix()
	return nil
}

func (s *SokobanState) refix() *SokobanState {
	sort.Slice(s.boxes, func(i, j int) bool {
		return s.boxes[i].x*(s.size.y+1)+s.boxes[i].y < s.boxes[j].x*(s.size.y+1)+s.boxes[j].y
	})

	if s.key == "" {
		s.key += fmt.Sprint(s.player.x*(s.size.y+1) + s.player.y)
		for _, box := range s.boxes {
			s.key += "_" + fmt.Sprint(box.x*(s.size.y+1)+box.y)
		}
	}
	return s
}

func (s *SokobanState) targetDistance() (distance float64) {
	for i, b := range s.boxes {
		distance += s.targets[i].distance(b)
	}
	return distance
}
func (s *SokobanState) boxDistance() (distance float64) {
	for _, b := range s.boxes {
		distance += s.player.distance(b)
	}
	return distance
}

func (s *SokobanState) alive() bool { return s.ttl > 0 && !s.boxInCorner() }

func (s *SokobanState) boxInCorner() bool {
	for _, box := range s.boxes {
		if box.in([]*pos{
			{0, 4}, {0, 5},
			{1, 2}, {1, 6},
			{3, 1}, {3, 6},
			{4, 1}, {4, 2}, {4, 3}, {4, 4}, {4, 5},
		}...) {
			return true
		}
	}
	return false
}

func (s *SokobanState) move(direct Direction) (nextState *SokobanState) {
	n := s.getPos(s.player.jump(direct, 1))
	nn := s.getPos(s.player.jump(direct, 2))

	switch {
	case n == BlankChar:
		nextState = s.next(direct)
		nextState.player.move(direct, 1)
	case n == BoxChar && nn == BlankChar:
		nextState = s.next(direct)
		nextState.player.move(direct, 1)
		nextState.getBox(nextState.player).move(direct, 1)
	}
	return nextState
}

func (s SokobanState) next(by Direction) *SokobanState {
	boxes := make([]*pos, len(s.boxes))
	for i, box := range s.boxes {
		boxes[i] = box.duplicate()
	}
	s.boxes = boxes

	s.player = s.player.duplicate()

	s.key = ""
	s.by = by
	s.ttl--

	return &s
}

func (s *SokobanState) getBox(p *pos) *pos {
	for _, box := range s.boxes {
		if box.on(p) {
			return box
		}
	}
	return nil
}

func (s *SokobanState) Done() bool {
	for i := range s.boxes {
		if !s.boxes[i].on(s.targets[i]) {
			return false
		}
	}
	return true
}

func (s *SokobanState) Key() string { return s.key }

func (s *SokobanState) Print() {
	var m [][]byte

	for i, row := range s.wall {
		m = append(m, make([]byte, len(row)))
		copy(m[i], row)
	}

	for _, p := range s.targets {
		m[p.x][p.y] = TargetChar
	}
	for _, p := range s.boxes {
		if m[p.x][p.y] == TargetChar {
			m[p.x][p.y] = BoxOnTargetChar
		} else {
			m[p.x][p.y] = BoxChar
		}
	}
	if s.player != nil {
		m[s.player.x][s.player.y] = PlayerChar
	}

	for _, row := range m {
		for _, col := range row {
			fmt.Printf("%c ", col)
		}
		fmt.Println()
	}
	fmt.Println("-------------------------")
}

func (s *SokobanState) getPos(p *pos) byte {
	if p.x < 0 || p.x > s.size.x || p.y < 0 || p.y > s.size.y {
		return OutRangeChar
	}
	for _, box := range s.boxes {
		if box.on(p) {
			return BoxChar
		}
	}
	return s.wall[p.x][p.y]
}

// ############## pos ############

type pos struct {
	x int
	y int
}

func (p *pos) distance(t *pos) float64 {
	return math.Abs(float64(p.x-t.x)) + math.Abs(float64(p.y-t.y))
}
func (p *pos) on(t *pos) bool { return p.x == t.x && p.y == t.y }
func (p *pos) in(ts ...*pos) bool {
	for _, t := range ts {
		if p.on(t) {
			return true
		}
	}
	return false
}
func (p *pos) move(direct Direction, stepSize int) *pos {
	switch direct {
	case UP:
		p.x -= stepSize
	case DOWN:
		p.x += stepSize
	case LEFT:
		p.y -= stepSize
	case RIGHT:
		p.y += stepSize
	}
	return p
}
func (p pos) jump(direct Direction, stepSize int) *pos { return p.move(direct, stepSize) }
func (p pos) duplicate() *pos                          { return &p }

// ############## main ############

func main() {
	start := time.Now()
	finalStep, err := NewBruter(process).Find(&SokobanState{
		size: &pos{4, 7},
		wall: [][]byte{
			{' ', '#', '#', '#', ' ', ' ', '#', '#'},
			{' ', '#', ' ', ' ', ' ', ' ', ' ', '#'},
			{'#', '#', ' ', ' ', ' ', '#', ' ', '#'},
			{'#', ' ', ' ', ' ', ' ', ' ', ' ', '#'},
			{'#', ' ', ' ', ' ', ' ', ' ', '#', '#'},
		},
		targets: []*pos{{2, 2}, {3, 2}, {3, 4}, {3, 5}},
		boxes:   []*pos{{2, 2}, {3, 2}, {3, 3}, {3, 4}},
		player:  &pos{4, 3},

		ttl: 10000,
	}, "bfs")
	if err != nil {
		fmt.Printf("pre process fail: %s", err)
		return
	}
	fmt.Printf("find path cost: %s\n", time.Since(start))

	steps := finalStep.Backtrack()
	if len(steps) == 0 {
		fmt.Printf("no path found")
		return
	}

	fmt.Printf("cost %d steps\n", len(steps)-1)
	for _, s := range steps {
		fmt.Print(string(s.State.by))
	}
	fmt.Println()

	for _, s := range steps {
		time.Sleep(300 * time.Millisecond)
		fmt.Printf("step: %d\n", s.cost)
		s.State.Print()
	}
}
