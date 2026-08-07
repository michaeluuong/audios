package main

import (
	"fmt"
	"regexp"
	"strings"
)

func main() {
	/*name := "Let Me Out (Live in Los Angeles, 1978) ((Live in Los Angeles, 1978))"

	regex := strings.ReplaceAll(" ((Live in Los Angeles, 1978))", "(", "\\(")
	regex += "$"
	fmt.Printf("regex: %s\n", regex)

	replace := ""

	replaceRegexp := regexp.MustCompile(regex)
	fmt.Printf("result: %v\n", replaceRegexp.ReplaceAllString(name, replace))*/

	URL := "http://m.media-amazon.com/images/I/81iqFKrMDjL._SL1500_.jpg"
	matched, _ := regexp.MatchString("^https?://", URL)
	fmt.Printf("result: %v\n", matched)

	//--------------------------------------------------------------------------
	//title := "harpoon ft. bloody sunday [remix]"
	//title := "harpoon ft. bloody sunday (remix)"
	//title := "harpoon ft. bloody sunday"
	title := "Trey's Song (Feat. H.R. Of Bad Brains)"
	titleRegex := "([fF]eaturing|[fF]eat\\.|[fF]t\\.?)"
	titleRe := regexp.MustCompile(titleRegex)
	replacement := "feat"
	titleResult := titleRe.ReplaceAllString(title, replacement)
	fmt.Printf("title: %s, title regex: %v, replacement: %s, titleResult: %v\n", title, titleRegex, replacement, titleResult)

	//parenRegex := " (feat [^(\\[]+)([(\\[])"
	parenRegex := " (feat [^(\\[]+)( [()]?|$)"
	replacement = " ($1)$2"
	parenRe := regexp.MustCompile(parenRegex)
	parenResult := parenRe.ReplaceAllString(titleResult, replacement)
	fmt.Printf("title: %s, parenRegex: %v, replacement: %s, parenResult: %v\n", title, titleRegex, replacement, parenResult)

	//--------------------------------------------------------------------------
	excludeFileRegex := "\\.(nfo|sfv|m3u)$|it|t[aA]boo"
	re, _ := regexp.Compile(excludeFileRegex)
	matchString := "tAboo"
	fmt.Printf("excludeFileRegex: %s, result :%v\n", excludeFileRegex, re.MatchString(matchString))

	//--------------------------------------------------------------------------
	//if matched, _ := regexp.MatchString("( (\\(|\\[)?([fF]eaturing|[fF]eat\\.?|[fF]t\\.?)) ", " I Left My Wallet In El Segundo "); matched {
	//if matched, _ := regexp.MatchString("( (\\(|\\[)?([fF]eaturing|[fF]eat\\.?|[fF]t\\.?)) ", " (Feat. "); matched {
	if matched, _ := regexp.MatchString(" [(\\[]?([fF]eaturing|[fF]eat\\.?|[fF]t\\.?) ", "Scenario (Feat. Charlie"); matched {
		fmt.Printf("\nfeat Matched\n")

	} else {
		fmt.Printf("\nfeat Not Matched\n")

	}

	//replaceRe := " (feat [^(\\[]+)([()]?|$)"
	replaceRe := "( [(\\[]?)([fF]eaturing|[fF]eat\\.?|[fF]t\\.?) "
	replaceRegexp, _ := regexp.Compile(replaceRe)
	value := "Award Tour Feat. Trugoy"
	replaceValue := "${1}feat "
	result1 := replaceRegexp.ReplaceAllString(value, replaceValue)
	fmt.Printf("result1: %s\n", result1)

	replaceRe2 := " (\\]?feat [^\\[]+)([()]?|$)"
	replaceRegexp, _ = regexp.Compile(replaceRe2)
	replaceValue2 := " ($1)$2"
	result2 := replaceRegexp.ReplaceAllString(result1, replaceValue2)
	fmt.Printf("result2: %s\n", result2)

	track := "11. You Are (Bonus Track)"
	bonusTrackRe := " \\(bonus track\\)"
	regex := regexp.MustCompile(bonusTrackRe)
	lowFValue := strings.ToLower(track)
	fValue := ""
	if regex.MatchString(lowFValue) {
		fValue = regex.ReplaceAllString(lowFValue, "")
	}
	fmt.Printf("bonus track fValue: %s\n", fValue)

}
