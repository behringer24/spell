package main

import (
	"fmt"
	"regexp"
	"strings"
)

// anchorEntry maps a user-defined anchor id to the chapter file it lives in.
type anchorEntry struct {
	chapterFile string
}

// indexEntry records one occurrence of an index term in the text.
type indexEntry struct {
	displayTerm string // text shown in the running text
	canonical   string // key used for grouping in the index (defaults to displayTerm)
	indexName   string
	chapterFile string
	htmlID      string
}

var (
	// anchors maps anchor id → chapter file (populated by scanAnchorsAndIndex)
	anchors = map[string]anchorEntry{}

	// indexes maps index name → ordered list of occurrences
	indexes = map[string][]indexEntry{}

	// indexCounters tracks how many times each (indexName+term) combo has been seen
	// during Pass 2 rendering so we can assign stable sequential IDs
	indexCounters = map[string]int{}
)

var (
	reAnchorDef   = regexp.MustCompile(`\{#([a-zA-Z0-9_-]+)\}`)
	reAnchorLink  = regexp.MustCompile(`\[([^\]]+)\]\(#([a-zA-Z0-9_-]+)\)`)
	// %[displayTerm](indexname) or %[displayTerm](indexname|canonical)
	reIndexEntry = regexp.MustCompile(`%\[([^\]]+)\]\(([^)|]+)(?:\|([^)]+))?\)`)
	// %index[name] or %index[name](Title)
	reIndexOutput = regexp.MustCompile(`^%index\[([^\]]+)\](?:\(([^)]+)\))?$`)
)

// resetAnchors clears all anchor/index state so processMarkdownFile is idempotent.
func resetAnchors() {
	anchors = map[string]anchorEntry{}
	indexes = map[string][]indexEntry{}
	indexCounters = map[string]int{}
}

// chapterFileForNumber returns the deterministic XHTML filename for a chapter number.
func chapterFileForNumber(n int) string {
	return fmt.Sprintf("xhtml/chapter_%05d.xhtml", n)
}

// scanAnchorsAndIndex performs Pass 1: walks all lines of the fully-included
// content, tracks the current chapter number exactly as chapterHandler does,
// and populates the anchors and indexes maps.
func scanAnchorsAndIndex(content string) {
	lines := reNewline.Split(content, -1)
	chapterNum := 0

	for _, line := range lines {
		// Mirror chapterHandler: a top-level heading advances the chapter counter.
		if reChapter.MatchString(line) {
			chapterNum++
		}

		currentFile := chapterFileForNumber(chapterNum)

		// Collect {#id} anchor definitions.
		for _, m := range reAnchorDef.FindAllStringSubmatch(line, -1) {
			id := m[1]
			if _, exists := anchors[id]; exists {
				logMsg(LogDefault, "WARNING: duplicate anchor id %q (second occurrence in %s ignored)", id, currentFile)
				continue
			}
			anchors[id] = anchorEntry{chapterFile: currentFile}
			logMsg(LogVerbose, "Anchor %q registered in %s", id, currentFile)
		}

		// Collect %[displayTerm](indexname) or %[displayTerm](indexname|canonical) entries.
		for _, m := range reIndexEntry.FindAllStringSubmatch(line, -1) {
			displayTerm, indexName, canonical := m[1], m[2], m[3]
			if canonical == "" {
				canonical = displayTerm
			}
			key := indexName + "\x00" + canonical
			seq := indexCounters[key]
			indexCounters[key]++
			htmlID := fmt.Sprintf("idx-%s-%s-%d", sanitizeID(indexName), sanitizeID(canonical), seq)
			indexes[indexName] = append(indexes[indexName], indexEntry{
				displayTerm: displayTerm,
				canonical:   canonical,
				indexName:   indexName,
				chapterFile: currentFile,
				htmlID:      htmlID,
			})
			logMsg(LogVerbose, "Index entry %q → %q (%s) registered as %s in %s", displayTerm, canonical, indexName, htmlID, currentFile)
		}
	}

	// Reset counters; Pass 2 rendering uses the same deterministic formula.
	indexCounters = map[string]int{}
}

// sanitizeID converts a string to a safe HTML id fragment.
func sanitizeID(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

// resolveAnchorHref returns the href value for a link to anchorID from currentChapterFile.
// In AZW3 mode all chapters form one concatenated document, so all links are plain #id.
func resolveAnchorHref(ctx *parseContext, anchorID, currentChapterFile string) string {
	entry, ok := anchors[anchorID]
	if !ok {
		logMsg(LogDefault, "WARNING: anchor %q not found", anchorID)
		return "#" + anchorID
	}
	if ctx.azw3Mode || entry.chapterFile == currentChapterFile {
		return "#" + anchorID
	}
	return "../" + entry.chapterFile + "#" + anchorID
}
