package main

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/writingtoole/epub"
)

const (
	BLOCKTYPE_NONE int = 0
	BLOCKTYPE_CODE int = 1
	BLOCKTYPE_CITE int = 2
	BLOCKTYPE_NOTE int = 3
	BLOCKTYPE_INFO int = 4
	BLOCKTYPE_WARN int = 5
)

var (
	currentChapterContent strings.Builder
	currentChapterTitle   string
	currentChapterNumber  [7]int
	currentNavpoint       [7]*epub.Navpoint
	currentImageId        int

	firstparagraph bool = true
	inUlList       bool
	inBlockType    int = 0

	laquo  = "\""
	raquo  = "\""
	lsaquo = "'"
	rsaquo = "'"
)

var (
	reChapter    = regexp.MustCompile(`^\s*(#)\s*([^#]+)$`)
	reHeadlines  = regexp.MustCompile(`^\s*(#{2,6})\s*([^#]+)$`)
	reDivider    = regexp.MustCompile(`^\s*([\*\-]\s*)+$`)
	rePagebreak  = regexp.MustCompile(`^\s*(_\s*)+$`)
	reMeta       = regexp.MustCompile(`\$\[(title|author|series|set|entry|uuid|language|quotes)\]\(([^\)]+)\)`)
	reCover      = regexp.MustCompile(`\!\[cover\]\(([^ \)]+)\s*(\"([^\"]*)\")?\)`)
	reImage      = regexp.MustCompile(`\!\[([^\]]*)\]\(([^ \)]+)\s*(\"([^\"]*)\")?\)`)
	reQuotes     = regexp.MustCompile(`(%"|"%|%'|'%)`)
	reBold       = regexp.MustCompile(`\*\*([^\*]+)\*\*`)
	reItalic     = regexp.MustCompile(`\*([^\*]+)\*`)
	reCode       = regexp.MustCompile("`([^`]+)`")
	reComment    = regexp.MustCompile(`(^|\s)//.*$`)
	reUlList     = regexp.MustCompile(`^\s*-\s*(.*)$`)
	reLongDash   = regexp.MustCompile(`\s+(---)\s+`)
	reMidDash    = regexp.MustCompile(`\s+(--)\s+`)
	reThreeDots  = regexp.MustCompile(`(\.\.\.)`)
	reBlockQuote = regexp.MustCompile("\\s*```\\s*([a-zA-Z]*)")
	reNewline    = regexp.MustCompile(`\r?\n`)
)

// parseContext carries the epub book and base directory through the handler pipeline.
type parseContext struct {
	book    *epub.EPub
	baseDir string
}

// lineHandler matches and transforms one line.
// handle returns (output, done): if done the output is the final result,
// otherwise output is the transformed line to pass back through the pipeline.
type lineHandler struct {
	match  func(line string, insideBlock bool) bool
	handle func(ctx *parseContext, line string, insideBlock bool) (string, bool)
}

// replaceAndRecurse creates a handler that applies a regex substitution and continues the pipeline.
func replaceAndRecurse(re *regexp.Regexp, replacement string) lineHandler {
	return lineHandler{
		match:  func(line string, _ bool) bool { return re.MatchString(line) },
		handle: func(_ *parseContext, line string, _ bool) (string, bool) { return re.ReplaceAllString(line, replacement), false },
	}
}

// staticResult creates a handler that returns a fixed HTML string for any matching line.
func staticResult(re *regexp.Regexp, result string) lineHandler {
	return lineHandler{
		match:  func(line string, _ bool) bool { return re.MatchString(line) },
		handle: func(_ *parseContext, _ string, _ bool) (string, bool) { return result, true },
	}
}

// replaceEachAndRecurse creates a handler that transforms each regex match via fn and continues the pipeline.
// fn receives (ctx, captureGroup1).
func replaceEachAndRecurse(re *regexp.Regexp, fn func(*parseContext, string) string) lineHandler {
	return lineHandler{
		match: func(line string, _ bool) bool { return re.MatchString(line) },
		handle: func(ctx *parseContext, line string, _ bool) (string, bool) {
			return re.ReplaceAllStringFunc(line, func(match string) string {
				return fn(ctx, re.FindStringSubmatch(match)[1])
			}), false
		},
	}
}

