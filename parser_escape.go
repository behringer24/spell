package main

import (
	"fmt"
	"strings"
)

// asciiPunct is the set of ASCII punctuation characters that a backslash may
// escape, following the CommonMark rule: a backslash before any ASCII
// punctuation is an escape; before anything else it is a literal backslash.
const asciiPunct = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

func isASCIIPunct(r rune) bool {
	return r < 128 && strings.ContainsRune(asciiPunct, r)
}

// resolveEscapes rewrites backslash escapes in a single left-to-right pass.
//
// A backslash before an ASCII punctuation character is replaced by that
// character's numeric HTML entity (e.g. \* -> &#42;). The entity is inert to
// every downstream regex (which match the literal character) and already
// renders as the intended character, so no later un-masking is needed — the
// same trick codeSpanHandler uses for spell's own trigger characters.
//
// Backticks are tracked so that escapes inside inline code are left literal
// (Markdown does not process escapes inside code spans). Because an escaped
// backtick is consumed as an escape before it can toggle code state, \`
// correctly yields a literal backtick instead of opening a code span.
func resolveEscapes(line string) string {
	runes := []rune(line)
	var b strings.Builder
	b.Grow(len(line))
	inCode := false
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '\\' && !inCode && i+1 < len(runes) {
			next := runes[i+1]
			if isASCIIPunct(next) {
				fmt.Fprintf(&b, "&#%d;", next)
				i++
				continue
			}
			// Backslash before a non-punctuation rune stays literal.
		}
		if r == '`' {
			inCode = !inCode
		}
		b.WriteRune(r)
	}
	return b.String()
}

// escapeHandler resolves backslash escapes before any other handler runs, so
// that escaped block markers (\#, \-, \1.) are not mistaken for structure and
// escaped inline markers (\*, \%, \[) are not interpreted. It is skipped
// inside verbatim code fences, where backslashes are literal.
//
// match reports true only when resolution would actually change the line;
// this both avoids needless work and guarantees termination, since after one
// pass no resolvable escape remains and the handler no longer matches.
func escapeHandler() lineHandler {
	return lineHandler{
		match: func(line string, _ bool) bool {
			if inBlockType == BLOCKTYPE_CODE {
				return false
			}
			if strings.IndexByte(line, '\\') < 0 {
				return false
			}
			return resolveEscapes(line) != line
		},
		handle: func(ctx *parseContext, line string, insideBlock bool) (string, bool) {
			return parseLine(ctx, resolveEscapes(line), insideBlock), true
		},
	}
}
