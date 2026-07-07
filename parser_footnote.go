package main

import (
	"fmt"
	"strings"
)

// footnoteDefHandler consumes footnote definition lines ([^id]: text). The
// definition text was already collected in Pass 1 (scanAnchorsAndIndex); in
// Pass 2 the line produces no inline output — the note is emitted as an
// <aside> when the chapter is finalized.
func footnoteDefHandler() lineHandler {
	return lineHandler{
		match: func(line string, insideBlock bool) bool {
			return !insideBlock && reFootnoteDef.MatchString(line)
		},
		handle: func(_ *parseContext, _ string, _ bool) (string, bool) {
			return "", true
		},
	}
}

// footnoteRefHandler rewrites an inline footnote reference [^id] into an
// EPUB3/KF8 note reference. The referenced note is numbered per chapter and
// queued for emission at the end of the chapter. The href="#fn-…" target is
// resolved to a same-file fragment (EPUB) or a kindle:pos:fid link (AZW3, via
// the mobi layer), which drives popup footnotes on supporting readers.
func footnoteRefHandler() lineHandler {
	return lineHandler{
		match: func(line string, _ bool) bool { return matchOutsideBackticks(line, reFootnoteRef) },
		handle: func(ctx *parseContext, line string, insideBlock bool) (string, bool) {
			out := replaceOutsideBackticks(line, reFootnoteRef, func(sub []string) string {
				id := sub[1]
				if _, ok := footnoteDefs[id]; !ok {
					logMsg(LogDefault, "WARNING: footnote %q referenced but not defined", id)
					return sub[0]
				}
				num, seen := footnoteAssigned[id]
				c := currentChapterNumber[1]
				if seen {
					// A repeat reference must not duplicate the fnref id; the
					// back-link points at the first reference only.
					return fmt.Sprintf(`<a epub:type="noteref" href="#fn-%d-%d"><sup>%d</sup></a>`,
						c, num, num)
				}
				footnoteNum++
				num = footnoteNum
				footnoteAssigned[id] = num
				pendingFootnotes = append(pendingFootnotes, pendingNote{
					chap: c,
					num:  num,
					id:   id,
				})
				return fmt.Sprintf(`<a epub:type="noteref" href="#fn-%d-%d" id="fnref-%d-%d"><sup>%d</sup></a>`,
					c, num, c, num, num)
			})
			return parseLine(ctx, out, insideBlock), true
		},
	}
}

// appendPendingFootnotes writes the footnotes referenced in the current
// chapter as <aside epub:type="footnote"> elements and resets the per-chapter
// footnote state. It must be called just before a chapter's content is handed
// to addChapter, so the notes land at the end of that chapter.
func appendPendingFootnotes(ctx *parseContext) {
	defer func() {
		pendingFootnotes = nil
		footnoteNum = 0
		footnoteAssigned = map[string]int{}
	}()
	if len(pendingFootnotes) == 0 {
		return
	}

	var b strings.Builder
	b.WriteString("<section epub:type=\"footnotes\" class=\"footnotes\">\n")
	for _, n := range pendingFootnotes {
		text := parseLine(ctx, footnoteDefs[n.id], true)
		b.WriteString(fmt.Sprintf(
			"<aside epub:type=\"footnote\" id=\"fn-%d-%d\"><p><sup>%d</sup> %s <a href=\"#fnref-%d-%d\">&#8617;</a></p></aside>\n",
			n.chap, n.num, n.num, text, n.chap, n.num))
	}
	b.WriteString("</section>\n")
	currentChapterContent.WriteString(b.String())
}
