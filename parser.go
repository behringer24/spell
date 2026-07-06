package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	currentNavpoint       [7]NavpointAdder
	currentImageId        int

	firstparagraph bool = true
	listStack      []listFrame
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
	reMeta       = regexp.MustCompile(`\$\[(title|author|series|set|entry|uuid|language|quotes|date|rights|source|relation|type)\]\(([^\)]+)\)`)
	reCover      = regexp.MustCompile(`\!\[cover\]\(([^ \)]+)\s*(\"([^\"]*)\")?\)`)
	reImage      = regexp.MustCompile(`\!\[([^\]]*)\]\(([^ \)]+)\s*(\"([^\"]*)\")?\)`)
	reQuotes     = regexp.MustCompile(`(%"|"%|%'|'%)`)
	reBold       = regexp.MustCompile(`\*\*([^\*]+)\*\*`)
	reItalic     = regexp.MustCompile(`\*([^\*]+)\*`)
	reCode       = regexp.MustCompile("`([^`]+)`")
	reComment    = regexp.MustCompile(`(^|\s)//.*$`)
	// reListItem matches a bullet (-, *, +) or ordered (1. / 1)) list item.
	// Group 1: leading indentation, group 2: bullet marker (empty if ordered),
	// group 3: ordinal digits (empty if unordered), group 4: item content.
	reListItem = regexp.MustCompile(`^([ \t]*)(?:([-*+])|(\d{1,9})[.)])[ \t]+(.*)$`)
	// reLink matches an inline link [text](url) or [text](url "title").
	reLink       = regexp.MustCompile(`\[([^\]]+)\]\(([^ \)]+)\s*(?:"([^"]*)")?\)`)
	reLongDash   = regexp.MustCompile(`\s+(---)\s+`)
	reMidDash    = regexp.MustCompile(`\s+(--)\s+`)
	reThreeDots  = regexp.MustCompile(`(\.\.\.)`)
	reBlockQuote = regexp.MustCompile("\\s*```\\s*([a-zA-Z]*)")
	reNewline    = regexp.MustCompile(`\r?\n`)
)

// parseContext carries the book and base directory through the handler pipeline.
type parseContext struct {
	book               SpellBook
	baseDir            string
	customCSSPaths     []string // book-internal paths (e.g. ["css/a.css", "css/b.css"])
	currentChapterFile string   // filename of the chapter currently being rendered
	azw3Mode           bool     // true when producing AZW3: all chapters form one document, so cross-chapter hrefs are plain #id
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
		match: func(line string, _ bool) bool { return re.MatchString(line) },
		handle: func(_ *parseContext, line string, _ bool) (string, bool) {
			return re.ReplaceAllString(line, replacement), false
		},
	}
}

// staticResult creates a handler that returns a fixed HTML string for any matching line.
func staticResult(re *regexp.Regexp, result string) lineHandler {
	return lineHandler{
		match:  func(line string, _ bool) bool { return re.MatchString(line) },
		handle: func(_ *parseContext, _ string, _ bool) (string, bool) { return result, true },
	}
}

// backtickSegments splits line into alternating non-code and code segments.
// Even indices are outside backticks; odd indices are inside (including the backticks).
func backtickSegments(line string) []string {
	var segs []string
	for len(line) > 0 {
		tick := strings.Index(line, "`")
		if tick == -1 {
			segs = append(segs, line)
			break
		}
		segs = append(segs, line[:tick])
		line = line[tick:]
		end := strings.Index(line[1:], "`")
		if end == -1 {
			segs = append(segs, line)
			break
		}
		segs = append(segs, line[:end+2])
		line = line[end+2:]
	}
	return segs
}

// matchOutsideBackticks returns true if re matches anywhere outside inline backtick spans.
func matchOutsideBackticks(line string, re *regexp.Regexp) bool {
	for i, seg := range backtickSegments(line) {
		if i%2 == 0 && re.MatchString(seg) {
			return true
		}
	}
	return false
}

