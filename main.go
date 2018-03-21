package main

import (
	"fmt"
	"log"

	"github.com/nickng/scribble-foreach-experiment/forrange"
	"github.com/nickng/scribble-foreach-experiment/fused"
	"github.com/nickng/scribble-foreach-experiment/nested"
	"github.com/nickng/scribble-foreach-experiment/proto"
	"github.com/nickng/scribble-foreach-experiment/recur"
)

// Example protocol - nested one-to-many:
//
// foreach A[i:1..k] {
//   foreach A[j:1..k] {
//     foo(int) to A[j];
//   }
//   bar(string) to A[i];
// }
//

func init() {
	log.SetPrefix("proto: ")
	log.SetFlags(log.Llongfile)
}

func main() {
	forrange.ProtoParam["k"] = 2
	fmt.Println("---- for-range ----")
	forrangeRun()

	nested.ProtoParam["k"] = 2
	fmt.Println("---- nested FSM ----")
	nestedRun()

	recur.ProtoParam["k"] = 2
	fmt.Println("---- recur foreach ----")
	recurRun()
	fmt.Println("---- recur foreach (inline) ----")
	recurInlineRun()

	proto.ProtoParam["k"] = 2
	fmt.Println("---- GOOD ----")
	protoGood()
	fmt.Println("---- BAD ----")
	//protoBad()

	fused.ProtoParam["k"] = 2
	fmt.Println("---- fused foreach ----")
	fusedRun() // fails linear usage check
}

func protoGood() {
	// This function is the good use of foreach

	s := new(proto.S0)
	j := 0
	for s.HasNext() {
		j++
		s1 := s.Foreach()
		fmt.Println("Outer loop", j)
		i := 0
		for s1.HasNext() {
			i++
			fmt.Println("Inner loop", i)
			s1 = s1.Foreach().
				Send_Aj_foo(i)
			fmt.Println("Inner loop end", i)
		}
		fmt.Println("End inner foreach, jump back to inner foreach init")
		s3 := s1.EndForeach()
		s = s3.Send_Ai_bar("outer foreach body")
		fmt.Println("End of outer foreach")
		fmt.Println("Outer loop end", j)
	}
	s.EndForeach().End()
}

func protoBad() {
	// This function runs the protocol as follows:
	// --> enter outer loop body
	//   --> enter inner loop body
	//     --> inner body (Send_Aj_foo)
	//     --> inner body (Send_Aj_foo)
	//   --> exit inner loop
	//   --> outer loop remaining parts (Send_Ai_bar)
	//   --> enter inner loop (does not run body, ERROR)
	//   --> exit inner loop
	//   --> outer loop remaining parts (Send_Ai_bar)
	// --> exit outer loop

	s := new(proto.S0)
	s.
		Foreach().                                                      // outer body
		Foreach().Send_Aj_foo(1).Foreach().Send_Aj_foo(2).EndForeach(). // inner
		Send_Ai_bar("").
		Foreach().EndForeach(). // enter inner again, empty body
		Send_Ai_bar("").
		EndForeach(). // exit outer
		End()
}

func fusedRun() {
	// This function is the good use of foreach in the fused API design

	s := new(fused.S0)
	j := 0
	var (
		s1        *fused.S1
		moreOuter bool
	)
	for s1, moreOuter = s.Foreach(); moreOuter; {
		j++
		fmt.Println("Outer loop", j)
		i := 0
		var (
			s2        *fused.S2
			moreInner bool
		)
		for s2, moreInner = s1.Foreach(); moreInner; {
			i++
			fmt.Println("Inner loop", i)
			s1 = s2.Send_Aj_foo(i)
			fmt.Println("Inner loop end", i)
		}
		fmt.Println("End inner foreach, jump back to inner foreach init")
		s3 := s1.EndForeach()
		s = s3.Send_Ai_bar("outer foreach body")
		fmt.Println("End of outer foreach")
		fmt.Println("Outer loop end", j)
	}
	s.EndForeach().End()
}

func recurRun() {
	// This function is the good use of foreach in the parameter (recur) API design

	i := 0
	innerLoop := func(s *recur.S2) *recur.S1 {
		i++
		fmt.Println("Inner loop", i)
		innerBodyEnd := s.Send_Aj_foo(i)
		fmt.Println("Inner loop end", i)
		return innerBodyEnd
	}

	j := 0
	outerLoop := func(s *recur.S1) *recur.S0 {
		j++
		i = 0
		fmt.Println("Outer loop", j)
		outerBodyEnd := s.
			Foreach(innerLoop).
			Send_Ai_bar("outer foreach body")
		fmt.Println("End of outer foreach")
		fmt.Println("Outer loop end", j)
		return outerBodyEnd
	}

	s := new(recur.S0)
	s.Foreach(outerLoop).End()
}

func recurInlineRun() {
	// This function is the good use of foreach in the parameter (recur) API design
	// This is the inlined version

	j := 0

	s := new(recur.S0)
	s.Foreach(
		func(s *recur.S1) *recur.S0 {
			j++
			fmt.Println("Outer loop", j)
			i := 0
			outerBodyEnd := s.
				Foreach(
					func(s *recur.S2) *recur.S1 {
						i++
						fmt.Println("Inner loop", i)
						innerBodyEnd := s.Send_Aj_foo(i)
						fmt.Println("Inner loop end", i)
						return innerBodyEnd
					}).
				Send_Ai_bar("outer foreach body")
			fmt.Println("End of outer foreach")
			fmt.Println("Outer loop end", j)
			return outerBodyEnd
		}).End()
}

func nestedRun() {
	// This function is the good use of foreach in the parameter API design
	// based on nested FSM style

	j := 0

	s := new(nested.S0)
	s.Foreach(
		func(s *nested.S1) *nested.S4 {
			j++
			fmt.Println("Outer loop", j)
			i := 0
			outerBodyEnd := s.
				Foreach(
					func(s *nested.S2) *nested.S5 {
						i++
						fmt.Println("Inner loop", i)
						innerBodyEnd := s.Send_Aj_foo(i)
						fmt.Println("Inner loop end", i)
						return innerBodyEnd
					}).
				Send_Ai_bar("outer foreach body")
			fmt.Println("End of outer foreach")
			fmt.Println("Outer loop end", j)
			return outerBodyEnd
		}).End()
}

func forrangeRun() {
	// This function is the good use of foreach in the parameter API design
	// based on nested FSM style

	j := 0
	s := new(forrange.S0)
	loop0, end0 := s.Foreach()
	for body0 := range loop0 {
		j++
		fmt.Println("Outer loop", j)
		i := 0
		loop1, end1 := body0.Foreach()
		for body1 := range loop1 {
			i++
			fmt.Println("Inner loop", i)
			body1.Send_Aj_foo(i).End()
			fmt.Println("Inner loop end", i)
		}
		end1.Send_Ai_bar("outer foreach body").End()
		fmt.Println("End of outer foreach")
		fmt.Println("Outer loop end", j)
	}
	end0.End()
}
