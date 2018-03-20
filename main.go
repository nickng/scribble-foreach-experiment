package main

import (
	"fmt"
	"log"

	"github.com/nickng/pscribble-foreach/proto"
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
	proto.ProtoParam["k"] = 2

	fmt.Println("---- GOOD ----")
	good()
	fmt.Println("---- BAD ----")
	bad()
}

func good() {
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

func bad() {
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
