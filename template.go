package main

import (
	"log"

	"github.com/writingtoole/epub"
)

const defaultCSS = `// Default CSS
h1, h2, h3, h4, h5, h6 {
	font-family: sans-serif;
	font-variant: small-caps;
	color: #888888;
}
h1 {
	border: 1px solid #888;
	border-width: 1px 0;
	width: 80%;
	padding: 0.5em;
	margin: 0 auto 0 auto;
}
`

func addDefaultTemplate(book *epub.EPub) {
	book.AddStylesheet("css/styles.css", defaultCSS)
	log.Print("Added default stylesheet css/styles.css")
}
