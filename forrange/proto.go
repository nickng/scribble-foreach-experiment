package forrange

import (
	"fmt"
	"log"
	"os"
)

type resource struct {
	used bool
}

func (res *resource) Use() {
	if res.used {
		log.Fatal("Resource used")
		os.Exit(1)
	}
	res.used = true
}

// ProtoParam is just a lookup table for protocol parameters
var ProtoParam = map[string]int{
	"k": 2, // role A(k=2)
}

// Example protocol - nested one-to-many:
//
// foreach A[i:1..k] {
//   foreach A[j:1..k] {
//     foo(int) to A[j];
//   }
//   bar(string) to A[i];
// }
//

// foreach outer → body outer → foreach inner → body inner → goto foreach inner
//               ↘ exit outer;                ↘ exit inner → goto foreach outer;
//
//

// foreachState and foreachStack is common code in all generated code --------
// body should be present for any generated API that uses foreach

// foreachState keeps track of foreach loop index in code.
type foreachState struct {
	ID         int // ID is the foreach ID
	curr, last int // curr, last is the current, last index of the foreach
}

func (state *foreachState) bodyOK() bool  { return state.curr <= state.last }
func (state *foreachState) hasNext() bool { return state.curr < state.last }

func (state *foreachState) String() string {
	return fmt.Sprintf("{%d: %d/%d}", state.ID, state.curr, state.last)
}

// foreachStack is a datastructure shared between the states.
type foreachStack struct {
	index int
	stack []*foreachState
}

func (fes *foreachStack) top() *foreachState {
	if fes.stack != nil {
		return fes.stack[fes.index]
	}
	return nil
}

func (fes *foreachStack) push(ID, rangeLen int) {
	// Pre: rangeLen > 0
	newForeach := &foreachState{ID: ID, curr: 0, last: rangeLen - 1}
	if fes.stack == nil {
		fes.index = 0 // it should already be 0
		fes.stack = []*foreachState{newForeach}
		return
	}
	fes.index++
	fes.stack = append(fes.stack, newForeach)
}

func (fes *foreachStack) pop() {
	stacksize := len(fes.stack)
	if stacksize > 0 {
		fes.stack = fes.stack[:stacksize-1]
		fes.index = stacksize - 2 // could be -1
	} else {
		log.Fatalf("Cannot pop empty foreach-stack")
	}
}

func (fes foreachStack) isEmpty() bool {
	return fes.stack == nil || len(fes.stack) == 0
}

func (fes foreachStack) String() string {
	return fmt.Sprintf("stack %v@%d", fes.stack, fes.index)
}

// fes is the shared foreach stack.
var fes = new(foreachStack)

/// --- end of common code ---------------------------------------------------

// S0 is the initial state.
// It is also the outer foreach init state.
//
// Every foreach init state contains four methods only
// - ID()                                          == foreach unique ID
// - Foreach(body func(*bodyBegin) *bodyEnd) *exit == Enter loop body, and call
// This may be optimised out BUT the bound index check should be here
// Note: I plan to factor out these later and embed in state object.
//
type S0 struct {
	resource
}

func (s *S0) ID() int { return 0 } // Generate as method returning constant so immutable

// Foreach moves s0 → s1, where s1 is the body of outer foreach i.e. inner foreach
func (s *S0) Foreach() (<-chan *S1, *SEnd) {
	s.Use()
	if fes.isEmpty() || fes.top().ID != s.ID() {
		// first time enter loop
		fes.push(s.ID(), ProtoParam["k"]) // where k is the param of foreach
	} else if fes.top().ID == s.ID() {
		// re-enter loop
		fes.top().curr++
	} else {
		panic("shouldn't get here")
	}

	ch := make(chan *S1, 1)
	if fes.top().bodyOK() {
		ch <- &S1{foreach: ch}
	}
	return ch, new(SEnd)
}

// S1 is the inner foreach init state.
type S1 struct {
	resource
	foreach chan *S1
}

func (s *S1) ID() int { return 1 } // Generate as method returning constant so immutable

// Foreach moves s1 → s2, where s2 is the body of inner foreach
func (s *S1) Foreach() (<-chan *S2, *S3) {
	s.Use()
	if fes.isEmpty() || fes.top().ID != s.ID() {
		// first time enter loop
		fes.push(s.ID(), ProtoParam["k"]) // where k is the param of foreach
	} else if fes.top().ID == s.ID() {
		// re-enter loop
		fes.top().curr++
	} else {
		panic("shouldn't get here")
	}

	ch := make(chan *S2, 1)
	if fes.top().bodyOK() {
		ch <- &S2{foreach: ch}
	}
	return ch, &S3{foreach: s.foreach}
}

// S2 is the body of inner foreach.
// It is the first statement of inner foreach.
type S2 struct {
	resource
	foreach chan *S2
}

// Send_Aj_foo is first and last method of inner foreach body.
// As the last statement of inner foreach, always go back to s1 (inner foreach init).
func (s *S2) Send_Aj_foo(v int) *S5 {
	s.Use()
	return &S5{foreach: s.foreach}
}

// S3 is in the body of outer foreach (statement after inner foreach)
type S3 struct {
	resource
	foreach chan *S1
}

// Send_Ai_bar is the last method of outer foreach body.
// As the last statement of outer foreach, always go back to s0 (outer foreach init).
func (s *S3) Send_Ai_bar(v string) *S4 {
	s.Use()
	return &S4{foreach: s.foreach}
}

// S4 is the ending state of the outer foreach loop.
type S4 struct {
	resource
	foreach chan *S1
}

func (s *S4) End() {
	s.Use()
	fes.top().curr++
	if fes.top().bodyOK() {
		s.foreach <- &S1{foreach: s.foreach}
	} else {
		fes.pop()
		close(s.foreach)
	}
}

// S5 is the ending state of the inner foreach loop.
type S5 struct {
	resource
	foreach chan *S2
}

func (s *S5) End() {
	s.Use()
	fes.top().curr++
	if fes.top().bodyOK() {
		s.foreach <- &S2{foreach: s.foreach}
	} else {
		fes.pop()
		close(s.foreach)
	}
}

// SEnd is the usual final state of a protocol.
type SEnd struct {
	resource
}

func (s *SEnd) End() {
	s.Use()
	// Close channels etc.
}
