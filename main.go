package main

import (
	"fmt"
	"os"
)

var SEPARATOR = os.Args[1]

var log = GetLogger("tempo", 5)

func main() {
	inputString := "Hello, this is an appointment reminder from ABC. Popeye has an appointment scheduled with Doctor Sailor on Friday, April 21 at 2:30 PM. Please bring your insurance card and photo ID. We are located at the beach. To confirm this appointment, please press 1. To request a reschedule, please call the office at boat number 1337. If this has reached your voicemail, please call our office to confirm.. Repeating. Hello, this is an appointment reminder from ABC. Popeye has an appointment scheduled with Doctor Sailor on Friday, April 21 at 2:30 PM. Please bring your insurance card and photo ID. We are located at the beach. To confirm this appointment, please press 1. To request a reschedule, please call the office at boat number 1337. If this has reached your voicemail, please call our office to confirm.."
	input := SplitAfterAny(inputString, SEPARATOR)
	got := Template{
		Tokens: input,
		Chars: len(inputString),
		Breaks: []int{0, len(input)},
		Gaps: []int{0, 0},
		LostCumulative: 0,
		Matches:[10]int{0},
		Gen: 0,
	}

	fmt.Printf("%#v", input)
	_, _, pois := got.Match(input)
	got.ImproveTemplate(pois, len(input))
	got.Match(input)

	var cache Cache
	cache.Templates = make(map[*[]Template]struct{})
	if os.Args[2] == "process" {
		cache.Stream()
		cache.Dump(os.Args[1])
	} else {
		cache.Load(os.Args[1])
		cache.PrintTemplates()
		cache.StreamSavings()
		log.Infof("Savings - %#v", cache)
	}
	return

}
