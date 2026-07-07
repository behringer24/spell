package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func blockquoteFenceHandler() lineHandler {
	return lineHandler{
		match: func(line string, insideBlock bool) bool { return reBlockQuote.MatchString(line) },
		handle: func(_ *parseContext, line string, _ bool) (string, bool) {
			if inBlockType > 0 {
				logMsg(LogVerbose, "blockQuote schließen")
				inBlockType = 0
				return "</blockquote>\n", true
			}
			matches := reBlockQuote.FindStringSubmatch(line)
			blocktype := "code"
			inBlockType = BLOCKTYPE_CODE
			if len(matches) == 2 && matches[1] != "" {
				blocktype = strings.ToLower(matches[1])
				switch blocktype {
				case "cite":
					inBlockType = BLOCKTYPE_CITE
				case "note":
					inBlockType = BLOCKTYPE_NOTE
				case "info":
					inBlockType = BLOCKTYPE_INFO
				case "warn":
					inBlockType = BLOCKTYPE_WARN
				}
			}
			logMsg(LogVerbose, "blockQuote opening: %s", blocktype)
			return fmt.Sprintf("<blockquote class=\"%s\">\n", blocktype), true
		},
	}
}

func blockquoteContentHandler() lineHandler {
	return lineHandler{
		match: func(line string, insideBlock bool) bool {
			return inBlockType != BLOCKTYPE_NONE && !insideBlock
		},
		handle: func(ctx *parseContext, line string, _ bool) (string, bool) {
			if inBlockType == BLOCKTYPE_CODE {
				logMsg(LogVerbose, "blockQuote CODE line")
				return fmt.Sprintf("%s</br>\n", line), true
			}
			logMsg(LogVerbose, "blockQuote non CODE but parsed line")
			return fmt.Sprintf("%s</br>\n", parseLine(ctx, line, true)), true
		},
	}
}

func chapterHandler() lineHandler {
	return lineHandler{
		match: func(line string, insideBlock bool) bool { return reChapter.MatchString(line) },
		handle: func(ctx *parseContext, line string, _ bool) (string, bool) {
			if currentChapterTitle != "" {
				appendPendingFootnotes(ctx)
				addChapter(ctx, currentChapterTitle, currentChapterNumber[1], currentChapterContent)
			}
			matches := reChapter.FindStringSubmatch(line)
			currentChapterTitle = parseLine(ctx, matches[2], true)
			currentChapterContent.Reset()
			currentChapterNumber[1]++
			filename := fmt.Sprintf("xhtml/chapter_%05d.xhtml", currentChapterNumber[1])
			ctx.currentChapterFile = filename
			// The first real chapter marks where body content begins.
			if !startReadingSet {
				ctx.book.SetStartReading(filename)
				startReadingSet = true
			}
			currentNavpoint[1] = ctx.book.AddNavpoint(currentChapterTitle, filename, 10)
			firstparagraph = true
			return fmt.Sprintf("<h1 id=\"label1_%d\">%s</h1>\n", currentChapterNumber[1], parseLine(ctx, matches[2], true)), true
		},
	}
}

// anchorDefHandler renders {#id} as an invisible span with that id.
// Matches inside backtick code spans are left untouched.
func anchorDefHandler() lineHandler {
	return lineHandler{
		match: func(line string, _ bool) bool { return matchOutsideBackticks(line, reAnchorDef) },
		handle: func(ctx *parseContext, line string, insideBlock bool) (string, bool) {
			out := replaceOutsideBackticks(line, reAnchorDef, func(sub []string) string {
				return fmt.Sprintf(`<span id="%s"></span>`, sub[1])
			})
			return parseLine(ctx, out, insideBlock), true
		},
	}
}