// replaceOutsideBackticks applies fn to every match of re, but only in segments of line
// that are outside inline backtick spans. Backtick-delimited segments are passed through unchanged.
func replaceOutsideBackticks(line string, re *regexp.Regexp, fn func([]string) string) string {
	var result strings.Builder
	for i, seg := range backtickSegments(line) {
		if i%2 == 0 {
			result.WriteString(re.ReplaceAllStringFunc(seg, func(m string) string {
				return fn(re.FindStringSubmatch(m))
			}))
		} else {
			result.WriteString(seg)
		}
	}
	return result.String()
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
		escapeHandler(),
		listCloseHandler(),
		blockquoteFenceHandler(),
		blockquoteContentHandler(),
		chapterHandler(),
		headlineHandler(),
		metaHandler(),
		coverHandler(),
		imageHandler(),
		indexOutputHandler(),
		tocOutputHandler(),
		anchorDefHandler(),
		anchorLinkHandler(),
		indexEntryHandler(),
		linkHandler(),
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
func addChapter(ctx *parseContext, chapterTitle string, chapterNumber int, chapterContent strings.Builder) error {
	filename := fmt.Sprintf("xhtml/chapter_%05d.xhtml", chapterNumber)

	var content string
	if ctx.azw3Mode {
		// AZW3: mobi chunks must be plain HTML fragments, not full documents.
		// CSS is applied globally via AddStylesheet; no wrapper needed.
		content = chapterContent.String()
	} else {
		var customCSSLinks string
		for _, p := range ctx.customCSSPaths {
			customCSSLinks += "\n\t\t<link rel=\"stylesheet\" href=\"../" + p + "\"/>"
		}
		content = `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
    <head>
        <meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>
        <title>` + chapterTitle + `</title>` + customCSSLinks + `
    </head>
    <body>
	` + chapterContent.String() + `
	</body>
</html>`
	}
	_, err := ctx.book.AddXHTML(filename, chapterTitle, content, 10)
	if err != nil {
		return err
	}
	logMsg(LogDefault, "Add chapter %s as %s", chapterTitle, filename)
	return nil
}

func addCover(book SpellBook, imageFile string, baseDir string, addCoverPage bool) error {
	currentImage := fmt.Sprintf("img/cover%s", filepath.Ext(imageFile))
	imageID, err := book.AddImageFile(filepath.Join(baseDir, imageFile), currentImage)
	if err != nil {
		return err
	}
	book.SetCoverImage(imageID)
	logMsg(LogVerbose, "Added cover image %s: %s", imageID, currentImage)

	if addCoverPage {
		isAZW3 := strings.Contains(imageID, "kindle:")
		imgSrc := "../" + currentImage
		if isAZW3 {
			imgSrc = imageID
		}
		var htmlContent string
		if isAZW3 {
			htmlContent = `<div style="text-align:center;padding:0;margin:0;"><img src="` + imgSrc + `" style="max-width:100%;"/></div>`
		} else {
			htmlContent = `<?xml version="1.0" encoding="utf-8"?>
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
                <image width="1240" height="1752" xlink:href="` + imgSrc + `"/>
            </svg>
        </div>
	</body>
</html>`
		}
		_, err = book.AddXHTML("xhtml/cover.xhtml", "Cover", htmlContent, 1)
		if err != nil {
			return err
		}
		logMsg(LogVerbose, "Add cover file cover.xhtml")
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
	return line
}

// isBlockElement returns true when s is a block-level HTML element that must
// not be wrapped in a <p> tag.
func isBlockElement(s string) bool {
	for _, prefix := range []string{
		"<h1", "<h2", "<h3", "<h4", "<h5", "<h6",
		"<blockquote", "<hr", "<MBP:", "<div",
		"<ul", "</ul>", "<ol", "</ol>", "<li", "</li",
		"<section",
	} {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

// Parse chapters and other Markdown commands
func parseMarkdown(book SpellBook, content string, baseDir string, customCSSFile string) error {
	// Pass 1: collect all anchors and index entries before rendering.
	resetAnchors()
	listStack = nil
	scanAnchorsAndIndex(content)

	// split contents by lines
	lines := reNewline.Split(content, -1)

	addDefaultTemplate(book)

	_, isAZW3 := book.(*azw3Book)
	ctx := &parseContext{book: book, baseDir: baseDir, azw3Mode: isAZW3}
	if !isAZW3 {
		ctx.customCSSPaths = append(ctx.customCSSPaths, "css/_spellDefault.css")
	}
	for _, cssFile := range strings.FieldsFunc(customCSSFile, func(r rune) bool { return r == ',' }) {
		cssFile = strings.TrimSpace(cssFile)
		cssContent, err := os.ReadFile(cssFile)
		if err != nil {
			logMsg(LogDefault, "WARNING: Could not read custom CSS file '%s': %v", cssFile, err)
			continue
		}
		internalPath := "css/" + filepath.Base(cssFile)
		book.AddStylesheet(internalPath, string(cssContent))
		ctx.customCSSPaths = append(ctx.customCSSPaths, internalPath)
		logMsg(LogDefault, "Added custom stylesheet %s", internalPath)
	}

	// Pass 2: render.
	// Consecutive non-blank lines are accumulated into a single <p>; a blank line
	// or a block-level element flushes the accumulator first.
	var paraAccum []string

	flushParagraph := func() {
		if len(paraAccum) == 0 {
			return
		}
		text := strings.Join(paraAccum, " ")
		paraAccum = nil
		if firstparagraph {
			firstparagraph = false
			currentChapterContent.WriteString("<p class=\"firstparagraph\">" + text + "</p>\n")
		} else {
			currentChapterContent.WriteString("<p>" + text + "</p>\n")
		}
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			flushParagraph()
			continue
		}
		newline := parseLine(ctx, line, false)
		trimmed := strings.TrimSpace(newline)
		if trimmed == "" {
			continue
		}
		if isBlockElement(trimmed) {
			flushParagraph()
			currentChapterContent.WriteString(newline)
		} else {
			paraAccum = append(paraAccum, trimmed)
		}
	}
	flushParagraph()
	// Close any list still open at end of input.
	currentChapterContent.WriteString(closeAllLists())

	// Add last chapter
	if currentChapterTitle != "" {
		addChapter(ctx, currentChapterTitle, currentChapterNumber[1], currentChapterContent)
	}

	return nil
}
