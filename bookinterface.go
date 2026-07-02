package main

// NavpointAdder is the common interface for epub and azw3 table-of-contents nodes.
type NavpointAdder interface {
	AddNavpoint(label, filename string, order int) NavpointAdder
}

// SpellBook is the common interface implemented by both epubBook and azw3Book wrappers.
type SpellBook interface {
	SetTitle(title string)
	AddAuthor(author string)
	AddLanguage(lang string) error
	SetSeries(series string) error
	SetSet(s string) error
	SetEntryNumber(entry string) error
	SetUUID(uuid string) error
	AddDate(date string)
	AddRights(rights string)
	AddSource(source string)
	AddRelation(rel string)
	AddType(t string)
	// AddXHTML adds an XHTML chapter. title is used by azw3; epub ignores it.
	AddXHTML(filename, title, content string, order int) (string, error)
	AddImageFile(source, dest string) (string, error)
	AddImage(path string, contents []byte) (string, error)
	SetCoverImage(id string)
	AddStylesheet(path, content string)
	AddNavpoint(label, filename string, order int) NavpointAdder
	Write(filename string) error
}
