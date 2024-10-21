package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/behringer24/argumentative"
	"github.com/russross/blackfriday/v2"
	"github.com/writingtoole/epub"
)

const (
	title       = "spell"
	description = "Smart Processing and Enhanced Lightweight Layout"
	version     = "v0.0.3"
)

var (
	inFileName  *string
	outFileName *string
	generateVer *string
	showHelp    *bool
	showVer     *bool
)

// Function for reading a file
func readFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Convert markdown to html
func markdownToHTML(content string) string {
	return string(blackfriday.Run([]byte(content)))
}

// Replace all includes of md files using markdown syntax for images like
// ![include](uri/uri.md "text") or
// ![include](uri/uri.md)
// text is optional and ignored, you can use it as internal reference
func replaceAllIncludes(content string, baseDir string) string {
	commandRegex := regexp.MustCompile(`\!\[include\]\(([^ \)]+)\s*(\"([^\"]*)\")?\)`)
	return commandRegex.ReplaceAllStringFunc(content, func(match string) string {
		// Extract includes and parameters
		matches := commandRegex.FindStringSubmatch(match)
		if len(matches) < 2 && strings.Compare(filepath.Ext(matches[2]), ".md") != 0 {
			log.Printf("Error including %s with URI %s", matches[0], matches[1])
			return match // Fallback: if the pattern is wrong or not an md file
		}

		includeContent, err := readFile(filepath.Join(baseDir, matches[1]))
		if err != nil {
			log.Printf("Error including %s with URI %s: %v", matches[0], matches[1], err)
			return match
		}

		log.Printf("Including markdown file %s (%s)", matches[1], matches[3])
		return includeContent
	})
}

// Process Markdown file
func processMarkdownFile(book *epub.EPub, filePath string) error {
	// Read markdown file
	content, err := readFile(filePath)
	if err != nil {
		return err
	}

	// Replace all includes
	baseDir := filepath.Dir(filePath)
	content = replaceAllIncludes(content, baseDir)

	// Parse markdown
	err = parseMarkdown(book, content, baseDir)
	if err != nil {
		return err
	}

	return nil
}

// Parse command line parameters
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
		fmt.Print(strings.ToUpper(title), " version: ", version)
		os.Exit(0)
	} else if strings.Compare(*generateVer, "2") != 0 && strings.Compare(*generateVer, "3") != 0 {
		fmt.Print("Error: epub version has to be 2 or 3")
		os.Exit(1)
	} else if err != nil {
		flags.Usage(title, description, err)
		os.Exit(1)
	}
}

func main() {
	// Use argumentative as command line parser
	parseArgs()

	// Create new empty epub book
	book := epub.New()

	// Process input file
	err := processMarkdownFile(book, *inFileName)
	if err != nil {
		log.Fatalf("Fehler beim Verarbeiten der Datei '%s': %v", *inFileName, err)
	}

	// Save epub
	epubVersion, _ := strconv.Atoi(*generateVer)
	book.SetVersion(float64(epubVersion))
	err = book.Write(*outFileName)
	if err != nil {
		log.Fatalf("Fehler beim Speichern der EPUB-Datei '%s': %v", *outFileName, err)
	}

	fmt.Printf("EPUB-Datei '%s' erfolgreich erstellt!\n", *outFileName)
}