// anchorLinkHandler renders [text](#id) as an internal epub hyperlink.
// Matches inside backtick code spans are left untouched.
func anchorLinkHandler() lineHandler {
	return lineHandler{
		match: func(line string, _ bool) bool { return matchOutsideBackticks(line, reAnchorLink) },
		handle: func(ctx *parseContext, line string, insideBlock bool) (string, bool) {
			out := replaceOutsideBackticks(line, reAnchorLink, func(sub []string) string {
				href := resolveAnchorHref(ctx, sub[2], ctx.currentChapterFile)
				return fmt.Sprintf(`<a href="%s">%s</a>`, href, sub[1])
			})
			return parseLine(ctx, out, insideBlock), true
		},
	}
}

// indexEntryHandler renders %[displayTerm](indexname) or %[displayTerm](indexname|canonical)
// as a classed span with a stable id. The display term appears in the text; the canonical
// term (after |) is used as the index key for grouping variants like singular/plural.
// Matches inside backtick code spans are left untouched.
func indexEntryHandler() lineHandler {
	return lineHandler{
		match: func(line string, _ bool) bool { return matchOutsideBackticks(line, reIndexEntry) },
		handle: func(ctx *parseContext, line string, insideBlock bool) (string, bool) {
			out := replaceOutsideBackticks(line, reIndexEntry, func(sub []string) string {
				displayTerm, indexName, canonical := sub[1], sub[2], sub[3]
				if canonical == "" {
					canonical = displayTerm
				}
				key := indexName + "\x00" + canonical
				seq := indexCounters[key]
				indexCounters[key]++
				htmlID := fmt.Sprintf("idx-%s-%s-%d", sanitizeID(indexName), sanitizeID(canonical), seq)
				return fmt.Sprintf(`<span id="%s" class="index-entry" epub:type="index-term">%s</span>`, htmlID, displayTerm)
			})
			return parseLine(ctx, out, insideBlock), true
		},
	}
}

// indexOutputHandler renders %index[name] or %index[name](Title) as a new chapter.
// Entries are grouped by their canonical term; each group shows the canonical term
// as a label followed by links to every occurrence in the text.
func indexOutputHandler() lineHandler {
	return lineHandler{
		match: func(line string, _ bool) bool { return reIndexOutput.MatchString(line) },
		handle: func(ctx *parseContext, line string, _ bool) (string, bool) {
			sub := reIndexOutput.FindStringSubmatch(line)
			indexName, title := sub[1], sub[2]
			if title == "" {
				title = indexName
			}
			entries, ok := indexes[indexName]
			if !ok || len(entries) == 0 {
				logMsg(LogDefault, "WARNING: no index entries found for %q", indexName)
				return "", true
			}

			// Flush current chapter before starting the index chapter.
			if currentChapterTitle != "" {
				appendPendingFootnotes(ctx)
				addChapter(ctx, currentChapterTitle, currentChapterNumber[1], currentChapterContent)
				currentChapterTitle = ""
				currentChapterContent.Reset()
			}

			currentChapterNumber[1]++
			filename := fmt.Sprintf("xhtml/chapter_%05d.xhtml", currentChapterNumber[1])
			ctx.currentChapterFile = filename

			// Group entries by canonical term, preserving first-seen order.
			type group struct {
				canonical string
				entries   []indexEntry
			}
			seen := map[string]int{}
			var groups []group
			for _, e := range entries {
				if idx, exists := seen[e.canonical]; exists {
					groups[idx].entries = append(groups[idx].entries, e)
				} else {
					seen[e.canonical] = len(groups)
					groups = append(groups, group{canonical: e.canonical, entries: []indexEntry{e}})
				}
			}

			// indexHref returns the correct href for an index entry link.
			// In AZW3 mode all chapters form one document, so just use #id.
			indexHref := func(e indexEntry) string {
				if ctx.azw3Mode || e.chapterFile == filename {
					return "#" + e.htmlID
				}
				return "../" + e.chapterFile + "#" + e.htmlID
			}

			var body strings.Builder
			if ctx.azw3Mode {
				body.WriteString(fmt.Sprintf("<section>\n<h1 id=\"label1_%d\">%s</h1>\n<ul class=\"index-list\">\n", currentChapterNumber[1], title))
			} else {
				body.WriteString(fmt.Sprintf("<section epub:type=\"index\">\n<h1 id=\"label1_%d\">%s</h1>\n<ul epub:type=\"index-entry-list\" class=\"index-list\">\n", currentChapterNumber[1], title))
			}
			for i, g := range groups {
				if len(g.entries) == 1 {
					e := g.entries[0]
					if ctx.azw3Mode {
						body.WriteString(fmt.Sprintf("  <li><span>%s</span> <a href=\"%s\">1</a></li>\n", g.canonical, indexHref(e)))
					} else {
						body.WriteString(fmt.Sprintf("  <li epub:type=\"index-entry\"><span epub:type=\"index-term\">%s</span> <a epub:type=\"index-locator\" href=\"%s\">1</a></li>\n", g.canonical, indexHref(e)))
					}
				} else {
					// Multiple occurrences: list canonical term once, link each occurrence.
					if ctx.azw3Mode {
						body.WriteString(fmt.Sprintf("  <li><span class=\"index-canonical\">%s</span>\n    <ul>\n", g.canonical))
					} else {
						body.WriteString(fmt.Sprintf("  <li epub:type=\"index-entry\"><span epub:type=\"index-term\" class=\"index-canonical\">%s</span>\n    <ul epub:type=\"index-locator-list\">\n", g.canonical))
					}
					for j, e := range g.entries {
						body.WriteString(fmt.Sprintf("      <li><a href=\"%s\">%d</a></li>\n", indexHref(e), j+1))
					}
					body.WriteString("    </ul>\n  </li>\n")
				}
				_ = i
			}
			body.WriteString("</ul>\n</section>\n")

			var htmlContent string
			if ctx.azw3Mode {
				// AZW3: plain fragment only, no document wrapper.
				htmlContent = body.String()
			} else {
				var customCSSLinks string
				for _, p := range ctx.customCSSPaths {
					customCSSLinks += "\n\t\t<link rel=\"stylesheet\" href=\"../" + p + "\"/>"
				}
				htmlContent = `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
    <head>
        <meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>
        <title>` + title + `</title>` + customCSSLinks + `
    </head>
    <body>
	` + body.String() + `
	</body>
</html>`
			}
			if _, err := ctx.book.AddXHTML(filename, title, htmlContent, 10); err != nil {
				logMsg(LogDefault, "ERROR: writing index chapter %s: %v", filename, err)
			}
			currentNavpoint[1] = ctx.book.AddNavpoint(title, filename, 10)
			logMsg(LogDefault, "Add index %q (%s) as %s", indexName, title, filename)
			return "", true
		},
	}
}

