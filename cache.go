package main

import (
	"strings"
	"sort"
	"os"
	"bufio"
	"fmt"
	"encoding/json"
	// "regexp"
	// "net/url"
)

type Cache struct {
	Templates map[*[]Template]struct{} // only exact matches
	Savings int
	Hits int
	Misses int
}

type Potential struct {
	Template
	Match int
	Pois []poi
}

func SplitAfterAny(s string, seps string) []string {
	fields := []string{}
	last := 0
	for i, r := range s {
        if strings.ContainsRune(seps, r) {
			if i != 0 && i == last {
				// nothing but separator in this section
				fields[len(fields) - 1] += string(r)
			} else {
				withoutSpace := strings.TrimSpace(s[last:i+1])
				if len(fields) > 0 &&
					((len(withoutSpace) == 1 && strings.Contains(seps, withoutSpace)) ||
						len(withoutSpace) < 5) {
					// possibly just a space with separator
					fields[len(fields) - 1] += s[last:i+1]
				} else {
					fields = append(fields, s[last:i+1])
				}
			}
			last = i + 1
		}
	}
	if s != "" {
		fields = append(fields, s[last:])
	}
	return fields
}

func (c *Cache) Process(inputString string, currentGen int) {
	// look for a match in templates
	var found Template
	var input = SplitAfterAny(inputString, SEPARATOR)
	for t := range c.Templates {
		// get all the data about matching here and decide what to do
		// select/improve/reject/evict
		// go through each version one by one and then sort them at the end
		// if an exact match is found, delete all other versions

		for i := 0; i < len(*t); i++ {
			// if currentGen - (*t)[i].Gen > 100 && (*t)[i].Matches[9] > 10 {
			if (*t)[i].Matches[0] < 5 && (currentGen - (*t)[i].Gen > 500) {
				*t = append((*t)[:i], (*t)[i+1:]...)
				i--
			}
		}
		if len(*t) == 0 {
			delete(c.Templates, t)
			continue
		}

		var consider []Potential
		gen := 0
		for _, template := range *t {
			// only interact with similar cardinality templates
			cardinalityP := (len(input) - template.Cardinality())*100/template.Cardinality()
			if cardinalityP < 0 {
				cardinalityP = -cardinalityP
			}
			if cardinalityP > 33 {
				// log.Warnf("Cardinality mismatch: %d%%, must be less than 25%%", cardinalityP)
				continue
			}
			gen += template.Gen
			misses, matchp, pois := template.Match(input)
			// log.Debugf("match percentage %d%%", matchp)
			if misses == 0 && matchp > 75 {
				found = template
				found.Gen = currentGen
				log.Debugf("Found template: %#v", found)
				break
			} else {
				// not an exact match
				if matchp > 50 {
					consider = append(consider, Potential{template, matchp, pois})
				} else {
				}
			}
		}

		// idleAge := currentGen - gen/len(*t)
		// log.Debugf("average idleAge: %d", idleAge)
		if found.Size() > 0 {
			// remove all the other versions
			log.Debugf("Removing all other version")
			*t = []Template{found}
			break
		} else {
			// should this be deleted?
			// if idleAge > 1000 {
			// 	log.Debugf("deleting unused template with idleAge %d", idleAge)
			// 	delete(c.Templates, t)
			// 	continue
			// }
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
			Chars: len(inputString),
			Breaks: []int{0, len(input)},
			Gaps: []int{0, 0},
			LostCumulative: 0,
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

func (c *Cache) CalculateSavings(inputString string) {
	// look for a match in templates
	var found Template
	var input = SplitAfterAny(inputString, SEPARATOR)
	for t := range c.Templates {
		for _, template := range *t {
			misses, matchp, pois := template.Match(input)
			// log.Debugf("match percentage %d%%", matchp)
			if misses == 0 && matchp > 75 {
				found = template
				log.Debugf("Found template: %#v", found)
				log.Debugf("TTS splits: %#v", template.ToTTS(input, pois))
				break
			}
		}

		if found.Size() > 0 {
			// remove all the other versions
			// log.Debugf("Removing all other version")
			// *t = []Template{found}
			log.Infof("found the exact match %v", found)
			chars := 0
			for _, tk := range found.Tokens {
				chars += len(tk)
			}
			log.Infof("saved chars %d", chars)
			c.Savings += chars
			c.Hits += 1
			break
		}
	}
	if found.Size() == 0 {
		c.Misses += 1
	}
}

func (c *Cache) StreamSavings() {

	// file, err := os.Open("data/tts.log")
	// file, err := os.Open("data/use1.txt.10K")
	file, err := os.Open("data/use1.txt")
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
			// log.Debugf("Templates cache size is %d", len(c.Templates))
			// text, _ := url.QueryUnescape(strings.Join(strings.Split(match[0], "=")[1:], "="))
			if len(SplitAfterAny(text, SEPARATOR)) > 1 {
				// log.Debugf(text)
				c.CalculateSavings(text)
				// continue_on_key()
			}
			count += 1
			if (count % 100 == 0) {
				log.Warnf("saved chars count: %d, total count: %d, cache size: %d", count, totalCount, len(c.Templates))
			}
		}
	}
}

func (c Cache) Stream() {

	// file, err := os.Open("data/tts.log")
	// file, err := os.Open("data/use1.txt.10K")
	file, err := os.Open("data/use1.txt")
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
			// log.Debugf("Templates cache size is %d", len(c.Templates))
			// text, _ := url.QueryUnescape(strings.Join(strings.Split(match[0], "=")[1:], "="))
			if len(SplitAfterAny(text, SEPARATOR)) > 1 {
				// log.Debugf(text)
				c.Process(text, count)
			}
			count += 1
			if (count % 100 == 0) {
				log.Warnf("saved chars count: %d, total count: %d, cache size: %d", count, totalCount, len(c.Templates))
			}
			// continue_on_key()
		}
	}
}

func (c Cache) PrintTemplates() {
	log.Debugf("Printing templates %d", len(c.Templates))
	for t := range c.Templates {
		log.Debugf("%#v", t)
	}
}

func (c Cache) Dump(suffix string) {
	log.Debugf("Dumping templates %d", len(c.Templates))
	var tmpls []Template
	for t := range c.Templates {
		sort.Slice((*t), func(i, j int) bool {
			return (*t)[i].Matches[0] > (*t)[j].Matches[0]
		})
		tmpls = append(tmpls, (*t)[0])
	}
	sort.Slice(tmpls, func(i, j int) bool {
		return tmpls[i].Matches[0] * tmpls[i].Chars < tmpls[j].Matches[0] * tmpls[i].Chars
	})

    f, _ := os.Create(fmt.Sprintf("templates-%s.txt", suffix))
    defer f.Close()
    w := bufio.NewWriter(f)
	for _, t := range tmpls {
		if t.Matches[0] > 5 {
			log.Debugf("matches:%d - %#v - %s", t.Matches[0], t, t.ToJson())
			_, _ = w.WriteString(t.ToJson())
			_, _ = w.WriteString("\n")
		}
	}
    w.Flush()
}

func (c Cache) Load(suffix string) {
	log.Debugf("Loading templates %d", len(c.Templates))
	file, err := os.Open(fmt.Sprintf("templates-%s.txt", suffix))
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

	for scanner.Scan() {
		text := scanner.Text()
		if len(text) > 0 {
			var t Template
			err := json.Unmarshal([]byte(text), &t)
			if err != nil {
				panic(err)
			}
			c.Templates[&[]Template{t}] = struct{}{}
		}
	}
}

func continue_on_key() {
    log.Debugf("Continue?")
    var input string
    fmt.Scanln(&input)
}
