package random

import (
	"testing"
)

type TestRandItem struct {
	Id     int
	Weight int
}

type TestRandList []*TestRandItem

func (t TestRandList) Len() int {
	return len(t)
}

func (t TestRandList) Index(i int) interface{} {
	return t[i]
}

func (t TestRandList) Weight(i int) int {
	return t[i].Weight
}

func TestRandomAward(t *testing.T) {
	list := TestRandList(makeList(10))
	result := make(map[int]int)
	for i := 0; i < 10000000; i++ {
		item := RandomAward(list).(*TestRandItem)
		result[item.Id] += 1
	}
	t.Log(result)
}

func makeList(count int) []*TestRandItem {
	list := make([]*TestRandItem, count)
	for i := range list {
		list[i] = &TestRandItem{Id: i, Weight: 10}
	}
	return list
}

/*
BenchmarkRandomAward5-8     	10000000	       140 ns/op
BenchmarkRandomAward30-8    	 5000000	       380 ns/op
BenchmarkRandomAward100-8   	 2000000	       982 ns/op
*/
func BenchmarkRandomAward5(b *testing.B) {
	list := TestRandList(makeList(5))
	for i := 0; i < b.N; i++ {
		RandomAward(list)
	}
}

func BenchmarkRandomAward30(b *testing.B) {
	list := TestRandList(makeList(30))
	for i := 0; i < b.N; i++ {
		RandomAward(list)
	}

}

func BenchmarkRandomAward100(b *testing.B) {
	list := TestRandList(makeList(100))
	for i := 0; i < b.N; i++ {
		RandomAward(list)
	}
}

func BenchmarkRandomAward5000(b *testing.B) {
	list := TestRandList(makeList(5000))
	for i := 0; i < b.N; i++ {
		RandomAward(list)
	}
}

type TestRandListWithTotal struct {
	data  []*TestRandItem
	total int
}

func (t TestRandListWithTotal) Len() int {
	return len(t.data)
}

func (t TestRandListWithTotal) Index(i int) interface{} {
	return t.data[i]
}

func (t TestRandListWithTotal) Weight(i int) int {
	return t.data[i].Weight
}

func (t TestRandListWithTotal) TotalWeight() int {
	return t.total
}

func makeListWithTotal(count int) *TestRandListWithTotal {
	list := make([]*TestRandItem, count)
	for i := range list {
		list[i] = &TestRandItem{Id: i, Weight: 10}
	}
	return &TestRandListWithTotal{
		data:  list,
		total: 10 * count,
	}
}

func BenchmarkRandomAwardWithTotal5(b *testing.B) {
	list := makeListWithTotal(5)
	for i := 0; i < b.N; i++ {
		RandomAward(list)
	}
}

func BenchmarkRandomAwardWithTotal30(b *testing.B) {
	list := makeListWithTotal(30)
	for i := 0; i < b.N; i++ {
		RandomAward(list)
	}

}

func BenchmarkRandomAwardWithTotal100(b *testing.B) {
	list := makeListWithTotal(100)
	for i := 0; i < b.N; i++ {
		RandomAward(list)
	}
}

func BenchmarkRandomAwardWithTotal5000(b *testing.B) {
	list := makeListWithTotal(5000)
	for i := 0; i < b.N; i++ {
		RandomAward(list)
	}
}
