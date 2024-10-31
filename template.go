package main

import (
	"fmt"
	"strings"
	"sort"
	"encoding/json"
	"net/http"
	"io"
	"os"
	"crypto/md5"
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
	Chars int // number of characters in Tokens
	LostCumulative int // this is a cumulative field which can only increase with template modification
	// lost + chars must remain constant for a template
	Breaks []int    // cuts to be made in the tokens to make parts of the template
	Gaps []int      // average number of tokens used in the gaps made by breaks
	Matches [10]int // match count categorized by the number of misses
	Sample string
	Gen int
}

func (t Template) Cardinality() int {
	gapsAcc := 0
	for _, gap := range t.Gaps {
		gapsAcc += gap
	}
	return len(t.Tokens) + gapsAcc
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
	// out := ""
	// lastB := 0
	// for i, b := range t.Breaks {
	// 	// if (i == 0 || i == len(t.Breaks) - 1) && t.Gaps[i] == 0 {
	// 	// 	continue
	// 	// }
	// 	out = fmt.Sprintf("%s%s{{%d}}", out, strings.Join(t.Tokens[lastB:b], ""), t.Gaps[i])
	// 	lastB = b
	// }
	// out = fmt.Sprintf("%s | Chars:%d | Lost:%d | Matches:%d | SavedChars:%d", out, t.Chars, t.LostCumulative, t.Matches[0], t.Matches[0]*t.Chars)
	// return out
	// // return fmt.Sprintf("%#v", t)
    out := ""
	if len(t.Gaps) > 0 && t.Gaps[0] > 0 {
		out += fmt.Sprintf("{{ %d }} ", t.Gaps[0])
	}
	lastBreak := 0
	for i, g := range t.Gaps {
		if i > 0 {
			lastBreak = t.Breaks[i-1]
			out += fmt.Sprintf("%s", strings.Join(t.Tokens[lastBreak:t.Breaks[i]], ""))
			if g > 0 {
				out += fmt.Sprintf("{{ %d }} ", g)
			}
		}

	}
	out = fmt.Sprintf("%s | Chars:%d | Lost:%d | Matches:%d | SavedChars:%d |||||%s", out, t.Chars, t.LostCumulative, t.Matches[0], t.Matches[0]*t.Chars, t.Sample)
	return out
}

func (t Template) ToJson() string {
	b, err := json.Marshal(t)
    if err != nil {
        log.Errorf("%s", err)
        return "{}"
    }
    return string(b)
}

func (t Template) ToTTS(input []string, pois []poi) (parts []string) {
	dynamic := 0
	static := 0
	lastY := 0
	for _, poi := range pois {
		if poi.y-poi.v-lastY < 0 {continue}
		// log.Warnf("%#v - %#v - %d", input, pois, i)
		// dynamic part
		parts = append(parts, strings.Join(input[lastY:poi.y-poi.v], ""))
		dynamic += len(parts[len(parts)-1])
		// static template part
		parts = append(parts, strings.Join(input[poi.y-poi.v:poi.y], ""))
		static += len(parts[len(parts)-1])
		lastY = poi.y
	}
	if lastY < len(input) {
		// dynamic
		parts = append(parts, strings.Join(input[lastY:], ""))
		dynamic += len(parts[len(parts)-1])
	}
	if strings.Join(parts, "") != strings.Join(input, "") {
		log.Fatalf("%#v != %#v", strings.Join(parts, ""), strings.Join(input, ""))
	}
	// investigate performance of joins
	// perform tts only when significant number of joins are present
	if (dynamic * 100) / static > 25 || static < 100 {
		return
	}
	return

	// get audio from polly usiing both modified and unmodified XML
	// compare the quality

	var pcmParts []byte
	req, _ := http.NewRequest("GET", "https://pollyurl", nil)

	q := req.URL.Query()
	q.Add("audioformat", "pcm")
	q.Add("samplerate", "8000")
	q.Add("voice", "Salli")
	for i, p := range parts {
		log.Debugf("fetching from tts - %s", p)
		if p == "" {continue}
		q.Del("ssml")
		// pause := "strong"
		// q.Add("ssml", fmt.Sprintf("<speak><break strength=\"%s\"/>%s</speak>", pause, p))
		q.Add("ssml", fmt.Sprintf("<speak>%s</speak>", p))
		req.URL.RawQuery = q.Encode()
		resp, err := http.Get(req.URL.String())
		if err != nil {
			log.Errorf("error tts request - %#v", err)
			break
		}
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			break
		}
		if i != 0 {
			// add silence at the begining
			silence := make([]byte,int(8000*0.75))
			pcmParts = append(pcmParts, silence...)
		}
		pcmParts = append(pcmParts, bodyBytes...)
	}
	pcmfile := fmt.Sprintf("pcm/%x.pcm", md5.Sum([]byte(strings.Join(parts, ""))))
	os.Remove(pcmfile)
	out, err := os.Create(pcmfile)
	if err != nil {
		log.Error(err)
		return
	}
	defer out.Close()
	if _, err := out.Write(pcmParts); err != nil {
		log.Errorf("%#v", err)
	}
	// save the original speech now for comparision
	q.Del("ssml")
	q.Add("ssml", fmt.Sprintf("<speak>%s</speak>", strings.Join(parts, "")))
	req.URL.RawQuery = q.Encode()
	resp, err := http.Get(req.URL.String())
	if err != nil {
		log.Errorf("error tts request - %#v", err)
	}
	defer resp.Body.Close()
	pcmfile = fmt.Sprintf("pcm/%x-orig.pcm", md5.Sum([]byte(strings.Join(input, ""))))
	os.Remove(pcmfile)
	out, err = os.Create(pcmfile)
	if err != nil {
		log.Error(err)
		return
	}
	defer out.Close()
	io.Copy(out, resp.Body)
	return
}

