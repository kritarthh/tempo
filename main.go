package main

import (
	"fmt"
)

var log = GetLogger("tempo", 5)

func main() {
	fmt.Printf("main\n")

	t := Template{
		Tokens:[]string{"a", "b", "c", "1", "2", "3"},
		Breaks: []int{3, 6},
		Matches:[10]int{1, 2, 3},
		IdleAge: 5,
	}
	t.Match("a b c sep 1 2 3")
	log.Info(t)

	got := Template{
		Tokens: []string{"This", "is", "a", "test", "a", "simple", "test"},
		Breaks: []int{4, 7},
		Matches:[10]int{1, 2, 3},
		IdleAge: 5,
	}

	_, _, pois := got.Match("This is a test but not a simple one")
	got.ImproveTemplate(pois)
	got.Match("This is a test but not a simple one")

	var cache Cache
	cache.Stream()
}
