package main

import (
	"strings"
	"sort"
	"os"
	"bufio"
	"regexp"
	"net/url"
)

type Cache struct {
	Templates [][]Template // only exact matches
}

type Potential struct {
	Template
	Match int
	Pois []poi
}

func (c *Cache) Process(input string) {
	// look for a match in templates
	var found Template
	var fidx int
	var toAdd []Template
	for i, t := range c.Templates {
		// get all the data about matching here and decide what to do
		// select/improve/reject/evict
		// go through each version one by one and then sort them at the end
		// if an exact match is found, delete all other versions

		var consider []Potential
		if len(t) > 1 {
			log.Debugf("number of version = %d", len(t))
		}
		for _, template := range t {
			misses, matchp, pois := template.Match(input)
			if misses == 0 {
				log.Infof("found the exact match")
				found = template
				break
			} else {
				// not an exact match
				if matchp > 50 {
					log.Debugf("match percentage %d%%", matchp)
					consider = append(consider, Potential{template, matchp, pois})
				}
			}
		}
		if found.Size() > 0 {
			toAdd = nil
			fidx = i
			break
		} else {
			// improve some templates
			if len(consider) == 0 {
				continue
			}
			// sort in order of matchp
			sort.Slice(consider, func(i, j int) bool {
				return consider[i].Match > consider[j].Match
			})
			// create the templates from those which require least number of changes to form a template/ or pick the one with max matchp
			consider[0].Template.ImproveTemplate(consider[0].Pois)
			t = append(t, consider[0].Template)
			// toAdd = append(toAdd, consider[0].Template)
		}
	}
	if len(toAdd) > 0 {
		log.Debugf("toAdd %v", toAdd[0])
		c.Templates = append(c.Templates, toAdd)
	}

	if found.Size() == 0 {
		// create a new template
		var t = Template{
			Tokens: strings.Split(input, " "),
			Breaks: []int{len(strings.Split(input, " "))},
			Matches: [10]int{0},
			IdleAge: 0,
		}
		c.Templates = append(c.Templates, []Template{t})
	} else {
			// remove all the other versions
			// c.Templates[fidx] = nil
			c.Templates[fidx] = []Template{found}
			log.Debugf("Removing all other version")
	}
}

func (c Cache) Stream() {
	file, err := os.Open("data/tts.log")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile("GET - /ssml\\?ssml=%3Cspeak%3E[^&]*%3C/speak%3E")

	limit := 10
	for scanner.Scan() {
		match := re.FindStringSubmatch(scanner.Text())
		if len(match) > 0 {
			log.Debugf("Templates cache size is %d", len(c.Templates))
			text, _ := url.QueryUnescape(strings.Join(strings.Split(match[0], "=")[1:], "="))
			if len(strings.Split(text, " ")) > 10 {
				log.Debugf(text)
				c.Process(text)
			}
			limit -= 1
			if (limit == 0) {
				break
			}
			// break
		}
	}
}