func (t *Template) FromString(tmpl string) {
	return
}

// return true only if exact match is found
func (t *Template) Match(Y []string) (misses int, matches int, templatePois []poi) {
	// log.Debugf("match string: %s against template: %#v", Y, t)
	// empty template matches with nothing
	if (len(t.Tokens) == 0) {
		return
	}

	var X []string
	X = t.Tokens
	matrix, pois := matchMatrix(X, Y)
	templatePois = extractTemplate(pois);

	// log.Debugf(printDetails(matrix, pois, templatePois))
	_ = matrix

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
		// TODO: FIX ME
		// t.Gaps[len(templatePois)] = ((len(Y) - templatePois[len(templatePois) - 1].y) * t.Matches[0] + t.Gaps[len(templatePois)])/(t.Matches[0] + 1)
		// gapsAcc += t.Gaps[len(templatePois)]
	}

	if misses > 9 {
		misses = 9
	}
	t.Matches[misses] += 1
	matches = ((t.MatchesTemplate(templatePois) + gapsAcc) * 100) / len(Y)
	if len(Y) < t.Cardinality() {
		matches = ((t.MatchesTemplate(templatePois) + gapsAcc) * 100) / t.Cardinality()
	}

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

	// t.size is based on breaks which is always 1 more than number of pois
	misses += len(pois) - (t.Size() - 1)
	if misses < 0 {
		misses = -misses
	}

	xpos := t.MatchesTemplate(pois) - len(t.Tokens)
	// // scale
	// xpos /= misses + 1
	if xpos > 0 {
		misses += xpos
	} else {
		misses -= xpos
	}

	// log.Debugf("misses: %d", misses)
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
	log.Debugf("improve template %#v with %#v", t, pois)
	var tokens Tokens
	var breaks []int
	var gaps []int
	acc := 0
	lost := 0
	lastY := 0
	lastX := 0
	// always add a break in the beginning
	breaks = append(breaks, 0)
	for i, poi := range pois {
		acc += poi.v
		for j := lastX ; j < poi.x - poi.v ; j++ {
			lost += len(t.Tokens[j])
		}
		tokens = append(tokens, t.Tokens[poi.x-poi.v:poi.x]...)
		breaks = append(breaks, acc)
		gaps = append(gaps, pois[i].y - pois[i].v - lastY)
		lastY = poi.y
		lastX = poi.x
	}
	gaps = append(gaps, inputLength - pois[len(pois) - 1].y)
	for j := lastX ; j < len(t.Tokens) ; j++ {
		lost += len(t.Tokens[j])
	}
	t.Tokens = tokens
	t.Breaks = breaks
	t.Gaps = gaps
	t.Chars = (t.Chars + t.LostCumulative) - lost // total - lost
	t.LostCumulative += lost
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
