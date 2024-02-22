package main

import (
	"fmt"
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

var (
	m = [][]byte{
		{' ', '#', '#', '#', ' ', ' ', '#', '#'},
		{' ', '#', ' ', ' ', ' ', ' ', ' ', '#'},
		{'#', '#', ' ', ' ', ' ', '#', ' ', '#'},
		{'#', ' ', ' ', ' ', ' ', ' ', ' ', '#'},
		{'#', ' ', ' ', ' ', ' ', ' ', '#', '#'},
	}
	targetPos = []*pos{{2, 2}, {3, 2}, {3, 4}, {3, 5}}
	boxPos    = []*pos{{2, 2}, {3, 2}, {3, 3}, {3, 4}}
)

func checker(s *SokobanState) []Direction {
	s.refix()
	if !s.alive() {
		return nil
	}
	return []Direction{UP, DOWN, LEFT, RIGHT}
}
func processor(s *SokobanState, direct Direction) (state *SokobanState) {
	defer func() {
		if state != nil {
			state.Print()
		}
	}()
	return s.move(direct)
}

var _ State = new(SokobanState)

type SokobanState struct {
	// static
	size    *pos // bottom-right coordinates
	wall    [][]byte
	targets []*pos

	boxes  []*pos
	player *pos

	ttl int
	key string
}

func (s *SokobanState) Preprocess() error {
	if len(s.boxes) != len(s.targets) {
		return fmt.Errorf("targets num must equal to box num: %d != %d", len(s.targets), len(s.boxes))
	}
	sort.Slice(s.targets, func(i, j int) bool { return s.targets[i].x < s.targets[j].x || s.targets[i].y < s.targets[j].y })
	return nil
}

func (s *SokobanState) refix() *SokobanState {
	sort.Slice(s.boxes, func(i, j int) bool { return s.boxes[i].x < s.boxes[j].x || s.boxes[i].y < s.boxes[j].y })
	return s
}

func (s *SokobanState) alive() bool {
	return s.ttl > 0 && !s.boxInCorner() && !s.Done()
}

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
		nextState = s.next()
		nextState.player.move(direct, 1)
	case n == BoxChar && nn == BlankChar:
		nextState = s.next()
		nextState.player.move(direct, 1)
		nextState.getBox(nextState.player).move(direct, 1)
	}
	return nextState
}

func (s SokobanState) next() *SokobanState {
	boxes := make([]*pos, len(s.boxes))
	for i, box := range s.boxes {
		boxes[i] = box.duplicate()
	}
	s.boxes = boxes

	s.player = s.player.duplicate()

	s.key = ""
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
	boxes, targets := s.boxes, s.targets

	for i := range boxes {
		if !boxes[i].on(targets[i]) {
			return false
		}
	}
	return true
}

func (s *SokobanState) Key() string {
	if s == nil {
		return ""
	}
	if s.key != "" {
		return s.key
	}

	key := fmt.Sprint(s.player.x*(s.size.y+1) + s.player.y)
	for _, box := range s.boxes {
		key += "_" + fmt.Sprint(box.x*(s.size.y+1)+box.y)
	}
	s.key = key

	return key
}

func (s *SokobanState) Print() {
	var m [][]byte

	for i, row := range s.wall {
		m = append(m, make([]byte, len(row)))
		for j, c := range row {
			m[i][j] = c
		}
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

type pos struct {
	x int
	y int
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

func main() {
	finalStep, err := NewBruter(checker, processor, true).Find(&SokobanState{
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

		ttl: 1000,
	})
	if err != nil {
		fmt.Printf("pre process fail: %s", err)
		return
	}
	steps := finalStep.GetFullSteps()
	if len(steps) == 0 {
		fmt.Printf("no path found")
		return
	}

	fmt.Printf("cost %d steps\n", len(steps))
	for _, s := range steps {
		fmt.Printf("%c", s.Operate)
	}
	fmt.Println()
	for _, s := range steps {
		time.Sleep(300 * time.Millisecond)
		fmt.Printf("Operate: %c\n", s.Operate)
		s.State.Print()
	}
	fmt.Println()
}
