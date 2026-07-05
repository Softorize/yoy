package yahoo

import (
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var wsRun = regexp.MustCompile(`[ \t\f\r\x{00a0}]+`)

// HTMLToText renders an HTML body into readable plain text. It drops
// script/style/head content, inserts line breaks around block-level
// elements, unescapes entities, and collapses runs of whitespace.
func HTMLToText(htmlStr string) string {
	z := html.NewTokenizer(strings.NewReader(htmlStr))
	var b strings.Builder
	skip := 0 // >0 while inside script/style/head/title/noscript

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			break // includes io.EOF
		}
		switch tt {
		case html.StartTagToken, html.SelfClosingTagToken:
			name, _ := z.TagName()
			switch atom.Lookup(name) {
			case atom.Script, atom.Style, atom.Head, atom.Title, atom.Noscript:
				if tt == html.StartTagToken {
					skip++
				}
			case atom.Br:
				b.WriteByte('\n')
			case atom.P, atom.Div, atom.Li, atom.Tr, atom.Ul, atom.Ol,
				atom.Blockquote, atom.Table, atom.Section, atom.Article,
				atom.Header, atom.Footer, atom.H1, atom.H2, atom.H3,
				atom.H4, atom.H5, atom.H6:
				b.WriteByte('\n')
			}
		case html.EndTagToken:
			name, _ := z.TagName()
			switch atom.Lookup(name) {
			case atom.Script, atom.Style, atom.Head, atom.Title, atom.Noscript:
				if skip > 0 {
					skip--
				}
			case atom.P, atom.Div, atom.Li, atom.Tr, atom.Blockquote,
				atom.Table, atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
				b.WriteByte('\n')
			}
		case html.TextToken:
			if skip == 0 {
				b.Write(z.Text())
			}
		}
	}

	return collapseWhitespace(b.String())
}

// collapseWhitespace trims each line, squeezes horizontal whitespace runs to a
// single space, and limits consecutive blank lines to one.
func collapseWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	blank := 0
	for _, ln := range lines {
		ln = strings.TrimSpace(wsRun.ReplaceAllString(ln, " "))
		if ln == "" {
			blank++
			if blank > 1 {
				continue
			}
		} else {
			blank = 0
		}
		out = append(out, ln)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
