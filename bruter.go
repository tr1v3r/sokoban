package main

import (
	"slices"
)

type State interface {
	Key() string
	Done() bool
	Preprocess() error
}

type StepArray[S State, O any] []*Step[S, O]

type Step[S State, O any] struct {
	State   S
	Operate O

	cost     int
	parent   *Step[S, O]
	childern StepArray[S, O]
}

func (s *Step[S, O]) visited(key string) bool {
	return key == s.State.Key() || (s.parent != nil && s.parent.visited(key))
}

func (s *Step[S, O]) GetFullSteps() (steps StepArray[S, O]) {
	if s == nil {
		return nil
	}

	for steps = append(steps, s); s.parent != nil; s = s.parent {
		steps = append(steps, s.parent)
	}
	slices.Reverse(steps)
	return steps
}

func (s *Step[S, O]) Graft(parent *Step[S, O], operate O) {
	s.parent = parent
	s.Operate = operate
	s.RefixCost(parent.cost + 1)
}

func (s *Step[S, O]) RefixCost(cost int) {
	if s.cost <= cost {
		return
	}
	s.cost = cost
	for _, child := range s.childern {
		child.RefixCost(s.cost + 1)
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
	step := b.walkByStep(&Step[S, O]{State: state})

	steps := step.GetFullSteps()
	_ = steps
	return step, nil
}

func (b Bruter[S, O]) walkByStep(s *Step[S, O]) (finalStep *Step[S, O]) {
	for _, o := range b.check(s.State) {
		nextState := b.process(s.State, o)

		key := nextState.Key()
		if key == "" {
			continue
		}

		if s.visited(key) {
			continue
		}

		if brutedState := b.steps[key]; brutedState != nil {
			if brutedState.cost > s.cost+1 {
				s.childern = append(s.childern, brutedState)
				brutedState.Graft(s, o)
			}
			continue
		}

		nextStep := &Step[S, O]{
			State:   nextState,
			Operate: o,

			cost:   s.cost + 1,
			parent: s,
		}
		b.steps[key] = nextStep
		s.childern = append(s.childern, nextStep)

		if nextState.Done() {
			return nextStep
		}

		if step := b.walkByStep(nextStep); step == nil {
			continue
		} else if !b.findBest {
			return step
		} else if finalStep == nil || finalStep.cost > step.cost {
			finalStep = step
		}
	}
	return finalStep
}
