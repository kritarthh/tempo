package main

import (
	"fmt"
	"testing"
)

func TestTemplateToString(t *testing.T) {
	got := Template{
		Tokens:[]string{"a", "b", "c", "1", "2", "3"},
		Breaks: []int{3, 6},
		Matches:[10]int{1, 2, 3},
		IdleAge: 5,
	}
	if fmt.Sprintf("%s", got) != "[a b c {{_}} 1 2 3 | 0:1 1:2 2:3]" {
		t.Errorf("Template(-1) = %s; want [a b c {{_}} 1 2 3 | 0:1 1:2 2:3]", got)
	}
}

func TestTemplateMatch(t *testing.T) {
	got := Template{
		Tokens:[]string{"This", "is", "a", "test", "a", "simple", "test"},
		Breaks: []int{4, 7},
		Matches:[10]int{1, 2, 3},
		IdleAge: 5,
	}

	got = Template{
		Tokens:[]string{"Hello", "Kritarth", "Chaudhary,", "your", "appointment", "is", "scheduled", "at", "1st", "Aug", "2020"},
		Breaks: []int{11},
		Matches:[10]int{1, 2, 3},
		IdleAge: 5,
	}

	_, _, pois := got.Match("Hello Budhha Prakash Singh, your appointment is scheduled at 2st Mar 2020")
	got.ImproveTemplate(pois)

	got.Match("Hello Budhha Prakash Singh, your appointment is scheduled at 2st Mar 2020")

	if fmt.Sprintf("%s", got) != "[Hello {{_}} your appointment is scheduled at {{_}} 2020 | 0:1 1:2 2:3]" {
		t.Errorf("Template(-1) = %s; want [Hello {{_}} your appointment is scheduled at {{_}} 2020 | 0:1 1:2 2:3]", got)
	}
}
