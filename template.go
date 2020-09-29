package main

import (
	"fmt"
	"strings"
	"sort"
)

type Tokens []string
func (t Tokens) String() string {
	return strings.Join(t, " ")
}

type poi struct {
	x int
	y int
	v int
}
func (p poi) String() string {
	return fmt.Sprintf("(%d, %d, %d)", p.x, p.y, p.v)
}

type Template struct {
	Tokens
	Breaks []int    // cuts to be made in the tokens to make parts of the template
	Gaps []int      // average number of tokens used in the gaps made by the cuts
	Matches [10]int // match count categorized by the number of misses
	Gen int
}

func (t Template) Size() int {
	return len(t.Breaks)
}

func (t Template) ExactMatches() (int) {
	return t.Matches[0]
}

func (t Template) NearMissMatches() (int) {
	return t.Matches[1] + t.Matches[2]
}

func (t Template) String() string {
	if len(t.Gaps) == 0 {
		return ""
	}
	out := fmt.Sprintf("{{%d}} ", t.Gaps[0])
	idx := 0
	for i, b := range t.Breaks {
		out += strings.Join(t.Tokens[idx:b], " ") + fmt.Sprintf(" {{%d}} ", t.Gaps[i+1])
		idx=b
	}
	out = out[:len(out)-1]
	return fmt.Sprintf("[%s | 0:%d 1:%d 2:%d]", out, t.Matches[0], t.Matches[1], t.Matches[2])
}

// return true only if exact match is found
func (t Template) Match(Y []string) (misses int, matches int, templatePois []poi) {
	log.Debugf("current template: %s", t)
	// empty template matches with nothing
	if (len(t.Tokens) == 0) {
		return
	}

	var X []string
	X = t.Tokens
	matrix, pois := matchMatrix(X, Y)
	templatePois = extractTemplate(pois);

	_ = matrix
	// log.Debugf(printDetails(matrix, pois, templatePois))

	misses = t.MissesTemplate(templatePois)


	// adapt the gaps if misses are zero
	gapsAcc := 0
	if misses == 0 {
		lastY := 0
		for i, poi := range templatePois {
			t.Gaps[i] = ((poi.y - poi.v - lastY) * t.Matches[0] + t.Gaps[i])/(t.Matches[0] + 1)
			gapsAcc += t.Gaps[i]
			lastY = poi.y
		}
		t.Gaps[len(templatePois)] = ((len(Y) - templatePois[len(templatePois) - 1].y) * t.Matches[0] + t.Gaps[len(templatePois)])/(t.Matches[0] + 1)
		gapsAcc += t.Gaps[len(templatePois)]
	}

	if misses > 9 {
		misses = 9
	}
	t.Matches[misses] += 1
	log.Debugf("misses: %d", misses)
	matches = ((t.MatchesTemplate(templatePois) + gapsAcc) * 100) / len(Y)
	if matches > 100 {
		matches = 100/matches
	}

	// log.Debugf("matches: %d", matches)

	// make the decision
	// b = true
	//
	// 1. comparable lengths , +- 20%
	// 2. sum of top 3 lengths >= 80% input => <20% of variables
	// 3. Number of breaks <= 5
	//
	// for exact match, the templatePois should exactly match the parts of the template
	// if a part differs by one token then it would account for 1 miss
	// if a part differs by two tokens then it would account for 2 miss
	// if 2 parts differs by 1 token each then it would account for 2 miss
	// basically, each token accounts for 1 miss
	return
}

func (t Template) MissesTemplate(pois []poi) (misses int) {
	// count the numbers of pois that you expect
	// find the xs, where you expect them to be in the pois
	// each deviation adds 1 towards the misses count
	misses += len(pois) - t.Size()
	if misses < 0 {
		misses *= -1
	}

	xpos := 0
	for _, poi := range pois {
		xpos += poi.v
	}
	xpos -= len(t.Tokens)
	// scale
	xpos /= misses + 1

	if xpos > 0 {
		misses += xpos
	} else {
		misses -= xpos
	}


	return
}