var handlers []lineHandler

func init() {
	handlers = []lineHandler{
		listCloseHandler(),
		blockquoteFenceHandler(),
		blockquoteContentHandler(),
		chapterHandler(),
		headlineHandler(),
		metaHandler(),
		coverHandler(),
		imageHandler(),
		dividerHandler(),
		pagebreakHandler(),
		quotesHandler(),
		listItemHandler(),
		boldHandler(),
		italicHandler(),
		codeSpanHandler(),
		commentHandler(),
		emDashHandler(),
		enDashHandler(),
		ellipsisHandler(),
	}
}

// Add a chapter file to the book
func addChapter(book *epub.EPub, chapterTitle string, chapterNumber int, chapterContent strings.Builder) error {
	//htmlContent := markdownToHTML(chapterContent.String())
	htmlContent := `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE xhtml>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
    <head>
        <meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>
        <title>` + chapterTitle + `</title>
		<link rel="stylesheet" href="../css/styles.css"/>
    </head>
    <body>
	` + chapterContent.String() + `
	</body>
</html>`
	filename := fmt.Sprintf("xhtml/chapter_%05d.xhtml", chapterNumber)
	_, err := book.AddXHTML(filename, htmlContent, 10)
	if err != nil {
		return err
	}
	log.Printf("Add chapter %s as %s", chapterTitle, filename)
	return nil
}

func addCover(book *epub.EPub, imageFile string, baseDir string, addCoverPage bool) error {
	currentImage := fmt.Sprintf("img/cover%s", filepath.Ext(imageFile))
	imageID, err := book.AddImageFile(filepath.Join(baseDir, imageFile), currentImage)
	if err != nil {
		return err
	} else {
		book.SetCoverImage(imageID)
		log.Printf("Added cover image %s: %s", imageID, currentImage)
	}

	if addCoverPage {
		htmlContent := `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE xhtml>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
    <head>
        <meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>
        <title>Cover</title>
		<style type="text/css">
            @page {padding: 0pt; margin:0pt}
            body { text-align: center; padding:0pt; margin: 0pt; }
        </style>
    </head>
    <body>
		<div>
            <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" width="100%" height="100%" viewBox="0 0 1240 1752" preserveAspectRatio="none">
                <image width="1240" height="1752" xlink:href="../` + currentImage + `"/>
            </svg>
        </div>
	</body>
</html>`
		_, err = book.AddXHTML("xhtml/cover.xhtml", htmlContent, 1)
		if err != nil {
			return err
		}
		log.Printf("Add cover file cover.xhtml")
	}
	return nil
}

func parseLine(ctx *parseContext, line string, insideBlock bool) string {
	for _, h := range handlers {
		if h.match(line, insideBlock) {
			output, done := h.handle(ctx, line, insideBlock)
			if done {
				return output
			}
			return parseLine(ctx, output, insideBlock)
		}
	}
	if !insideBlock && strings.TrimSpace(line) != "" {
		if firstparagraph {
			firstparagraph = false
			return "<p class=\"firstparagraph\">" + line + "</p>\n"
		}
		return "<p>" + line + "</p>\n"
	}
	return line
}

// Parse chapters and other Markdown commands
func parseMarkdown(book *epub.EPub, content string, baseDir string) error {
	// split contents by lines
	lines := reNewline.Split(content, -1)

	addDefaultTemplate(book)

	ctx := &parseContext{book: book, baseDir: baseDir}
	for _, line := range lines {
		newline := parseLine(ctx, line, false)
		if len(strings.TrimSpace(newline)) > 0 {
			currentChapterContent.WriteString(newline)
		}
	}

	// Add last chapter
	if currentChapterTitle != "" {
		addChapter(book, currentChapterTitle, currentChapterNumber[1], currentChapterContent)
	}

	return nil
}