// tocOutputHandler renders %toc or %toc(Title) as a generated table of
// contents chapter at the position of the command. The optional title makes
// localization easy, e.g. %toc(Inhaltsverzeichnis). The TOC lists every
// chapter and subchapter of the whole book (collected in Pass 1), nested by
// heading level, and works identically for EPUB and AZW3.
func tocOutputHandler() lineHandler {
	return lineHandler{
		match: func(line string, _ bool) bool { return reTocOutput.MatchString(line) },
		handle: func(ctx *parseContext, line string, _ bool) (string, bool) {
			sub := reTocOutput.FindStringSubmatch(line)
			title := sub[1]
			if title == "" {
				title = "Table of Contents"
			}
			if len(tocEntries) == 0 {
				logMsg(LogDefault, "WARNING: %%toc found but the book has no chapters")
				return "", true
			}

			// Flush current chapter before starting the TOC chapter.
			if currentChapterTitle != "" {
				appendPendingFootnotes(ctx)
				addChapter(ctx, currentChapterTitle, currentChapterNumber[1], currentChapterContent)
				currentChapterTitle = ""
				currentChapterContent.Reset()
			}

			currentChapterNumber[1]++
			filename := fmt.Sprintf("xhtml/chapter_%05d.xhtml", currentChapterNumber[1])
			ctx.currentChapterFile = filename

			// tocHref returns the link target for a heading. In AZW3 mode all
			// chapters form one document, so links are plain #label anchors
			// (resolved to exact positions by the KF8 writer). In EPUB mode
			// chapters link to their file, subchapters to file#label.
			tocHref := func(e tocEntry) string {
				if ctx.azw3Mode {
					return "#" + e.label
				}
				if e.level == 1 {
					return "../" + e.chapterFile
				}
				return "../" + e.chapterFile + "#" + e.label
			}

			var body strings.Builder
			body.WriteString(fmt.Sprintf("<section>\n<h1 id=\"label1_%d\">%s</h1>\n", currentChapterNumber[1], title))
			body.WriteString("<ol class=\"toc-list\">\n")
			level := 1
			openLi := false
			for _, e := range tocEntries {
				if e.level > level {
					// Nest deeper inside the currently open list item.
					for level < e.level {
						body.WriteString("\n<ol>\n")
						level++
					}
					openLi = false
				} else {
					if openLi {
						body.WriteString("</li>\n")
					}
					for level > e.level {
						body.WriteString("</ol>\n</li>\n")
						level--
					}
				}
				body.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s</a>", tocHref(e), parseLine(ctx, e.title, true)))
				openLi = true
			}
			if openLi {
				body.WriteString("</li>\n")
			}
			for level > 1 {
				body.WriteString("</ol>\n</li>\n")
				level--
			}
			body.WriteString("</ol>\n</section>\n")

			var htmlContent string
			if ctx.azw3Mode {
				// AZW3: plain fragment only, no document wrapper.
				htmlContent = body.String()
			} else {
				var customCSSLinks string
				for _, p := range ctx.customCSSPaths {
					customCSSLinks += "\n\t\t<link rel=\"stylesheet\" href=\"../" + p + "\"/>"
				}
				htmlContent = `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
    <head>
        <meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>
        <title>` + title + `</title>` + customCSSLinks + `
    </head>
    <body>
	` + body.String() + `
	</body>
</html>`
			}
			if _, err := ctx.book.AddXHTML(filename, title, htmlContent, 10); err != nil {
				logMsg(LogDefault, "ERROR: writing TOC chapter %s: %v", filename, err)
			}
			currentNavpoint[1] = ctx.book.AddNavpoint(title, filename, 10)
			firstparagraph = true
			logMsg(LogDefault, "Add table of contents %q as %s", title, filename)
			return "", true
		},
	}
}