func (t Template) MatchesTemplate(pois []poi) (matches int) {
	// count the number of matching tokens

	for _, poi := range pois {
		matches += poi.v
	}
	return
}

// pass the pointer so that the write happens on the original
func (t *Template) ImproveTemplate(pois []poi, inputLength int) {
	if len(pois) == 0 {
		log.Warnf("empty pois, cannot improve template")
		return
	}
	var tokens Tokens
	var breaks []int
	var gaps []int
	acc := 0
	lastY := 0
	for i, poi := range pois {
		tokens = append(tokens, t.Tokens[poi.x-poi.v:poi.x]...)
		breaks = append(breaks, acc+poi.v)
		gaps = append(gaps, pois[i].y - pois[i].v - lastY)
		acc += poi.v
		lastY = poi.y
	}
	gaps = append(gaps, inputLength - pois[len(pois) - 1].y)
	t.Tokens = tokens
	t.Breaks = breaks
	t.Gaps = gaps
	return
}

func extractTemplate(pois []poi) []poi {
	// start from the longest substring and build the template from there
	// var m = len(matrix[0])
	// var n = len(matrix)

	if len(pois) < 2 {
		return pois
	}

	best := pois[0]
	var leftPois, rightPois []poi
	for _, p := range pois {
		if p.v < 2 {
			continue
		}
		if p.x <= (best.x - best.v) && p.y <= (best.y - best.v) {
			leftPois = append(leftPois, p)
		} else if p.x > best.x && p.y > best.y {
			rightPois = append(rightPois, p)
		}
	}
	sort.Slice(leftPois, func(i, j int) bool {
		return leftPois[i].v > leftPois[j].v
	})
	sort.Slice(rightPois, func(i, j int) bool {
		return rightPois[i].v > rightPois[j].v
	})

	leftTemplate := extractTemplate(leftPois)
	rightTemplate := extractTemplate(rightPois)

	var templatePois []poi
	templatePois = append(templatePois, leftTemplate...)
	templatePois = append(templatePois, best)
	templatePois = append(templatePois, rightTemplate...)

	return templatePois
}

func matchMatrix(X []string, Y []string) (dptab [][]int, pois []poi) {
	dptab = make([][]int, len(Y)+1)
	for i := range dptab {
		dptab[i] = make([]int, len(X)+1)
	}
	for i := 0; i <= len(X); i++ {
		for j := 0; j <= len(Y); j++ {
			if (i == 0 || j == 0) {
				dptab[j][i] = 0
			} else if (X[i-1] == Y[j-1]) {
				dptab[j][i] = dptab[j-1][i-1] + 1
				// if this is on edge, it is potentially max
				if (i == len(X) || j == len(Y)) {
					pois = append(pois, poi{i, j, dptab[j][i]})
				}
			} else {
				dptab[j][i] = 0
				// if the previous is non zero then this is potentially a max
				if (dptab[j-1][i-1] > 0) {
					pois = append(pois, poi{i-1, j-1, dptab[j-1][i-1]})
				}
			}
		}
	}
	sort.Slice(pois, func(i, j int) bool {
		return pois[i].v > pois[j].v
	})
	return
}

func printDetails(dptab [][]int, pois []poi, templatePois []poi) string {
	out := fmt.Sprintf("\nmatching matrix\n")
	for i := 1; i < len(dptab); i++ {
		for j := 1; j < len(dptab[0]); j++ {
			out += fmt.Sprintf("|%d", dptab[i][j])
		}
		out += fmt.Sprintf("\n")
	}
	out += "points of interest\n"
	for _, p := range pois {
		out += fmt.Sprintf("poi%s\n", p)
	}
	out += "\npoints of interest of the best template\n"
	for _, p := range templatePois {
		out += fmt.Sprintf("poi%s\n", p)
	}
	return out
}
