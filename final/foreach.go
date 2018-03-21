// Package final is the final design of the foreach API.
//
// Generated state code
//
// To use foreach tracking, first generate a Foreach method in the API:
//
//     // sstack is a stack to keep track of nested states.
//     // note: this should not be a package-level variable,
//     // and should be a field of the protocol instance if possible
//     //
//     var sstack = newStack()
//
//     // S is the initial foreach state.
//     type S struct { ... }
//
//     func (s *S) ID() int { return 1 } // GENERATE THIS unique constant id
//
//     // Foreach function takes as parameter a function that implements
//     // the loop body where:
//     //  - SbodyStart is the first state in the loop body
//     //  - SbodyEnd is the last state in the loop body
//     //  - SloopExit is the state after exiting the loop
//     //
//     func (s *S) Foreach(bodyFn(*SbodyStart) *SbodyEnd) *SloopExit {
//     	s.Use() // use resource
//     	if sstack.isEmpty() || sstack.top().ID != s.ID() {
//     		// first time enter loop
//     		sstack.push(s.ID(), ProtoParam["k"]) // GENERATE THIS k is the param
//     	} else if sstack.top().ID == s.ID() {
//     		// re-enter loop
//     		sstack.top().increment()
//     	} else {
//     		panic("shouldn't get here")
//     	}
//
//     	for sstack.top().canEnter() {
//     		// Run the body then run the internal end() method.
//     		body(new(SbodyStart)).end()
//     	}
//     	sstack.pop()
//     	return &SloopExit{ ... }
//     }
//
//     // SbodyEnd is the special loop body end state.
//     type SbodyEnd struct { ... }
//
//     // end is the unexported internal method for keeping track of loop index.
//     func (s *SbodyEnd) end() {
//     	s.Use() // use resource
//     	sstack.top().increment()
//     }
//
// Nested FSM tracking
//
// Simply copy the content of this source code file to the API.
// This file implements a stack for tracking foreach executions.
//
package final

// This file contains common code for nested FSM tracking.

import (
	"errors"
	"fmt"
	"log"
)

var (
	ErrPopEmptyStack = errors.New("cannot pop: stack empty")
)

// foreachState keeps track of foreach loop index in code.
type foreachState struct {
	ID         int // ID is the unique sub-FSM ID
	curr, last int // curr/last is the current/last index of the foreach
}

// canEnter returns true if the current index is within foreach bounds.
func (state *foreachState) canEnter() bool {
	return state.curr <= state.last
}

func (state *foreachState) increment() { state.curr++ }

func (state *foreachState) String() string {
	return fmt.Sprintf("{%d: %d/%d}", state.ID, state.curr, state.last)
}

// fsmStack is a data structure shared between the states.
type fsmStack struct {
	index int             // top of stack, i.e. the current stack element
	stack []*foreachState // the stack of states
}

func newStack() *fsmStack {
	return new(fsmStack)
}

func (s *fsmStack) top() *foreachState {
	if s.stack != nil {
		return s.stack[s.index]
	}
	return nil
}

func (s *fsmStack) push(ID, rangeLen int) {
	// Pre: rangeLen > 0
	newForeach := foreachState{
		ID:   ID,
		curr: 0,
		last: rangeLen - 1,
	}
	s.stack = append(s.stack, &newForeach)
	s.index = len(s.stack) - 1 // cached top index
}

func (s *fsmStack) pop() {
	stacksize := len(s.stack) // don't rely on s.index
	if stacksize > 0 {
		s.stack, s.index = s.stack[:stacksize-1], stacksize-2
	} else {
		log.Fatal(ErrPopEmptyStack)
	}
}

func (s fsmStack) isEmpty() bool {
	return s.stack == nil || len(s.stack) == 0
}

func (s fsmStack) String() string {
	return fmt.Sprintf("stack %v@%d", s.stack, s.index)
}