func headlineHandler() lineHandler {
	return lineHandler{
		match: func(line string, insideBlock bool) bool { return reHeadlines.MatchString(line) },
		handle: func(ctx *parseContext, line string, _ bool) (string, bool) {
			matches := reHeadlines.FindStringSubmatch(line)
			chapterLevel := strings.Count(matches[1], "#")
			currentChapterNumber[chapterLevel]++
			currentChapterLabel := fmt.Sprintf("label%d_%d", chapterLevel, currentChapterNumber[chapterLevel])
			firstparagraph = true
			if currentNavpoint[chapterLevel-1] != nil {
				anchorname := fmt.Sprintf("xhtml/chapter_%05d.xhtml#%s", currentChapterNumber[1], currentChapterLabel)
				currentNavpoint[chapterLevel] = currentNavpoint[chapterLevel-1].AddNavpoint(parseLine(ctx, matches[2], true), anchorname, 0)
				logMsg(LogVerbose, "Add subchapter %s as %s", matches[2], anchorname)
			} else {
				logMsg(LogVerbose, "Subchapter %s outside chapter", matches[2])
			}
			return fmt.Sprintf("<h%d id=\"%s\">%s</h%d>\n", chapterLevel, currentChapterLabel, parseLine(ctx, matches[2], true), chapterLevel), true
		},
	}
}

