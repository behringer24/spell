package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/behringer24/argumentative"
	"github.com/russross/blackfriday/v2"
	"github.com/writingtoole/epub"
)

const (
	title       = "spell"
	description = "A small demonstration"
	version     = "v0.0.1"
)

var (
	inFileName  *string
	outFileName *string
	generateVer *string
	showHelp    *bool
	showVer     *bool
)

// Funktion zum Einlesen einer Datei
func readFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Konvertiert Markdown-Inhalt in HTML
func markdownToHTML(content string) string {
	return string(blackfriday.Run([]byte(content)))
}

func addChapter(book *epub.EPub, chapterTitle string, chapterNumber int, chapterContent strings.Builder) error {
	//htmlContent := markdownToHTML(chapterContent.String())
	htmlContent := chapterContent.String()
	filename := fmt.Sprintf("chapter_%05d.xhtml", chapterNumber)
	_, err := book.AddXHTML(filename, htmlContent, 10)
	if err != nil {
		return err
	}
	log.Printf("Add chapter %s as %s", chapterTitle, filename)
	return nil
}

// Funktion zum Verarbeiten der Kapitel und anderer Markdown-Kommandos
func parseMarkdown(book *epub.EPub, content string) error {
	lines := strings.Split(content, "\r\n")
	var currentChapterContent strings.Builder
	var currentChapterTitle string
	var currentChapterNumber [7]int
	var currentNavpoint [7]*epub.Navpoint
	chapterRegex := regexp.MustCompile(`^\s*(#)\s*([^#]+)$`)
	headlinesRegex := regexp.MustCompile(`^\s*(#{2,6})\s*([^#]+)$`)
	dividerRegex := regexp.MustCompile(`^\s*([\*,\-,_]\s*)+$`)

	for _, line := range lines {
		if chapterRegex.MatchString(line) {
			// Chapter starting with one # char
			if currentChapterTitle != "" {
				addChapter(book, currentChapterTitle, currentChapterNumber[1], currentChapterContent)
			}

			// New chapter headline
			matches := chapterRegex.FindStringSubmatch(line)
			currentChapterTitle = matches[2]
			currentChapterContent.Reset()
			currentChapterNumber[1]++
			filename := fmt.Sprintf("chapter_%05d.xhtml", currentChapterNumber[1])
			currentNavpoint[1] = book.AddNavpoint(currentChapterTitle, filename, 10)

			currentChapterContent.WriteString("<h1>" + matches[2] + "</h1>\n")
		} else if headlinesRegex.MatchString(line) {
			// Headline with 2 or more #
			matches := headlinesRegex.FindStringSubmatch(line)
			chapterLevel := strings.Count(matches[1], "#")
			currentChapterNumber[chapterLevel]++
			currentChapterLabel := fmt.Sprintf("label%d_%d", chapterLevel, currentChapterNumber[chapterLevel])
			currentChapterContent.WriteString(fmt.Sprintf("<h%d id=\"%s\">%s</h%d>\n", chapterLevel, currentChapterLabel, matches[2], chapterLevel))
			if currentNavpoint[chapterLevel-1] != nil {
				anchorname := fmt.Sprintf("chapter_%05d.xhtml#%s", currentChapterNumber[1], currentChapterLabel)
				currentNavpoint[chapterLevel] = currentNavpoint[chapterLevel-1].AddNavpoint(matches[2], anchorname, 0)
				log.Printf("Add subchapter %s as %s", matches[2], anchorname)
			} else {
				log.Printf("Subchapter %s outside chapter", matches[2])
			}
		} else if dividerRegex.MatchString(line) {
			currentChapterContent.WriteString("<hr/>\n")
		} else {
			// Normal line just add if not empty
			if strings.Compare(strings.TrimSpace(line), "") != 0 {
				currentChapterContent.WriteString("<p>" + line + "</p>\n")
			}
		}
	}

	// Füge das letzte Kapitel hinzu
	if currentChapterTitle != "" {
		addChapter(book, currentChapterTitle, currentChapterNumber[1], currentChapterContent)
	}

	return nil
}

// Funktion zum Verarbeiten einer Markdown-Datei
func processMarkdownFile(book *epub.EPub, filePath string) error {
	// Lese den Inhalt der Markdown-Datei ein
	content, err := readFile(filePath)
	if err != nil {
		return err
	}

	// Füge Kapitel basierend auf den Markdown-Überschriften hinzu
	err = parseMarkdown(book, content)
	if err != nil {
		return err
	}

	return nil
}

func parseArgs() {
	flags := &argumentative.Flags{}
	showHelp = flags.Flags().AddBool("help", "h", "Show this help text")
	showVer = flags.Flags().AddBool("version", "v", "Show version information")
	generateVer = flags.Flags().AddString("epub", "e", false, "3", "Generate epub version 2 or 3")
	inFileName = flags.Flags().AddPositional("infile", true, "", "File to read from")
	outFileName = flags.Flags().AddPositional("outfile", false, "./ebook.epub", "File to write to")

	err := flags.Parse(os.Args)
	if *showHelp {
		flags.Usage(title, description, nil)
		os.Exit(0)
	} else if *showVer {
		fmt.Print(title, "version", version)
		os.Exit(0)
	} else if strings.Compare(*generateVer, "2") != 0 && strings.Compare(*generateVer, "3") != 0 {
		fmt.Print("Error, epub version has to be 2 or 3")
		os.Exit(1)
	} else if err != nil {
		flags.Usage(title, description, err)
		os.Exit(1)
	}
}

func main() {
	// Verwende argumentative für die Kommandozeilenparameter
	parseArgs()

	// Erstelle ein neues EPUB-Buch
	book := epub.New()

	// Setze Metadaten für das EPUB
	book.SetTitle("Markdown to EPUB")
	book.AddAuthor("Unknown Author")

	// Verarbeite alle angegebenen Markdown-Dateien
	err := processMarkdownFile(book, *inFileName)
	if err != nil {
		log.Fatalf("Fehler beim Verarbeiten der Datei '%s': %v", *inFileName, err)
	}

	// EPUB speichern
	epubVersion, _ := strconv.Atoi(*generateVer)
	book.SetVersion(float64(epubVersion))

	err = book.Write(*outFileName)
	if err != nil {
		log.Fatalf("Fehler beim Speichern der EPUB-Datei '%s': %v", *outFileName, err)
	}

	fmt.Printf("EPUB-Datei '%s' erfolgreich erstellt!\n", *outFileName)
}
