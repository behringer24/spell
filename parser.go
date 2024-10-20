package main

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/writingtoole/epub"
)

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

func addCover(book *epub.EPub, imageFile string, baseDir string) error {
	currentImage := fmt.Sprintf("img/cover%s", filepath.Ext(imageFile))
	imageID, err := book.AddImageFile(filepath.Join(baseDir, imageFile), currentImage)
	if err != nil {
		return err
	} else {
		book.SetCoverImage(imageID)
		log.Printf("Added cover image %s: %s", imageID, currentImage)
	}

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
	return nil
}

// Parse chapters and other Markdown commands
func parseMarkdown(book *epub.EPub, content string, baseDir string) error {
	var currentChapterContent strings.Builder
	var currentChapterTitle string
	var currentChapterNumber [7]int
	var currentNavpoint [7]*epub.Navpoint
	var currentImageId int

	// split contents by lines
	splitRegex := regexp.MustCompile(`\r?\n`)
	lines := splitRegex.Split(content, -1)

	// define regular expressions do detect commands
	chapterRegex := regexp.MustCompile(`^\s*(#)\s*([^#]+)$`)
	headlinesRegex := regexp.MustCompile(`^\s*(#{2,6})\s*([^#]+)$`)
	dividerRegex := regexp.MustCompile(`^\s*([\*,\-]\s*)+$`)
	pagebreakRegex := regexp.MustCompile(`^\s*(_\s*)+$`)
	metaRegex := regexp.MustCompile(`\!\[(title|author|series|set|entry|uuid|language)\]\(\"([^\"]+)\"\)`)
	coverRegex := regexp.MustCompile(`\!\[cover\]\(([^ \)]+)\s*(\"([^\"]*)\")?\)`)
	imageRegex := regexp.MustCompile(`\!\[([^\]]*)\]\(([^ \)]+)\s*(\"([^\"]*)\")?\)`)

	addDefaultTemplate(book)

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
			filename := fmt.Sprintf("xhtml/chapter_%05d.xhtml", currentChapterNumber[1])
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
				anchorname := fmt.Sprintf("xhtml/chapter_%05d.xhtml#%s", currentChapterNumber[1], currentChapterLabel)
				currentNavpoint[chapterLevel] = currentNavpoint[chapterLevel-1].AddNavpoint(matches[2], anchorname, 0)
				log.Printf("Add subchapter %s as %s", matches[2], anchorname)
			} else {
				log.Printf("Subchapter %s outside chapter", matches[2])
			}
		} else if metaRegex.MatchString(line) {
			// Set meta variables
			matches := metaRegex.FindStringSubmatch(line)
			if len(matches) < 2 {
				log.Printf("Error setting meta %s to %s", matches[1], matches[2])
				currentChapterContent.WriteString("<p>" + line + "</p>\n")
			} else {
				if strings.Compare(matches[1], "title") == 0 {
					book.SetTitle(matches[2])
				} else if strings.Compare(matches[1], "author") == 0 {
					book.AddAuthor(matches[2])
				} else if strings.Compare(matches[1], "series") == 0 {
					err := book.SetSeries(matches[2])
					if err != nil {
						log.Printf("ERROR: Add series to %s: %v", matches[2], err)
					}
				} else if strings.Compare(matches[1], "set") == 0 {
					err := book.SetSet(matches[2])
					if err != nil {
						log.Printf("ERROR: Add set to %s: %v", matches[2], err)
					}
				} else if strings.Compare(matches[1], "entry") == 0 {
					err := book.SetEntryNumber(matches[2])
					if err != nil {
						log.Printf("ERROR: Add entry number to %s: %v", matches[2], err)
					}
				} else if strings.Compare(matches[1], "uuid") == 0 {
					err := book.SetUUID(matches[2])
					if err != nil {
						log.Printf("ERROR: Set UUID to %s: %v", matches[2], err)
					}
				} else if strings.Compare(matches[1], "language") == 0 {
					err := book.AddLanguage(matches[2])
					if err != nil {
						log.Printf("ERROR: Add language to %s: %v", matches[2], err)
					}
				}
			}
		} else if coverRegex.MatchString(line) {
			// Add cover image and page
			matches := coverRegex.FindStringSubmatch(line)

			err := addCover(book, matches[1], baseDir)

			if err != nil {
				log.Printf("Error including image %s with URI %s: %v", matches[0], filepath.Join(baseDir, matches[1]), err)
			}
		} else if imageRegex.MatchString(line) {
			// Add image
			line = imageRegex.ReplaceAllStringFunc(line, func(match string) string {
				// Extract includes and parameters
				matches := imageRegex.FindStringSubmatch(match)
				if len(matches) < 2 {
					log.Printf("Error including %s with URI %s", matches[0], matches[2])
					return match // Fallback: if the pattern is wrong
				}

				currentImageId++
				currentImage := fmt.Sprintf("img/image_%05d%s", currentImageId, filepath.Ext(matches[2]))
				imageID, err := book.AddImageFile(filepath.Join(baseDir, matches[2]), currentImage)
				if err != nil {
					log.Printf("Error including image %s with URI %s: %v", matches[0], filepath.Join(baseDir, matches[2]), err)
					return match
				}
				log.Printf("Including image %s: %s", imageID, currentImage)
				return fmt.Sprintf(`<img title="%s" alt="%s" src="../%s"/>`, matches[4], matches[1], currentImage)
			})

			currentChapterContent.WriteString("<div>" + line + "</div>\n")
		} else if dividerRegex.MatchString(line) {
			// Add horizontal break
			currentChapterContent.WriteString("<hr/>\n")
		} else if pagebreakRegex.MatchString(line) {
			// Add page break (not working on many ebook readers)
			currentChapterContent.WriteString(`<div style="page-break-after: always"></div>` + "\n")
		} else {
			// Normal line just add if not empty
			if strings.Compare(strings.TrimSpace(line), "") != 0 {
				currentChapterContent.WriteString("<p>" + line + "</p>\n")
			}
		}
	}

	// Add last chapter
	if currentChapterTitle != "" {
		addChapter(book, currentChapterTitle, currentChapterNumber[1], currentChapterContent)
	}

	return nil
}