func metaHandler() lineHandler {
	return lineHandler{
		match: func(line string, insideBlock bool) bool { return reMeta.MatchString(line) },
		handle: func(ctx *parseContext, line string, _ bool) (string, bool) {
			matches := reMeta.FindStringSubmatch(line)
			if len(matches) < 2 {
				logMsg(LogDefault, "Error setting meta %s to %s", matches[1], matches[2])
				currentChapterContent.WriteString("<p>" + line + "</p>\n")
				return "", true
			}
			switch matches[1] {
			case "title":
				ctx.book.SetTitle(matches[2])
			case "author":
				ctx.book.AddAuthor(matches[2])
			case "series":
				if err := ctx.book.SetSeries(matches[2]); err != nil {
					logMsg(LogDefault, "ERROR: Add series to %s: %v", matches[2], err)
				}
			case "set":
				if err := ctx.book.SetSet(matches[2]); err != nil {
					logMsg(LogDefault, "ERROR: Add set to %s: %v", matches[2], err)
				}
			case "entry":
				if err := ctx.book.SetEntryNumber(matches[2]); err != nil {
					logMsg(LogDefault, "ERROR: Add entry number to %s: %v", matches[2], err)
				}
			case "uuid":
				if err := ctx.book.SetUUID(matches[2]); err != nil {
					logMsg(LogDefault, "ERROR: Set UUID to %s: %v", matches[2], err)
				}
			case "language":
				if err := ctx.book.AddLanguage(matches[2]); err != nil {
					logMsg(LogDefault, "ERROR: Add language to %s: %v", matches[2], err)
				}
			case "date":
				ctx.book.AddDate(matches[2])
			case "rights":
				ctx.book.AddRights(matches[2])
			case "source":
				ctx.book.AddSource(matches[2])
			case "relation":
				ctx.book.AddRelation(matches[2])
			case "type":
				ctx.book.AddType(matches[2])
			case "quotes":
				quotes := strings.Split(matches[2], ",")
				if len(quotes) != 4 {
					logMsg(LogDefault, "ERROR: quotes definition has to have 4 values seperated by a colon %s %v", matches[2], quotes)
				} else {
					laquo = quotes[0]
					raquo = quotes[1]
					lsaquo = quotes[2]
					rsaquo = quotes[3]
				}
			}
			return "", true
		},
	}
}

func coverHandler() lineHandler {
	return lineHandler{
		match: func(line string, insideBlock bool) bool { return reCover.MatchString(line) },
		handle: func(ctx *parseContext, line string, _ bool) (string, bool) {
			matches := reCover.FindStringSubmatch(line)
			if err := addCover(ctx.book, matches[1], ctx.baseDir, *generateCover); err != nil {
				logMsg(LogDefault, "Error including image %s with URI %s: %v", matches[0], filepath.Join(ctx.baseDir, matches[1]), err)
			}
			return "", true
		},
	}
}

func imageHandler() lineHandler {
	return lineHandler{
		match: func(line string, insideBlock bool) bool { return reImage.MatchString(line) },
		handle: func(ctx *parseContext, line string, _ bool) (string, bool) {
			transformed := reImage.ReplaceAllStringFunc(line, func(match string) string {
				matches := reImage.FindStringSubmatch(match)
				if len(matches) < 2 {
					logMsg(LogDefault, "Error including %s with URI %s", matches[0], matches[2])
					return match
				}
				firstparagraph = true
				currentImageId++
				currentImage := fmt.Sprintf("img/image_%05d%s", currentImageId, filepath.Ext(matches[2]))
				imageID, err := ctx.book.AddImageFile(filepath.Join(ctx.baseDir, matches[2]), currentImage)
				if err != nil {
					logMsg(LogDefault, "Error including image %s with URI %s: %v", matches[0], filepath.Join(ctx.baseDir, matches[2]), err)
					return match
				}
				logMsg(LogVerbose, "Including image %s: %s", imageID, currentImage)
				imgSrc := "../" + currentImage
				if strings.Contains(imageID, "kindle:") {
					imgSrc = imageID
				}
				return fmt.Sprintf(`<img title="%s" alt="%s" src="%s"/>`, matches[4], matches[1], imgSrc)
			})
			return "<div>" + parseLine(ctx, transformed, true) + "</div>\n", true
		},
	}
}

func dividerHandler() lineHandler {
	firstparagraph = true
	return staticResult(reDivider, "<hr/>\n")
}
func pagebreakHandler() lineHandler {
	firstparagraph = true
	return staticResult(rePagebreak, "<MBP:PAGEBREAK/>\n")
}
