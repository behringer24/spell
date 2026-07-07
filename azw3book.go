package main

import (
	"encoding/binary"
	"encoding/hex"
	"strings"
	"time"

	"github.com/behringer24/azw3"
)

// azw3Navpoint wraps *azw3.Navpoint to implement NavpointAdder.
// Fragment identifiers (#anchor) are stripped from targets because KF8
// navigation links to whole chapters, not sub-chapter positions.
type azw3Navpoint struct{ np *azw3.Navpoint }

func (n *azw3Navpoint) AddNavpoint(label, target string, order int) NavpointAdder {
	return &azw3Navpoint{np: n.np.AddNavpoint(label, stripFragment(target), order)}
}

// azw3Book wraps *azw3.Book to implement SpellBook.
type azw3Book struct{ book *azw3.Book }

func newAZW3Book() *azw3Book { return &azw3Book{book: azw3.New()} }

func (b *azw3Book) SetTitle(title string)        { b.book.SetTitle(title) }
func (b *azw3Book) AddAuthor(author string)       { b.book.AddAuthor(author) }
func (b *azw3Book) AddLanguage(lang string) error { return b.book.AddLanguage(lang) }

// Series/collection metadata is intentionally dropped for AZW3. Unlike EPUB3
// (belongs-to-collection), KF8/MOBI has no native series record: verified
// against calibre (writes no series EXTH), KindleUnpack (no series entry in
// the EXTH 500-544 table) and libmobi (no series tag). Amazon groups series
// only via its catalog by ASIN, for purchased titles. The only device-visible
// alternative is folding the series into the title, which was deliberately
// declined so the displayed title stays clean. Do not "fix" these into title
// mangling without revisiting that decision.
func (b *azw3Book) SetSeries(_ string) error      { return nil }
func (b *azw3Book) SetSet(_ string) error         { return nil }
func (b *azw3Book) SetEntryNumber(_ string) error { return nil }
func (b *azw3Book) SetUUID(uu string) error {
	// Derive a uint32 from the first 4 bytes of the UUID hex digits.
	clean := strings.ReplaceAll(uu, "-", "")
	if len(clean) >= 8 {
		if raw, err := hex.DecodeString(clean[:8]); err == nil {
			b.book.SetUniqueID(binary.BigEndian.Uint32(raw))
		}
	}
	return nil
}
func (b *azw3Book) AddRights(_ string)            {}
func (b *azw3Book) AddSource(_ string)            {}
func (b *azw3Book) AddRelation(_ string)          {}
func (b *azw3Book) AddType(_ string)              {}

func (b *azw3Book) AddDate(date string) {
	for _, layout := range []string{"2006-01-02", "2006", "January 2, 2006", "02 January 2006"} {
		if t, err := time.Parse(layout, date); err == nil {
			b.book.SetPublishedDate(t)
			return
		}
	}
}

func (b *azw3Book) AddXHTML(filename, title, content string, order int) (string, error) {
	id, err := b.book.AddChapter(filename, title, content, order)
	return string(id), err
}

func (b *azw3Book) AddImageFile(source, dest string) (string, error) {
	id, err := b.book.AddImageFile(source, dest)
	return string(id), err
}

func (b *azw3Book) AddImage(path string, contents []byte) (string, error) {
	id, err := b.book.AddImage(path, contents)
	return string(id), err
}

func (b *azw3Book) SetCoverImage(id string) { b.book.SetCoverImage(azw3.Id(id)) }

func (b *azw3Book) AddStylesheet(path, content string) { b.book.AddStylesheet(path, content) }

func (b *azw3Book) AddNavpoint(label, target string, order int) NavpointAdder {
	return &azw3Navpoint{np: b.book.AddNavpoint(label, stripFragment(target), order)}
}

func (b *azw3Book) SetStartReading(filename string) { b.book.SetStartReading(azw3.Id(filename)) }

func (b *azw3Book) Write(filename string) error { return b.book.Write(filename) }

// stripFragment removes any "#anchor" suffix from a navpoint target path.
func stripFragment(target string) string {
	if i := strings.Index(target, "#"); i >= 0 {
		return target[:i]
	}
	return target
}
