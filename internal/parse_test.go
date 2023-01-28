package internal

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParser(t *testing.T) {
	tests := []struct {
		name    string
		entity  any
		wantRes []*Node
		wantErr error
	}{
		{
			name:    "nil entity",
			entity:  nil,
			wantErr: errors.New("invalid entity"),
		},
		{
			name:    "Not struct entity",
			entity:  NotStruct,
			wantErr: errors.New("invalid entity"),
		},
		{
			name: "basic struct", // 结构体
			entity: Basic{
				Id:    1,
				Name:  "",
				Score: 0,
				Email: "",
			},
			wantRes: []*Node{&Node{
				Field: "Id",
				Tags: []*Tag{{
					Operator: "gt",
					Value:    0,
				}, {
					Operator: "lt",
					Value:    100,
				}},
				Fields: nil,
			}},
		},
		{
			name: "basic ptr struct", // 指针类型
			entity: &Basic{
				Id:    10,
				Name:  "",
				Score: 10.2,
				Email: "",
			},
		},
		{
			// 内嵌结构体
			name:   "Nested struct",
			entity: Nested{},
		},
	}

	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			check := NewEntity()
			err := check.Parser(ts.entity)
			assert.Equal(t, ts.wantErr, err)
			if err != nil {
				return
			}

			fmt.Printf("%#v\n", check)
		})
	}
}

var NotStruct int = 10

type Basic struct {
	Id    int     `check:"gt 0;lt 100"`
	Name  string  `check:"notEmpty"`
	Score float32 `check:"gt 0.00"`
	Email string  `check:"email"`
}

type Nested struct {
	Id      int     `check:"gt 0;lt 100"`
	Name    string  `check:"notEmpty"`
	Score   float32 `check:"gt 0.00"`
	Email   string  `check:"email"`
	Address `check:"required"`
	Addr    *Address `check:"required"`
}

type Address struct {
	Id       int `check:"lt 10"`
	Province string
	City     string
}
