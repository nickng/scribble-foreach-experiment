package main

import (
	"fmt"
	"log"

	"github.com/nickng/scribble-foreach-experiment/fused"
	"github.com/nickng/scribble-foreach-experiment/proto"
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
	fused.ProtoParam["k"] = 2
	fmt.Println("---- fused foreach ----")
	fusedRun()

	proto.ProtoParam["k"] = 2
	fmt.Println("---- GOOD ----")
	protoGood()
	fmt.Println("---- BAD ----")
	protoBad()

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
