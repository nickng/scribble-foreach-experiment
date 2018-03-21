package final

import (
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

// sstack is the shared foreach stack.
var sstack = newStack()

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

// Foreach moves S0 → S1, where S1 is the body of outer foreach i.e. inner foreach
func (s *S0) Foreach(bodyFn func(*S1) *S4) *SEnd {
	s.Use()
	if sstack.isEmpty() || sstack.top().ID != s.ID() {
		// first time enter loop
		sstack.push(s.ID(), ProtoParam["k"]) // where k is the param of foreach
	} else if sstack.top().ID == s.ID() {
		// re-enter loop
		sstack.top().increment()
	} else {
		panic("shouldn't get here")
	}

	// Run at least once (range is never empty)
	for sstack.top().canEnter() {
		bodyFn(new(S1)).end()
	}
	sstack.pop()
	return new(SEnd)
}

// S4 is the ending state of the outer foreach loop.
type S4 struct {
	resource
}

func (s *S4) end() {
	s.Use()
	sstack.top().increment()
}

// S1 is the inner foreach init state.
type S1 struct {
	resource
}

func (s *S1) ID() int { return 1 } // Generate as method returning constant so immutable

// Foreach moves s1 → s2, where s2 is the body of inner foreach
func (s *S1) Foreach(bodyFn func(*S2) *S5) *S3 {
	s.Use()
	if sstack.isEmpty() || sstack.top().ID != s.ID() {
		// first time enter loop
		sstack.push(s.ID(), ProtoParam["k"]) // where k is the param of foreach
	} else if sstack.top().ID == s.ID() {
		// re-enter loop
		sstack.top().increment()
	} else {
		panic("shouldn't get here")
	}

	// Run at least once (range is never empty)
	for sstack.top().canEnter() {
		bodyFn(new(S2)).end()
	}
	sstack.pop()
	return new(S3)
}

// S2 is the body of inner foreach.
// It is the first statement of inner foreach.
type S2 struct {
	resource
}

// Send_Aj_foo is first and last method of inner foreach body.
// As the last statement of inner foreach, always go back to s1 (inner foreach init).
func (s *S2) Send_Aj_foo(v int) *S5 {
	s.Use()
	return new(S5)
}

// S3 is in the body of outer foreach (statement after inner foreach)
type S3 struct {
	resource
}

// Send_Ai_bar is the last method of outer foreach body.
// As the last statement of outer foreach, always go back to s0 (outer foreach init).
func (s *S3) Send_Ai_bar(v string) *S4 {
	s.Use()
	return new(S4)
}

// S5 is the ending state of the inner foreach loop.
type S5 struct {
	resource
}

func (s *S5) end() {
	s.Use()
	sstack.top().increment()
}

// SEnd is the usual final state of a protocol.
type SEnd struct {
	resource
}

func (s *SEnd) End() {
	s.Use()
	// Close channels etc.
}
