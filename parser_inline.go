package main

import "strings"

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
		// Encode spell-specific trigger characters so that anchor/index handlers
		// cannot match content that was written as inline code.
		inner = strings.ReplaceAll(inner, "%", "&#37;")
		inner = strings.ReplaceAll(inner, "{", "&#123;")
		inner = strings.ReplaceAll(inner, "}", "&#125;")
		inner = strings.ReplaceAll(inner, "[", "&#91;")
		inner = strings.ReplaceAll(inner, "]", "&#93;")
		return `<span class="code">` + inner + "</span>"
	})
}

func commentHandler() lineHandler  { return replaceAndRecurse(reComment, "$1") }
func emDashHandler() lineHandler   { return replaceAndRecurse(reLongDash, "&nbsp;&mdash;&nbsp;") }
func enDashHandler() lineHandler   { return replaceAndRecurse(reMidDash, "&nbsp;&ndash;&nbsp;") }
func ellipsisHandler() lineHandler { return replaceAndRecurse(reThreeDots, "&hellip;") }
