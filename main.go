package main

import (
	"fmt"
	"strings"
)

var log = GetLogger("tempo", 5)

func main() {
	fmt.Printf("main\n")

	t := Template{
		Tokens:[]string{"a", "b", "c", "1", "2", "3"},
		Breaks: []int{3, 6},
		Gaps: []int{0, 1, 0},
		Matches:[10]int{1, 2, 3},
		Gen: 0,
	}
	t.Match(strings.Split("a b c sep 1 2 3", " "))
	log.Info(t)

	got := Template{
		Tokens: []string{"This", "is", "a", "test", "a", "simple", "test"},
		Breaks: []int{4, 7},
		Gaps: []int{0, 1, 0},
		Matches:[10]int{1, 2, 3},
		Gen: 0,
	}

	input := strings.Split("This is a test but not a simple one", " ")
	_, _, pois := got.Match(input)
	got.ImproveTemplate(pois, 9)
	got.Match(input)

	var cache Cache
	cache.Stream()
}
