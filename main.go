package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/thoas/go-funk"
)

const ROW, COL = 5, 8

var (
	m maze = [ROW][COL]byte{
		{' ', '#', '#', '#', ' ', ' ', '#', '#'},
		{' ', '#', ' ', ' ', ' ', ' ', ' ', '#'},
		{'#', '#', ' ', ' ', ' ', '#', ' ', '#'},
		{'#', ' ', ' ', ' ', ' ', ' ', ' ', '#'},
		{'#', ' ', ' ', ' ', ' ', ' ', '#', '#'},
	}
	targetPos = []*pos{{2, 2}, {3, 2}, {3, 4}, {3, 5}}
	boxPos    = []*pos{{2, 2}, {3, 2}, {3, 3}, {3, 4}}
)

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
func (p pos) jump(direction byte, stepSize int) *pos {
	switch direction {
	case 'u':
		p.x -= stepSize
	case 'd':
		p.x += stepSize
	case 'l':
		p.y -= stepSize
	case 'r':
		p.y += stepSize
	}
	return &p
}

type maze [ROW][COL]byte

func (m maze) render(cur *pos, targets []*pos, boxes []*pos) *maze {
	for _, p := range targets {
		m[p.x][p.y] = 'x'
	}
	for _, p := range boxes {
		if m[p.x][p.y] == 'x' {
			m[p.x][p.y] = '@'
		} else {
			m[p.x][p.y] = 'O'
		}
	}
	if cur != nil {
		m[cur.x][cur.y] = '*'
	}
	return &m
}

func (m *maze) print() {
	fmt.Println("-------------------------")
	for _, row := range m {
		for _, col := range row {
			fmt.Printf("%c ", col)
		}
		fmt.Println()
	}
	fmt.Println("-------------------------")
}

func (m *maze) getPos(p *pos) byte {
	if p.x < 0 || p.x > len(m)-1 || p.y < 0 || p.y > len(m[0])-1 {
		return 0
	}
	return m[p.x][p.y]
}

type step struct {
	direction string
	pos       pos
	boxPos    []*pos
}

func (s *step) inCorner() bool {
	for _, box := range s.boxPos {
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

func (s *step) finish() bool { return disorderMatch(s.boxPos, targetPos) }
func (s *step) key() string {
	k := fmt.Sprintf("%d_%d", s.pos.x, s.pos.y)
	var boxIndexes []int
	for _, box := range s.boxPos {
		boxIndexes = append(boxIndexes, box.x*COL+box.y)
	}
	sort.Ints(boxIndexes)
	for _, index := range boxIndexes {
		k += "_" + fmt.Sprint(index)
	}
	return k
}

func WalkFrom(start pos, boxPos ...*pos) *Walker {
	return &Walker{
		steps: []step{{
			pos:    start,
			boxPos: boxPos,
		}},
		deadPaths: make(map[string]bool, 1024),
	}
}

type Walker struct {
	steps     []step
	records   []string
	deadPaths map[string]bool
}

func (w *Walker) walk(ttl int) *Walker {
	if w == nil {
		return nil
	}
	if ttl <= 0 {
		return nil
	}

	s := w.lastStep()

	if s.inCorner() { // check box in dead position
		return nil
	}
	if s.finish() { // check if game is finished
		return w
	}
	key := s.key()
	if w.deadPaths[key] || funk.InStrings(w.records, key) { // check deadpath and loop
		return nil
	} else {
		w.records = append(w.records, key)
	}

	// fmt.Printf("ttl: %d  key: %s\n", ttl, key)
	// m.render(&s.pos, targetPos, s.boxPos).print()

	ttl--

	if next := w.move(s, 'u').walk(ttl); next != nil {
		return next
	}
	if next := w.move(s, 'd').walk(ttl); next != nil {
		return next
	}
	if next := w.move(s, 'l').walk(ttl); next != nil {
		return next
	}
	if next := w.move(s, 'r').walk(ttl); next != nil {
		return next
	}

	w.deadPaths[s.key()] = true

	return nil
}

func (w Walker) move(s *step, direction byte) *Walker {
	n, nn := w.getNext2Pos(s, direction)
	if !w.canMove(n, nn) {
		return nil
	}

	newStep := step{direction: string(direction), pos: *s.pos.jump(direction, 1)}

	if n == ' ' {
		newStep.boxPos = s.boxPos
	} else if n == 'O' {
		for _, box := range s.boxPos {
			if box.on(&newStep.pos) {
				newStep.boxPos = append(newStep.boxPos, box.jump(direction, 1))
			} else {
				newStep.boxPos = append(newStep.boxPos, box)
			}
		}
	}
	w.steps = append(w.steps, newStep)
	return &w
}

func (*Walker) canMove(n, nn byte) bool {
	if n == 0 || n == '#' {
		return false
	}
	if n == ' ' {
		return true
	}

	// n must be box 'O'
	if nn == '#' || nn == 'O' || nn == 0 {
		return false
	}
	return true
}

func (w *Walker) getNext2Pos(s *step, direction byte) (next, nextNext byte) {
	p, m := s.pos, m.render(nil, nil, s.boxPos)
	return m.getPos(p.jump(direction, 1)), m.getPos(p.jump(direction, 2))
}

func (w *Walker) lastStep() *step { return &w.steps[len(w.steps)-1] }

func main() {
	cur := pos{4, 3}

	m.render(&cur, targetPos, boxPos).print()

	result := WalkFrom(cur, boxPos...).walk(1e3)
	if result == nil {
		fmt.Println("path not found")
		return
	}
	for _, s := range result.steps {
		m.render(&s.pos, targetPos, s.boxPos).print()
		time.Sleep(300 * time.Millisecond)
	}
}

func disorderMatch(a, b []*pos) bool {
	for _, s := range a {
		if !s.in(b...) {
			return false
		}
	}
	return true
}
