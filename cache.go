package main

import (
	"strings"
	"sort"
	"os"
	"bufio"
	// "regexp"
	// "net/url"
)

type Cache struct {
	Templates map[*[]Template]struct{} // only exact matches
}

type Potential struct {
	Template
	Match int
	Pois []poi
}

func (c *Cache) Process(inputString string, currentGen int) {
	// look for a match in templates
	var found Template
	input := strings.Split(inputString, " ")
	for t := range c.Templates {
		// get all the data about matching here and decide what to do
		// select/improve/reject/evict
		// go through each version one by one and then sort them at the end
		// if an exact match is found, delete all other versions

		var consider []Potential
		if len(*t) > 1 {
			log.Debugf("number of version = %d", len(*t))
			if len(*t) > 3 {
				// todo: remove only unused versions
				delete(c.Templates, t)
				continue
			}
		}

		gen := 0
		for _, template := range *t {
			gen += template.Gen
			misses, matchp, pois := template.Match(input)
			if misses == 0 && matchp > 50 {
				found = template
				found.Gen = currentGen
				break
			} else {
				// not an exact match
				if matchp > 50 {
					log.Debugf("match percentage %d%%", matchp)
					consider = append(consider, Potential{template, matchp, pois})
				} else {

				}
			}
		}
		idleAge := currentGen - gen/len(*t)
		log.Debugf("average idleAge: %d", idleAge)
		if found.Size() > 0 {
			// remove all the other versions
			log.Debugf("Removing all other version")
			*t = []Template{found}
			break
		} else {
			// should this be deleted?
			if idleAge > 1000 {
				log.Debugf("deleting unused template with idleAge %d", idleAge)
				delete(c.Templates, t)
				continue
			}
			// improve some templates
			if len(consider) == 0 {
				continue
			}
			// sort in order of matchp
			sort.Slice(consider, func(i, j int) bool {
				return consider[i].Match > consider[j].Match
			})
			// create the templates from those which require least number of changes to form a template/ or pick the one with max matchp
			consider[0].Template.ImproveTemplate(consider[0].Pois, len(input))
			*t = append(*t, consider[0].Template)
		}

	}
	if found.Size() == 0 {
		// create a new template
		var t = Template{
			Tokens: input,
			Breaks: []int{len(input)},
			Gaps: []int{0, 0},
			Matches: [10]int{0},
			Gen: currentGen,
		}
		c.Templates[&[]Template{t}] = struct{}{}
	} else {
		log.Infof("found the exact match %v", found)
		chars := 0
		for _, tk := range found.Tokens {
			chars += len(tk)
		}
		chars += len(found.Tokens)
		log.Infof("saved chars %d", chars)
	}
}

func (c Cache) Stream() {
	c.Templates = make(map[*[]Template]struct{})

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
	// re := regexp.MustCompile("GET - /ssml\\?ssml=%3Cspeak%3E[^&]*%3C/speak%3E")

	totalCount := 0
	count := 0
	for scanner.Scan() {
		totalCount += 1
		// match := re.FindStringSubmatch(scanner.Text())
		text := scanner.Text()
		// if len(match) > 0 {
		if len(text) > 0 {
			log.Debugf("Templates cache size is %d", len(c.Templates))
			// text, _ := url.QueryUnescape(strings.Join(strings.Split(match[0], "=")[1:], "="))
			if len(strings.Split(text, " ")) > 10 {
				log.Debugf(text)
				c.Process(text, count)
			}
			count += 1
			if (count % 100 == 0) {
				log.Warnf("saved chars count: %d, total count: %d, cache size: %d", count, totalCount, len(c.Templates))
			}
			// log.Info("press any key to continue")
			// input := bufio.NewScanner(os.Stdin)
			// input.Scan()
		}
	}
}
