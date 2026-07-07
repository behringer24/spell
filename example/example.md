$[title](Spell Example Book)
$[author](Jane Doe)
$[language](EN)
$[uuid](18b04cfb-6d81-405f-833f-aa5182212bbb)
$[series](Spell Demos)
$[date](2026-06-25)
$[rights](© 2026 Jane Doe. All rights reserved.)
$[source](https://github.com/behringer24/spell)
$[relation](https://github.com/behringer24/spell/wiki)
$[type](Text)
$[entry](1)
$[quotes](&raquo;,&laquo;,&rsaquo;,&lsaquo;)
![cover](cover.jpg "Cover image")

%toc(Table of Contents)

# Introduction

This book demonstrates the features of the **spell** markdown-to-epub converter.

Build this file on Windows with:
```
..\spell.exe example.md example.epub
```

Or with a custom stylesheet:
```
..\spell.exe -s custom.css example.md example.epub
```

For more details see: https://github.com/behringer24/spell

Footnotes are written with a reference[^intro] and a matching definition; on
supporting readers they appear as tap-to-open popups.[^popup]

[^intro]: This is a footnote. Its definition may sit anywhere in the chapter.
[^popup]: On Kindle and EPUB3 readers the note opens in place instead of jumping.

# Text Formatting

Spell supports the usual inline formatting: **bold**, *italic* and `inline code`.

Typographic quotes use the configured quote characters: %"double quoted"% and %'single quoted'% text.

Long dashes --- like this --- and short dashes -- like this -- are converted automatically.
Ellipsis... are converted too.

A backslash escapes any special character so it is shown literally: \*not italic\*,
\# not a heading, and a literal backslash before a letter like C:\path stays as is.

## Lists

Unordered lists use `-`, `*` or `+` and nest by indentation:

- First item
- Second item
  - Nested item
  - Another nested item
    - Even deeper
- Third item

Ordered lists use `1.` or `1)` and can be nested and mixed with bullets:

1. Step one
2. Step two
   1. Sub-step
   2. Another sub-step
3. Step three

## Links

Link to an external resource with the familiar markdown syntax:
[the spell repository](https://github.com/behringer24/spell) or with a tooltip
[the wiki](https://github.com/behringer24/spell/wiki "Documentation").

## Block Types

``` cite
The only way to do great work is to love what you do.
(Steve Jobs)
```

``` note
This is a **note** block.
It can span multiple lines.
```

``` info
This is a single-line **info** block.
```

``` warn
This is a **warning** block — pay attention to this.
```

# Included Content

The following chapter is pulled in from a separate file using the include syntax:

![include](example2.md)

# Anchors and Internal Links

Anchors mark a position in the text invisibly. Links can jump to any anchor within the
same chapter or across chapter boundaries.

## Defining Anchors

Place an anchor anywhere in a line using `{#id}`:

{#anchor-intro}
This paragraph is marked with the anchor `anchor-intro`. A link elsewhere in this book
can jump directly here.

## Linking to Anchors

Links to anchors use the familiar markdown link syntax with a `#` prefix.

Jump to a section in the **same chapter**: [Back to anchor intro](#anchor-intro)

Jump to a section in **another chapter**: [Go to the cross-chapter target](#cross-chapter-target)

# Index Entries

Spell can collect index entries from the running text and generate a clickable keyword
index at the end of the book.

## Basic Index Entry

Use `%[term](IndexName)` to mark a term. The word appears normally in the text and is
registered in the named index.

The %[printing press](Technology) was invented by Johannes Gutenberg around 1440. It
transformed the way knowledge was spread. The %[movable type](Technology) system he
developed made book production vastly more efficient.

## Canonical Term — Grouping Variants

Use `%[display term](IndexName|canonical)` when the word in the text differs from the
desired index entry, for example singular vs. plural forms, or inflected variants.

The first %[steam engine](Technology|Steam engine) was a milestone of the industrial
revolution. By the mid-1800s %[steam engines](Technology|Steam engine) were driving
factories, ships and railways. The efficiency of these %[steam-powered machines](Technology|Steam engine)
improved steadily throughout the century.

All three variants appear under a single **Steam engine** entry in the index with
numbered links to each occurrence.

## Multiple Indexes

Entries can belong to different indexes at the same time. Here the same sentence
contributes to both a technology and a persons index:

%[James Watt](Persons) significantly improved the %[steam engine](Technology|Steam engine)
in the 1760s, making it practical for widespread industrial use.

%[George Stephenson](Persons) applied %[steam power](Technology|Steam engine) to railways,
building the first public steam railway in 1825.

{#cross-chapter-target}
This paragraph is the cross-chapter link target referenced from the previous chapter.
[Back to where the link came from](#anchor-intro)

%index[Technology](Technical Terms)

%index[Persons](Notable Persons)
