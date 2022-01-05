package main

import (
	"fmt"
	"net/rpc"
)

type Args struct {
	A int
	B int
}

type Quotient struct {
	Quo int
	Rem int
}

type Arith int

func main() {
	client, _ := rpc.DialHTTP("tcp", "127.0.0.1:1234")

	// Synchronous call
	args := &Args{16, 8}

	var reply int

	_ = client.Call("Arith.Multiply", args, &reply)

	fmt.Printf("Arith: %d*%d=%d\n", args.A, args.B, reply)

	var reply2 = &Quotient{}
	client.Call("Arith.Divide", args, reply2)
	fmt.Printf("Arith: %d / %d=%d...%d\n", args.A, args.B, reply2.Quo, reply2.Rem)
}
