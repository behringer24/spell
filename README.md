# spell
Smart Processing and Enhanced Lightweight Layout - Markdown to epub converter

## Meaning:
- Smart Processing: The project offers intelligent processing of Markdown files.
- Processing: The conversion of Markdown to EPUB3 and the ability to use enhanced commands like includes.
- Enhanced Layout: The generated EPUB files are well-structured and can be styled with templates.
- Lightweight: The system's modularity and simplicity are key features.
- Layout: The system ensures clean EPUB generation.

## Why SPELL?
SPELL evokes "magic spells" and therefore aligns well with the "magical" way your system extends Markdown files with additional commands and converts them into EPUB.

It's a short and memorable name that suggests the program is both flexible and easy to use, while offering powerful features.

## Why SPELL as another tool
I used asciidoctor, latex, Sigil and other tools to write books and ebooks. None of them really satified my needs for easy writing and nice looking results.

latex generates outstanding PDF files but is not good for generating ebooks and it is a behemoth - if I would aim for print that would still be my go-to-system.

asciidoctor/asciidoc is a good system but lacks some basic features like language specific quote characters - and I am no big fan of ruby and having to install the whole setup - if I am no ruby dev it means a certain overhead to deal with this.

Sigil is a Windows UI WYSIWYG editor for epub files - a great tool, but not for me as an author to write longer books and where I want to focus on the contents.

Go as programming language compiles to system specific binary executables and until now this means spell is only *one* file to download and run - and it is _fast_. Probably one day I will also provide msi installer packages for windows to make it even more simple to run it.

Also i want to expand the set of available commands to my (and hopefully many other) needs. And I would be glad to colaborate on this. If you find an elegant way to improve the architecture of spell (especially the parser I am unhappy with) or if you like to implement more commands, feel free to fork and send me the merge requests for your improvements.

# Installation
## Prebuild binaries
The project features an autobuild action for the different OS and architectures. You can download the files under [Releases](https://github.com/behringer24/spell/releases) and there under *Assets*.

## Build from source
Download or checkout the files from the repo and build them with:
```
go build .
```

# Usage
You find detailed documentation for the *spell* syntax in [the Github Wiki](https://github.com/behringer24/spell/wiki)
## General usage help
You can get a general help by calling:
```
./spell.exe -h
```
You will get a help overview over all command line parameters, like:
```
spell
Smart Processing and Enhanced Lightweight Layout

Usage: spell [-h] [-v] [-e] infile [outfile]

Flags:
-h, --help               Show this help text
-v, --version            Show version information

Options:
-e, --epub               Generate epub version 2 or 3 (Default: 3)

Positional arguments:
infile                   File to read from
outfile                  File to write to (Default: ./ebook.epub)
```
The simplest way to call *spell* is to give it a Markdown file:
```
./spell.exe example.md
```
This will make spell parse the file `example.md` and generate a file `ebook.epub` (default value for the output file) in the same folder.

## Version information
To check for the currently installed version:
```
./spell.exe -v
```
You will get a version number, like:
```
SPELL version: v0.0.2
```

## Example
You find an example-folder in this repo. Download all files and compile them with:
```
./spell.exe example
```
The example shows a lot of the built in features of *spell*.
