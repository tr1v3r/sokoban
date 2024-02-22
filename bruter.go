package main

import (
	"slices"
)

type State interface {
	Key() string
	Done() bool
	Preprocess() error
}

func NewStep[S State, O any](s S, lastStep *Step[S, O]) *Step[S, O] {
	if lastStep == nil {
		return &Step[S, O]{State: s, children: make(map[*O]*Step[S, O])}
	}
	return &Step[S, O]{
		State: s,

		cost:     lastStep.cost + 1,
		parent:   lastStep,
		children: make(map[*O]*Step[S, O]),
	}
}

type Step[S State, O any] struct {
	State S

	cost     int
	parent   *Step[S, O]
	children map[*O]*Step[S, O]
}

func (s *Step[S, O]) visited(key string) bool {
	return key == s.State.Key() || (s.parent != nil && s.parent.visited(key))
}

func (s *Step[S, O]) GetFullSteps() (steps []*Step[S, O], operations []O) {
	if s == nil {
		return nil, nil
	}

	for steps = append(steps, s); s.parent != nil; s = s.parent {
		steps = append(steps, s.parent)
	}
	slices.Reverse(steps)

	for i, step := range steps[1:] {
		for o, s := range steps[i].children {
			if s.State.Key() == step.State.Key() {
				operations = append(operations, *o)
				break
			}
		}
	}
	return steps, operations
}

func (s *Step[S, O]) RefixChildren() {
	for _, child := range s.children {
		if child.cost > s.cost+1 {
			child.parent = s
			child.cost = s.cost + 1
			child.RefixChildren()
		}
	}
}

func NewBruter[S State, O any](checker func(S) []O, processor func(S, O) S, findBest bool) *Bruter[S, O] {
	return &Bruter[S, O]{
		steps:    make(map[string]*Step[S, O]),
		findBest: findBest,
		check:    checker,
		process:  processor,
	}
}

type Bruter[S State, O any] struct {
	steps map[string]*Step[S, O]

	findBest bool

	check   func(S) []O  // get next operations
	process func(S, O) S // process state to next state
}

func (b Bruter[S, O]) Find(state S) (finalStep *Step[S, O], err error) {
	if err := state.Preprocess(); err != nil {
		return nil, err
	}
	return b.walkByStep(NewStep[S, O](state, nil)), nil
}

func (b Bruter[S, O]) walkByStep(s *Step[S, O]) (finalStep *Step[S, O]) {
	var finalSteps []*Step[S, O]
	for _, o := range b.check(s.State) {
		o := o
		// TODO check nextStep == nil
		nextState := b.process(s.State, o)

		key := nextState.Key()
		if key == "" {
			continue
		}

		if s.visited(key) {
			continue
		}

		var nextStep *Step[S, O]
		if nextStep = b.steps[key]; nextStep != nil {
			s.children[&o] = nextStep
			if s.cost+1 >= nextStep.cost {
				continue
			}
			s.RefixChildren()
		} else {
			nextStep = NewStep(nextState, s)
			s.children[&o] = nextStep
			b.steps[key] = nextStep
		}

		if nextState.Done() {
			return nextStep
		}

		if step := b.walkByStep(nextStep); step != nil {
			if !b.findBest {
				return step
			}
			finalSteps = append(finalSteps, step)
		}
	}

	for _, s := range finalSteps {
		if finalStep == nil || finalStep.cost > s.cost {
			finalStep = s
		}
	}
	return finalStep
}
