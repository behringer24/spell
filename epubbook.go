package main

import "github.com/behringer24/epub"

// epubNavpoint wraps *epub.Navpoint to implement NavpointAdder.
type epubNavpoint struct{ np *epub.Navpoint }

func (n *epubNavpoint) AddNavpoint(label, filename string, order int) NavpointAdder {
	return &epubNavpoint{np: n.np.AddNavpoint(label, filename, order)}
}

// epubBook wraps *epub.EPub to implement SpellBook.
type epubBook struct{ book *epub.EPub }

func newEpubBook(version float64) *epubBook {
	b := epub.New()
	b.SetVersion(version)
	return &epubBook{book: b}
}

func (b *epubBook) SetTitle(title string)        { b.book.SetTitle(title) }
func (b *epubBook) AddAuthor(author string)       { b.book.AddAuthor(author) }
func (b *epubBook) AddLanguage(lang string) error { return b.book.AddLanguage(lang) }
func (b *epubBook) SetSeries(s string) error      { return b.book.SetSeries(s) }
func (b *epubBook) SetSet(s string) error         { return b.book.SetSet(s) }
func (b *epubBook) SetEntryNumber(n string) error { return b.book.SetEntryNumber(n) }
func (b *epubBook) SetUUID(uu string) error       { return b.book.SetUUID(uu) }
func (b *epubBook) AddDate(date string)           { b.book.AddDate(date) }
func (b *epubBook) AddRights(rights string)       { b.book.AddRights(rights) }
func (b *epubBook) AddSource(source string)       { b.book.AddSource(source) }
func (b *epubBook) AddRelation(rel string)        { b.book.AddRelation(rel) }
func (b *epubBook) AddType(t string)              { b.book.AddType(t) }

func (b *epubBook) AddXHTML(filename, _ /* title */, content string, order int) (string, error) {
	id, err := b.book.AddXHTML(filename, content, order)
	return string(id), err
}

func (b *epubBook) AddImageFile(source, dest string) (string, error) {
	id, err := b.book.AddImageFile(source, dest)
	return string(id), err
}

func (b *epubBook) AddImage(path string, contents []byte) (string, error) {
	id, err := b.book.AddImage(path, contents)
	return string(id), err
}

func (b *epubBook) SetCoverImage(id string) { b.book.SetCoverImage(epub.Id(id)) }

func (b *epubBook) AddStylesheet(path, content string) { b.book.AddStylesheet(path, content) }

func (b *epubBook) AddNavpoint(label, filename string, order int) NavpointAdder {
	return &epubNavpoint{np: b.book.AddNavpoint(label, filename, order)}
}

func (b *epubBook) SetStartReading(filename string) { b.book.SetStartReading(filename) }

func (b *epubBook) Write(filename string) error { return b.book.Write(filename) }
