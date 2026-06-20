package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

func blockquoteFenceHandler() lineHandler {
	return lineHandler{
		match: func(line string, insideBlock bool) bool { return reBlockQuote.MatchString(line) },
		handle: func(_ *parseContext, line string, _ bool) (string, bool) {
			if inBlockType > 0 {
				log.Printf("blockQuote schließen")
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
			log.Printf("blockQuote opening: %s", blocktype)
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
				log.Printf("blockQuote CODE line")
				return fmt.Sprintf("%s</br>\n", line), true
			}
			log.Printf("blockQuote non CODE but parsed line")
			return fmt.Sprintf("%s</br>\n", parseLine(ctx, line, true)), true
		},
	}
}

func chapterHandler() lineHandler {
	return lineHandler{
		match: func(line string, insideBlock bool) bool { return reChapter.MatchString(line) },
		handle: func(ctx *parseContext, line string, _ bool) (string, bool) {
			if currentChapterTitle != "" {
				addChapter(ctx, currentChapterTitle, currentChapterNumber[1], currentChapterContent)
			}
			matches := reChapter.FindStringSubmatch(line)
			currentChapterTitle = parseLine(ctx, matches[2], true)
			currentChapterContent.Reset()
			currentChapterNumber[1]++
			filename := fmt.Sprintf("xhtml/chapter_%05d.xhtml", currentChapterNumber[1])
			currentNavpoint[1] = ctx.book.AddNavpoint(currentChapterTitle, filename, 10)
			firstparagraph = true
			return "<h1>" + parseLine(ctx, matches[2], true) + "</h1>\n", true
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
				log.Printf("Add subchapter %s as %s", matches[2], anchorname)
			} else {
				log.Printf("Subchapter %s outside chapter", matches[2])
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
				log.Printf("Error setting meta %s to %s", matches[1], matches[2])
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
					log.Printf("ERROR: Add series to %s: %v", matches[2], err)
				}
			case "set":
				if err := ctx.book.SetSet(matches[2]); err != nil {
					log.Printf("ERROR: Add set to %s: %v", matches[2], err)
				}
			case "entry":
				if err := ctx.book.SetEntryNumber(matches[2]); err != nil {
					log.Printf("ERROR: Add entry number to %s: %v", matches[2], err)
				}
			case "uuid":
				if err := ctx.book.SetUUID(matches[2]); err != nil {
					log.Printf("ERROR: Set UUID to %s: %v", matches[2], err)
				}
			case "language":
				if err := ctx.book.AddLanguage(matches[2]); err != nil {
					log.Printf("ERROR: Add language to %s: %v", matches[2], err)
				}
			case "quotes":
				quotes := strings.Split(matches[2], ",")
				if len(quotes) != 4 {
					log.Printf("ERROR: quotes definition has to have 4 values seperated by a colon %s %v", matches[2], quotes)
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
				log.Printf("Error including image %s with URI %s: %v", matches[0], filepath.Join(ctx.baseDir, matches[1]), err)
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
					log.Printf("Error including %s with URI %s", matches[0], matches[2])
					return match
				}
				currentImageId++
				currentImage := fmt.Sprintf("img/image_%05d%s", currentImageId, filepath.Ext(matches[2]))
				imageID, err := ctx.book.AddImageFile(filepath.Join(ctx.baseDir, matches[2]), currentImage)
				if err != nil {
					log.Printf("Error including image %s with URI %s: %v", matches[0], filepath.Join(ctx.baseDir, matches[2]), err)
					return match
				}
				log.Printf("Including image %s: %s", imageID, currentImage)
				return fmt.Sprintf(`<img title="%s" alt="%s" src="../%s"/>`, matches[4], matches[1], currentImage)
			})
			return "<div>" + parseLine(ctx, transformed, true) + "</div>\n", true
		},
	}
}

func dividerHandler() lineHandler   { return staticResult(reDivider, "<hr/>\n") }
func pagebreakHandler() lineHandler { return staticResult(rePagebreak, "<MBP:PAGEBREAK/>\n") }
