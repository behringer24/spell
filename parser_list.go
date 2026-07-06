package main

import (
	"fmt"
	"strconv"
	"strings"
)

// listFrame is one open list level. kind is "ul" or "ol"; indent is the
// display width of the leading whitespace that opened this level. The most
// recent <li> at each level is left open so that a deeper list nests inside
// it (as required for valid HTML), and is closed when the level is left.
type listFrame struct {
	indent int
	kind   string
}

// indentWidth returns the display width of leading whitespace, expanding
// tabs to the next multiple of 4 so that tab- and space-indented sublists
// map onto the same depth scale.
func indentWidth(s string) int {
	w := 0
	for _, r := range s {
		if r == '\t' {
			w += 4 - (w % 4)
		} else {
			w++
		}
	}
	return w
}

// openListTag returns the opening tag for a list, adding a start attribute
// for ordered lists that do not begin at 1 (matching CommonMark).
func openListTag(kind string, start int) string {
	if kind == "ol" && start > 1 {
		return fmt.Sprintf("<ol start=\"%d\">\n", start)
	}
	return "<" + kind + ">\n"
}

// closeAllLists pops every open list level, closing each open <li> and its
// list tag, and empties the stack. Used when a non-list block or the end of
// the input ends the current list.
func closeAllLists() string {
	var b strings.Builder
	for len(listStack) > 0 {
		top := listStack[len(listStack)-1]
		b.WriteString("</li>\n</" + top.kind + ">\n")
		listStack = listStack[:len(listStack)-1]
	}
	return b.String()
}

// listCloseHandler closes the open list(s) when a line that is not a list
// item appears. The closing tags are written to the current chapter
// immediately (not returned) so that a following chapter/index/toc command,
// which flushes the chapter mid-recursion, sees a properly closed list.
func listCloseHandler() lineHandler {
	return lineHandler{
		match: func(line string, insideBlock bool) bool {
			return len(listStack) > 0 && !insideBlock && !reListItem.MatchString(line)
		},
		handle: func(ctx *parseContext, line string, _ bool) (string, bool) {
			currentChapterContent.WriteString(closeAllLists())
			return parseLine(ctx, line, false), true
		},
	}
}

// listItemHandler renders bullet (-, *, +) and ordered (1. / 1)) list items,
// nesting by indentation. A deeper-indented item opens a sublist inside the
// current item; a shallower one closes levels back down. Mixed markers at
// the same indent start a new list of the other kind.
func listItemHandler() lineHandler {
	return lineHandler{
		match: func(line string, insideBlock bool) bool {
			return !insideBlock && reListItem.MatchString(line)
		},
		handle: func(ctx *parseContext, line string, _ bool) (string, bool) {
			m := reListItem.FindStringSubmatch(line)
			indent := indentWidth(m[1])
			kind := "ul"
			start := 0
			if m[3] != "" {
				kind = "ol"
				start, _ = strconv.Atoi(m[3])
			}
			content := parseLine(ctx, strings.TrimSpace(m[4]), true)

			var b strings.Builder

			// Leave any level more indented than this item.
			for len(listStack) > 0 && indent < listStack[len(listStack)-1].indent {
				top := listStack[len(listStack)-1]
				b.WriteString("</li>\n</" + top.kind + ">\n")
				listStack = listStack[:len(listStack)-1]
			}

			switch {
			case len(listStack) == 0 || indent > listStack[len(listStack)-1].indent:
				// Start a new list; when nesting, it opens inside the still
				// open <li> of the parent level.
				b.WriteString(openListTag(kind, start))
				listStack = append(listStack, listFrame{indent: indent, kind: kind})
			case listStack[len(listStack)-1].kind != kind:
				// Same indent, different marker type: close the old list and
				// begin a new one of the other kind.
				top := listStack[len(listStack)-1]
				b.WriteString("</li>\n</" + top.kind + ">\n")
				listStack = listStack[:len(listStack)-1]
				b.WriteString(openListTag(kind, start))
				listStack = append(listStack, listFrame{indent: indent, kind: kind})
			default:
				// Same list, next sibling: close the previous item.
				b.WriteString("</li>\n")
			}

			b.WriteString("  <li>" + content)
			return b.String(), true
		},
	}
}
