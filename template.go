package main

import (
	"log"

	"github.com/writingtoole/epub"
)

const defaultCSS = `// Default CSS
body {
  font-family: Palatino;
}
h1, h2, h3, h4, h5, h6 {
	font-family: Helvetica;
	font-variant: small-caps;
	color: #666666;
}
h1 {
	border: 1px solid #888;
	border-width: 1px 0;
	text-align: center;
	width: 75%;
	padding: 0.5em;
	margin: 2em auto 4em auto;
}
h2 {
	border: 1px solid #888;
	border-width: 0 0 1px 0;
	padding: 0.5em 0;
}
p.firstparagraph {
	text-indent: 2em;
}
p {
	text-indent: 1em;
}
p, li {
	margin-top: 0.3em;
    margin-bottom: 0.3em;
}
hr {
	border: 1px solid #444;
	height: 2px;
	width: 25%;
	margin: 3em auto 3em auto;
}
blockquote {
	margin-left: 3em;
	padding: 1em;
	background-color: #aaa;
	border-radius: 5px;
}
blockquote.code {
	font-family: "monospace"
}
span.code {
	background-color: #aaa;
	border-radius: 2px;
	padding: 0.1em;
	font-family: "monospace"
}
`

func addDefaultTemplate(book *epub.EPub) {
	book.AddStylesheet("css/styles.css", defaultCSS)
	log.Print("Added default stylesheet css/styles.css")
}
