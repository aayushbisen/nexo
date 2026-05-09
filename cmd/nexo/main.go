package main

import (
	"fmt"
	"nexo/internal/hashring"
	"nexo/internal/network"
	"nexo/internal/store"
)

func main() {
	// fmt.Println("Nexo Cache Starting...")
	// ss := store.New[string]()
	// s.Set("name", "aayush")
	// val, ok := s.Get("name")
	// if ok {
	// 	fmt.Printf("My name is %s", val)
	// }
	// val2, ok2 := s.Get("age")
	// fmt.Println(val2)
	// fmt.Println(ok2)
	// s.Delete("name")
	// val3, ok3 := s.Get("name")
	// if ok3 {
	// 	fmt.Printf("My name is %s", val3)
	// }
	// s2 := store.New[int]()
	// s2.Set("age", 24)
	// val4, ok4 := s2.Get("age")
	// if ok4 {
	// 	fmt.Printf("My age is %d", val4)
	// }

	// s := network.Server{St: ss, Port: 9090}
	// s.Start()

	workers := []int{9091, 9092, 9093, 9094}

	for _, port := range workers {
		ss := store.New[string](100)
		go func() {
			s := network.Server{St: ss, Port: port}
			s.Start()
		}()
	}

	hr := hashring.New()
	for _, port2 := range workers {
		hr.AddNode(fmt.Sprintf("localhost:%d", port2), 5)
	}

	c := network.Coordinator{Port: 9090, Ring: hr}
	c.Start()
}
