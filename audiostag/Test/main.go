package main

import (
	"fmt"
	"regexp"
	"strings"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func manualOverwrite(artist string) {
	newValue := "Four"
	//newValue := ""
	artistParts := strings.Split(artist, ",")
	artistPartsLen := len(artistParts)

	fmt.Printf("\n\nartist: %s, artistParts[0]: %s\n", artist, artistParts[0])
	var artistBuilder strings.Builder
	for i := 0; i < artistPartsLen; i++ {
		if newValue != strings.TrimSpace(artistParts[i]) {
			if artistBuilder.Len() > 0 {
				//fmt.Printf("i: %d, artistParts[%d]: %v, artistParts[%d]: %v, artistBuilder: %s\n", i, i, artistParts[i], i+1, artistParts[i+1], artistBuilder.String())
				if i == artistPartsLen-1 || (i == artistPartsLen-2 && strings.TrimSpace(artistParts[i+1]) == newValue) {
					artistBuilder.WriteString(" & ")

				} else {
					artistBuilder.WriteString(", ")

				}

			}

			artistBuilder.WriteString(strings.TrimSpace(artistParts[i]))

		}

	}
	fmt.Printf("artistBuilder: %v, newValue: %s\n", artistBuilder.String(), newValue)
}

func main() {
	titler := cases.Title(language.English)
	album := "Celestial Crown/barael's Blade (2004 Demo Version)"

	regTitle := regexp.MustCompile("^[A-Z][a-z0-9]+\\s+")
	if regTitle.MatchString(album) {
		fmt.Printf("album: %v\n", titler.String(album))

	}


	//manualOverwrite("One, Two, Three, Four")
	manualOverwrite("One, Two, Three, Four")

}
