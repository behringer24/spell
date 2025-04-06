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
	text-align: left;
	color: #666666;
	page-break-after:avoid;
	page-break-inside: avoid;
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
	margin-right: 3em;
	padding: 1em;
}
blockquote.code {
	font-family: "monospace";
	background-color: #ccc;
	border-radius: 5px;
	text-align: left;
}
blockquote.cite {
	font-size:120%;
	border: 1px solid #888;
	border-width: 1px 0;
	padding: 0.5em;
	position: relative;	
}
blockquote.cite::before {
	content: url("data:image/svg+xml,%3C%3Fxml version='1.0' encoding='UTF-8' standalone='no'%3F%3E%3Csvg width='14.731999mm' height='10.160001mm' viewBox='0 0 14.731999 10.160001' version='1.1' id='svg1' xmlns='http://www.w3.org/2000/svg' xmlns:svg='http://www.w3.org/2000/svg'%3E%3Cdefs id='defs1' /%3E%3Cg id='layer1' transform='translate(-70.44453,-66.836947)'%3E%3Cpath fill='%23CCC' d='m 72.98453,72.323347 q -0.9652,0 -1.778,-0.6096 -0.762,-0.6604 -0.762,-1.6256 0,-1.4732 1.1176,-2.3368 1.1684,-0.9144 2.6416,-0.9144 1.4732,0 2.286,0.9144 0.8128,0.9144 0.8128,2.3876 0,1.9304 -1.4224,4.2672 -1.5748,2.5908 -3.4036,2.5908 -0.9652,0 -0.9652,-0.8128 0.8636,-0.8128 1.5748,-1.8288 0.762,-1.016 0.9144,-2.2352 z m 7.8232,0 q -0.9144,0 -1.7272,-0.6096 -0.762,-0.6604 -0.762,-1.6256 0,-1.4732 1.1176,-2.3368 1.1684,-0.9144 2.6924,-0.9144 1.4224,0 2.235199,0.9144 0.8128,0.9144 0.8128,2.3876 0,2.032 -1.371599,4.2672 -1.5748,2.5908 -3.4544,2.5908 -0.9652,0 -0.9652,-0.8128 0.8128,-0.8128 1.524,-1.778 0.762,-0.9652 0.9652,-2.286 -0.6096,0.2032 -1.0668,0.2032 z' /%3E%3C/g%3E%3C/svg%3E%0A");
	position: absolute;
	height: 80px;
	top: -25px;
	left: -25px;
  	z-index: -1;
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
