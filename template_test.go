package main

import (
	"fmt"
	"testing"
	"strings"
)

func TestTemplateToString(t *testing.T) {
	got := Template{
		Tokens:[]string{"a", "b", "c", "1", "2", "3"},
		Breaks: []int{3, 6},
		Gaps: []int{0, 1, 0},
		Matches:[10]int{1, 2, 3},
		Gen: 0,
	}
	if fmt.Sprintf("%s", got) != "[a b c {{_}} 1 2 3 | 0:1 1:2 2:3]" {
		t.Errorf("Template(-1) = %s; want [a b c {{_}} 1 2 3 | 0:1 1:2 2:3]", got)
	}
}

func TestTemplateMatch(t *testing.T) {
	got := Template{
		Tokens:[]string{"This", "is", "a", "test", "a", "simple", "test"},
		Breaks: []int{4, 7},
		Gaps: []int{0, 1, 0},
		Matches:[10]int{1, 2, 3},
		Gen: 0,
	}

	got = Template{
		Tokens:[]string{"Hello", "John", "Doe", "your", "appointment", "is", "scheduled", "at", "1st", "Aug", "2020"},
		Breaks: []int{11},
		Gaps: []int{0, 0},
		Matches:[10]int{1, 2, 3},
		Gen: 0,
	}

	input := strings.Split("Hello A B C, your appointment is scheduled at 2st Mar 2020", " ")
	_, _, pois := got.Match(input)
	got.ImproveTemplate(pois, len(input))

	got.Match(input)

	if fmt.Sprintf("%s", got) != "[Hello {{_}} your appointment is scheduled at {{_}} 2020 | 0:1 1:2 2:3]" {
		t.Errorf("Template(-1) = %s; want [Hello {{_}} your appointment is scheduled at {{_}} 2020 | 0:1 1:2 2:3]", got)
	}
}
