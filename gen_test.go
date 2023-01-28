package main

import (
	"SJT/struct-validate/pkg"
	"SJT/struct-validate/test_data/b/c/d"
	"fmt"
	"testing"
)

func TestGen(t *testing.T) {
	g := pkg.NewGenDefinition()
	fmt.Println(g.Gen(d.Nested{}))
}

func TestValidate(t *testing.T) {
	//id := 99
	//n := &d.Nested{
	//	Id:      &id,
	//	Name:    "1",
	//	Score:   10.6,
	//	Email:   "",
	//	Slice:   []int{},
	//	Map:     map[string]int{},
	//	Chan:    make(chan int),
	//	Address: b.Address{AddressId: 91, Detail: b.Detail{Detail: "1"}},
	//	Addr: &b.Address{
	//		AddressId: 11,
	//		Province:  "",
	//		City:      "",
	//		Detail:    b.Detail{Detail: "1"},
	//	},
	//}
	//if err := n.Validator(); err != nil {
	//	fmt.Println(err)
	//}
}
