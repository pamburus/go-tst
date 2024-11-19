package main

import (
	"math/rand/v2"
	"reflect"
	"testing"
	"unique"
)

func BenchmarkTypeComparison(b *testing.B) {
	b.StopTimer()
	x := [2]reflect.Type{reflect.TypeFor[string](), reflect.TypeFor[int]()}
	var r bool
	y := x[rand.Int()%2]
	z := x[rand.Int()%2]
	b.StartTimer()
	for range b.N {
		r = y == z
	}
	b.StopTimer()
	res = r
}

func BenchmarkUniqueComparison(b *testing.B) {
	b.StopTimer()
	x := [2]unique.Handle[reflect.Type]{unique.Make(reflect.TypeFor[string]()), unique.Make(reflect.TypeFor[int]())}
	var r bool
	y := x[rand.Int()%2]
	z := x[rand.Int()%2]
	b.StartTimer()
	for range b.N {
		r = y == z
	}
	b.StopTimer()
	res = r
}

func BenchmarkIntComparison(b *testing.B) {
	b.StopTimer()
	var r bool
	y := rand.Int() % 2
	z := rand.Int() % 2
	b.StartTimer()
	for range b.N {
		r = y == z
	}
	b.StopTimer()
	res = r
}

var res bool
