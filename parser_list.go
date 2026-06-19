package main

import "log"

func listCloseHandler() lineHandler {
	return lineHandler{
		match: func(line string, insideBlock bool) bool {
			return inUlList && !reUlList.MatchString(line) && !insideBlock
		},
		handle: func(ctx *parseContext, line string, _ bool) (string, bool) {
			inUlList = false
			return "</ul>\n" + parseLine(ctx, line, false), true
		},
	}
}

func listItemHandler() lineHandler {
	return lineHandler{
		match: func(line string, insideBlock bool) bool { return reUlList.MatchString(line) },
		handle: func(ctx *parseContext, line string, _ bool) (string, bool) {
			matches := reUlList.FindStringSubmatch(line)
			newline := ""
			if !inUlList {
				newline = "<ul>\n"
				inUlList = true
			}
			log.Print("Add LI Element")
			return newline + "  <li>" + parseLine(ctx, matches[1], true) + "</li>\n", true
		},
	}
}
