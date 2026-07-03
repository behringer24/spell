package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/behringer24/argumentative"
)

const (
	title       = "spell"
	description = "Smart Processing and Enhanced Lightweight Layout. Command line parser for converting enhanced markdown to epub."
	version     = "v1.5.1"
)

var (
	inFileName    *string
	outFileName   *string
	outputFormat  *string
	generateCover *bool
	customCSS     *string
	showHelp      *bool
	showVer       *bool
	verboseFlag   *bool
)

const (
	LogDefault = iota
	LogVerbose
)

func logMsg(level int, format string, args ...any) {
	if level == LogDefault || *verboseFlag {
		log.Printf(format, args...)
	}
}

// Function for reading a file
func readFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
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
			logMsg(LogDefault, "Error including %s with URI %s", matches[0], matches[1])
			return match // Fallback: if the pattern is wrong or not an md file
		}

		includeContent, err := readFile(filepath.Join(baseDir, matches[1]))
		if err != nil {
			logMsg(LogDefault, "Error including %s with URI %s: %v", matches[0], matches[1], err)
			return match
		}

		logMsg(LogVerbose, "Including markdown file %s (%s)", matches[1], matches[3])
		return includeContent
	})
}

// Process Markdown file
func processMarkdownFile(book SpellBook, filePath string, customCSSFile string) error {
	// Read markdown file
	content, err := readFile(filePath)
	if err != nil {
		return err
	}

	// Replace all includes
	baseDir := filepath.Dir(filePath)
	content = replaceAllIncludes(content, baseDir)

	// Parse markdown
	err = parseMarkdown(book, content, baseDir, customCSSFile)
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
	generateCover = flags.Flags().AddBool("cover", "c", "Generate cover page. This is normally not recommended")
	outputFormat = flags.Flags().AddString("format", "f", false, "epub3", "Output format: epub2, epub3, or azw3")
	customCSS = flags.Flags().AddString("style", "s", false, "", "Comma-separated list of CSS files to include")
	verboseFlag = flags.Flags().AddBool("verbose", "V", "Enable verbose logging")
	inFileName = flags.Flags().AddPositional("infile", true, "", "File to read from")
	outFileName = flags.Flags().AddPositional("outfile", false, "", "File to write to (default: ./ebook.epub or ./ebook.azw3)")

	err := flags.Parse(os.Args)
	if *showHelp {
		flags.Usage(title, description, nil)
		os.Exit(0)
	} else if *showVer {
		fmt.Print(strings.ToUpper(title), " version: ", version)
		os.Exit(0)
	} else if *outputFormat != "epub2" && *outputFormat != "epub3" && *outputFormat != "azw3" {
		fmt.Print("Error: format must be epub2, epub3, or azw3")
		os.Exit(1)
	} else if err != nil {
		flags.Usage(title, description, err)
		os.Exit(1)
	}
}

func main() {
	// Use argumentative as command line parser
	parseArgs()

	// Apply default output filename based on format when not specified
	if *outFileName == "" {
		if *outputFormat == "azw3" {
			*outFileName = "./ebook.azw3"
		} else {
			*outFileName = "./ebook.epub"
		}
	}

	// Create the book for the requested output format
	var book SpellBook
	switch *outputFormat {
	case "epub2":
		book = newEpubBook(2.0)
	case "epub3":
		book = newEpubBook(3.0)
	case "azw3":
		book = newAZW3Book()
	}

	// Process input file
	err := processMarkdownFile(book, *inFileName, *customCSS)
	if err != nil {
		log.Fatalf("Error processing file '%s': %v", *inFileName, err)
	}

	// Write output
	err = book.Write(*outFileName)
	if err != nil {
		log.Fatalf("Error writing file '%s': %v", *outFileName, err)
	}

	fmt.Printf("File '%s' created successfully!\n", *outFileName)
}
