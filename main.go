package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

// Chapter represents information about a chapter
type Chapter struct {
	ChapterNumber   int    `json:"chapter_number"`
	VersesCount     int    `json:"verses_count"`
	Name            string `json:"name"`
	Translation     string `json:"translation,omitempty"`
	Transliteration string `json:"transliteration,omitempty"`
	Meaning         struct {
		En string `json:"en,omitempty"`
		Hi string `json:"hi,omitempty"`
	} `json:"meaning,omitempty"`
	Summary struct {
		En string `json:"en,omitempty"`
		Hi string `json:"hi,omitempty"`
	} `json:"summary,omitempty"`
}

// Sloka represents a verse
type Sloka struct {
	ID              string `json:"_id"`
	Chapter         int    `json:"chapter"`
	Verse           int    `json:"verse"`
	Slok            string `json:"slok"`
	Transliteration string `json:"transliteration"`
	Tej             struct {
		Author string `json:"author"`
		Ht     string `json:"ht"`
	} `json:"tej"`
	Siva struct {
		Author string `json:"author"`
		Et     string `json:"et"`
		Ec     string `json:"ec"`
	} `json:"siva"`
	Purohit struct {
		Author string `json:"author"`
		Et     string `json:"et"`
	} `json:"purohit"`
	Chinmay struct {
		Author string `json:"author"`
		Hc     string `json:"hc"`
	} `json:"chinmay"`
	San struct {
		Author string `json:"author"`
		Et     string `json:"et"`
	} `json:"san"`
	Adi struct {
		Author string `json:"author"`
		Et     string `json:"et"`
	} `json:"adi"`
}

// AllSlokas stores all the verses
type AllSlokas struct {
	Chapters []Chapter `json:"chapters"`
	Slokas   []Sloka   `json:"slokas"`
}

//go:embed gita.json
var gitaFS embed.FS // embed gita.json

// Translator constants
const (
	Siva    = "siva"
	Purohit = "purohit"
	Adi     = "adi"
	San     = "san"
	Tej     = "tej"
	Chinmay = "chinmay"
)

const displayWidth = 70 // max line width for wrapping

// wrapText wraps text to fit the terminal width
func wrapText(text string) string {
	var result strings.Builder
	current := 0

	words := strings.Fields(text)
	for i, word := range words {
		wordLen := utf8.RuneCountInString(word)
		if current+wordLen+1 > displayWidth && current > 0 {
			result.WriteString("\n")
			current = 0
		}
		if current > 0 {
			result.WriteString(" ")
			current++
		}
		result.WriteString(word)
		current += wordLen

		if i < len(words)-1 && strings.ContainsAny(word, ".!?") &&
			!strings.HasPrefix(words[i+1], ")") &&
			!strings.HasPrefix(words[i+1], ",") {
			result.WriteString("\n")
			current = 0
		}
	}

	return result.String()
}

// ANSI styling
const (
	Bold  = "\033[1m"
	Dim   = "\033[2m"
	Reset = "\033[0m"
)

func main() {
	// CLI flags
	translationSource := flag.String("translation", "siva", "Translation source (siva, purohit, adi, san, tej, chinmay)")
	includeChapter := flag.Bool("chapter-info", false, "Show chapter information")
	chapterFlag := flag.Int("c", 0, "Specific chapter number (use with -v)")
	verseFlag := flag.Int("v", 0, "Specific verse number (use with -c)")
	flag.Parse()

	// validate translation source
	validSources := []string{Siva, Purohit, Adi, San, Tej, Chinmay}
	validSource := false
	for _, source := range validSources {
		if *translationSource == source {
			validSource = true
			break
		}
	}
	if !validSource {
		fmt.Printf("Invalid translation source: %s\n", *translationSource)
		fmt.Println("Valid sources: siva, purohit, adi, san, tej, chinmay")
		os.Exit(1)
	}

	// read embedded JSON file
	data, err := gitaFS.ReadFile("gita.json")
	if err != nil {
		fmt.Printf("Error reading embedded data: %v\n", err)
		os.Exit(1)
	}

	// parse JSON into structs
	var allSlokas AllSlokas
	err = json.Unmarshal(data, &allSlokas)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	var selectedSloka Sloka

	// if specific verse requested
	if *chapterFlag > 0 && *verseFlag > 0 {
		found := false
		for _, sloka := range allSlokas.Slokas {
			if sloka.Chapter == *chapterFlag && sloka.Verse == *verseFlag {
				selectedSloka = sloka
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("Chapter %d, Verse %d not found.\n", *chapterFlag, *verseFlag)
			os.Exit(1)
		}
	} else {
		// pick random sloka
		rand.Seed(time.Now().UnixNano())
		if len(allSlokas.Slokas) == 0 {
			fmt.Println("No slokas found in the JSON data.")
			os.Exit(1)
		}
		selectedSloka = allSlokas.Slokas[rand.Intn(len(allSlokas.Slokas))]
	}

	fmt.Println()

	// show chapter info if requested
	if *includeChapter {
		for _, chapter := range allSlokas.Chapters {
			if chapter.ChapterNumber == selectedSloka.Chapter {
				fmt.Printf("%sChapter %d: %s%s\n", Bold, chapter.ChapterNumber, chapter.Name, Reset)
				if chapter.Translation != "" {
					fmt.Printf("(%s)\n", chapter.Translation)
				}
				if chapter.Meaning.En != "" {
					fmt.Println(wrapText("Meaning: " + chapter.Meaning.En))
				}
				fmt.Println()
				break
			}
		}
	}

	// display chapter and verse header
	fmt.Printf("%sChapter %d, Verse %d%s\n\n", Bold, selectedSloka.Chapter, selectedSloka.Verse, Reset)

	// print sanskrit
	sanskritLines := strings.Split(selectedSloka.Slok, "\n")
	for _, line := range sanskritLines {
		if strings.TrimSpace(line) != "" {
			fmt.Println(wrapText(strings.TrimSpace(line)))
		}
	}
	fmt.Println()

	// print transliteration
	transLines := strings.Split(selectedSloka.Transliteration, ".")
	for _, line := range transLines {
		if strings.TrimSpace(line) != "" {
			fmt.Println(wrapText(strings.TrimSpace(line)))
		}
	}
	fmt.Println()

	// pick translation text and author
	var translationText, author string
	switch *translationSource {
	case Siva:
		translationText = selectedSloka.Siva.Et
		author = selectedSloka.Siva.Author
	case Purohit:
		translationText = selectedSloka.Purohit.Et
		author = selectedSloka.Purohit.Author
	case Adi:
		translationText = selectedSloka.Adi.Et
		author = selectedSloka.Adi.Author
	case San:
		translationText = selectedSloka.San.Et
		author = selectedSloka.San.Author
	case Tej:
		translationText = selectedSloka.Tej.Ht
		author = selectedSloka.Tej.Author
	case Chinmay:
		translationText = selectedSloka.Chinmay.Hc
		author = selectedSloka.Chinmay.Author
	}

	// print translation
	fmt.Println(wrapText(translationText))
	fmt.Printf("%s(%s)%s\n", Dim, author, Reset)

	fmt.Println()
}
