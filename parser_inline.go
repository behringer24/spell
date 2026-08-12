package main

import (
	"fmt"
	"strings"
)

func quotesHandler() lineHandler {
	return replaceEachAndRecurse(reQuotes, func(_ *parseContext, marker string) string {
		switch marker {
		case `%"`:
			return laquo
		case `"%`:
			return raquo
		case `%'`:
			return lsaquo
		case `'%`:
			return rsaquo
		default:
			return marker
		}
	})
}

// linkHandler renders inline links [text](url) and [text](url "title") as
// external hyperlinks. Internal fragment links [text](#id) are handled
// earlier by anchorLinkHandler; images ![alt](url) by imageHandler. Matches
// inside backtick code spans are left untouched.
func linkHandler() lineHandler {
	return lineHandler{
		match: func(line string, _ bool) bool { return matchOutsideBackticks(line, reLink) },
		handle: func(ctx *parseContext, line string, insideBlock bool) (string, bool) {
			out := replaceOutsideBackticks(line, reLink, func(sub []string) string {
				if sub[3] != "" {
					return fmt.Sprintf(`<a href="%s" title="%s">%s</a>`, sub[2], sub[3], sub[1])
				}
				return fmt.Sprintf(`<a href="%s">%s</a>`, sub[2], sub[1])
			})
			return parseLine(ctx, out, insideBlock), true
		},
	}
}

func boldHandler() lineHandler {
	return replaceEachAndRecurse(reBold, func(ctx *parseContext, inner string) string {
		return "<b>" + parseLine(ctx, inner, true) + "</b>"
	})
}

func italicHandler() lineHandler {
	return replaceEachAndRecurse(reItalic, func(ctx *parseContext, inner string) string {
		return "<i>" + parseLine(ctx, inner, true) + "</i>"
	})
}

func codeSpanHandler() lineHandler {
	return replaceEachAndRecurse(reCode, func(_ *parseContext, inner string) string {
		// Encode spell-specific trigger characters so that downstream handlers
		// cannot match content that was written as inline code. The backslash
		// is encoded first (it is a backslash-escape trigger, and encoding it
		// keeps escapes literal inside code); its replacement introduces no
		// further trigger characters.
		inner = strings.ReplaceAll(inner, "\\", "&#92;")
		inner = strings.ReplaceAll(inner, "%", "&#37;")
		inner = strings.ReplaceAll(inner, "{", "&#123;")
		inner = strings.ReplaceAll(inner, "}", "&#125;")
		inner = strings.ReplaceAll(inner, "[", "&#91;")
		inner = strings.ReplaceAll(inner, "]", "&#93;")
		return `<span class="code">` + inner + "</span>"
	})
}

// Numeric character references, not named HTML entities: an EPUB3 XHTML document
// is parsed as XML, where only &lt; &gt; &amp; &quot; &apos; are predefined, so
// &nbsp; and friends are undeclared-entity errors under epubcheck.
func commentHandler() lineHandler  { return replaceAndRecurse(reComment, "$1") }
func emDashHandler() lineHandler   { return replaceAndRecurse(reLongDash, "&#160;&#8212;&#160;") }
func enDashHandler() lineHandler   { return replaceAndRecurse(reMidDash, "&#160;&#8211;&#160;") }
func ellipsisHandler() lineHandler { return replaceAndRecurse(reThreeDots, "&#8230;") }
