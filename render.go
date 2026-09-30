package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/hatolife/HatoNote/internal/bundle"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	xhtml "golang.org/x/net/html"
)

var referenceBlockPattern = regexp.MustCompile("(?s)<!-- HATONOTE_REFERENCES_BEGIN -->\\s*(.*?)\\s*<!-- HATONOTE_REFERENCES_END -->")
var referenceIDPattern = regexp.MustCompile("<!-- HATONOTE_REF:([A-Za-z0-9_-]{1,80}) -->")
var citationPattern = regexp.MustCompile("\\[@([A-Za-z0-9_-]{1,80})\\]")
var safeClassPattern = regexp.MustCompile("^[A-Za-z0-9_ -]{1,200}$")
var safeColorPattern = regexp.MustCompile("(?i)^(#[0-9a-f]{3,8}|rgba?\\([0-9 ,.]+\\))$")
var safeFontSizePattern = regexp.MustCompile("^([89]|[1-8][0-9]|9[0-6])px$")

var allowedHTMLTags = map[string]bool{
	"p": true, "br": true, "hr": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"strong": true, "b": true, "em": true, "i": true, "del": true, "s": true, "strike": true,
	"a": true, "img": true, "ul": true, "ol": true, "li": true, "blockquote": true,
	"pre": true, "code": true, "table": true, "thead": true, "tbody": true, "tfoot": true,
	"tr": true, "th": true, "td": true, "input": true, "details": true, "summary": true, "span": true,
}

var droppedHTMLTags = map[string]bool{
	"script": true, "style": true, "iframe": true, "object": true, "embed": true, "svg": true, "math": true, "template": true,
}

type renderedReferences struct {
	source string
	body   string
	order  map[string]int
}

// renderMarkdown は許可済みHTMLだけを残し、HatoNoteの引用表示も適用します。
func renderMarkdown(text string) (string, error) {
	if len(text) > bundle.MaxFile {
		return "", fmt.Errorf("本文が大きすぎます")
	}
	body, references := splitReferences(text)
	html, err := renderSafeMarkdown(body)
	if err != nil {
		return "", err
	}
	html, err = decorateCitations(html, references.order)
	if err != nil {
		return "", err
	}
	if references.source == "" {
		return html, nil
	}
	referenceHTML, err := renderSafeMarkdown(references.body)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString([]byte(references.source))
	return html + "<section class=\"hatonote-references\" id=\"hatonote-references\" data-hatonote-source=\"" + encoded + "\">" + referenceHTML + "</section>", nil
}

// renderSafeMarkdown はGoldmarkでHTMLを生成した後、許可したDOMだけを再構築します。
func renderSafeMarkdown(text string) (string, error) {
	var output bytes.Buffer
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(gmhtml.WithUnsafe()),
	)
	if err := md.Convert([]byte(text), &output); err != nil {
		return "", err
	}
	return sanitizeHTML(output.String())
}

// splitReferences はHatoNote管理の参考文献ブロックを本文から分離します。
func splitReferences(text string) (string, renderedReferences) {
	match := referenceBlockPattern.FindStringSubmatchIndex(text)
	if match == nil {
		return text, renderedReferences{order: map[string]int{}}
	}
	source := text[match[0]:match[1]]
	inner := text[match[2]:match[3]]
	ids := referenceIDPattern.FindAllStringSubmatch(inner, -1)
	order := make(map[string]int, len(ids))
	for _, item := range ids {
		if _, exists := order[item[1]]; !exists {
			order[item[1]] = len(order) + 1
		}
	}
	clean := referenceIDPattern.ReplaceAllString(inner, "")
	body := strings.TrimSpace(text[:match[0]] + "\n" + text[match[1]:])
	if body != "" {
		body += "\n"
	}
	return body, renderedReferences{source: source, body: clean, order: order}
}

// sanitizeHTML は未許可タグを展開または破棄し、危険な属性を削除します。
func sanitizeHTML(fragment string) (string, error) {
	doc, err := xhtml.Parse(strings.NewReader("<!doctype html><html><body>" + fragment + "</body></html>"))
	if err != nil {
		return "", err
	}
	body := findHTMLBody(doc)
	if body == nil {
		return "", fmt.Errorf("HTML本文を生成できません")
	}
	safe := &xhtml.Node{Type: xhtml.ElementNode, Data: "body"}
	for child := body.FirstChild; child != nil; child = child.NextSibling {
		appendSanitizedHTML(safe, child)
	}
	var output bytes.Buffer
	for child := safe.FirstChild; child != nil; child = child.NextSibling {
		if err := xhtml.Render(&output, child); err != nil {
			return "", err
		}
	}
	return output.String(), nil
}

func findHTMLBody(node *xhtml.Node) *xhtml.Node {
	if node.Type == xhtml.ElementNode && node.Data == "body" {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if result := findHTMLBody(child); result != nil {
			return result
		}
	}
	return nil
}

// appendSanitizedHTML は許可タグだけを新しいDOMへコピーします。
func appendSanitizedHTML(parent, source *xhtml.Node) {
	switch source.Type {
	case xhtml.TextNode:
		parent.AppendChild(&xhtml.Node{Type: xhtml.TextNode, Data: source.Data})
		return
	case xhtml.ElementNode:
	default:
		return
	}
	tag := strings.ToLower(source.Data)
	if droppedHTMLTags[tag] {
		return
	}
	if !allowedHTMLTags[tag] {
		for child := source.FirstChild; child != nil; child = child.NextSibling {
			appendSanitizedHTML(parent, child)
		}
		return
	}
	node := &xhtml.Node{Type: xhtml.ElementNode, Data: tag}
	node.Attr = safeHTMLAttributes(tag, source.Attr)
	if tag == "input" && attributeValue(node.Attr, "type") != "checkbox" {
		return
	}
	parent.AppendChild(node)
	for child := source.FirstChild; child != nil; child = child.NextSibling {
		appendSanitizedHTML(node, child)
	}
}

func safeHTMLAttributes(tag string, attrs []xhtml.Attribute) []xhtml.Attribute {
	result := []xhtml.Attribute{}
	for _, attr := range attrs {
		name := strings.ToLower(attr.Key)
		value := strings.TrimSpace(attr.Val)
		switch name {
		case "href":
			if tag == "a" && safeURL(value) {
				result = append(result, xhtml.Attribute{Key: name, Val: value})
			}
		case "src":
			if tag == "img" && safeURL(value) {
				result = append(result, xhtml.Attribute{Key: name, Val: value})
			}
		case "alt", "title":
			if tag == "img" || tag == "a" {
				result = append(result, xhtml.Attribute{Key: name, Val: value})
			}
		case "id":
			if strings.HasPrefix(tag, "h") && len(value) <= 500 {
				result = append(result, xhtml.Attribute{Key: name, Val: value})
			}
		case "class":
			if safeClassPattern.MatchString(value) && (tag == "code" || tag == "ul" || tag == "li" || tag == "input") {
				result = append(result, xhtml.Attribute{Key: name, Val: value})
			}
		case "start":
			if tag == "ol" {
				if _, err := strconv.Atoi(value); err == nil {
					result = append(result, xhtml.Attribute{Key: name, Val: value})
				}
			}
		case "align":
			if (tag == "th" || tag == "td") && (value == "left" || value == "center" || value == "right") {
				result = append(result, xhtml.Attribute{Key: name, Val: value})
			}
		case "type":
			if tag == "input" && value == "checkbox" {
				result = append(result, xhtml.Attribute{Key: name, Val: value})
			}
		case "checked", "disabled":
			if tag == "input" {
				result = append(result, xhtml.Attribute{Key: name, Val: name})
			}
		case "open":
			if tag == "details" {
				result = append(result, xhtml.Attribute{Key: name, Val: ""})
			}
		case "style":
			if tag == "span" {
				if style := safeInlineStyle(value); style != "" {
					result = append(result, xhtml.Attribute{Key: name, Val: style})
				}
			}
		}
	}
	return result
}

func attributeValue(attrs []xhtml.Attribute, name string) string {
	for _, attr := range attrs {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}

func safeURL(value string) bool {
	if value == "" || strings.ContainsAny(value, "\x00\r\n") {
		return false
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	switch strings.ToLower(parsed.Scheme) {
	case "", "http", "https", "mailto":
		return true
	default:
		return false
	}
}

func safeInlineStyle(value string) string {
	result := []string{}
	for _, declaration := range strings.Split(value, ";") {
		parts := strings.SplitN(declaration, ":", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		switch name {
		case "color":
			if safeColorPattern.MatchString(val) {
				result = append(result, "color:"+val)
			}
		case "font-size":
			if safeFontSizePattern.MatchString(val) {
				result = append(result, "font-size:"+val)
			}
		}
	}
	return strings.Join(result, ";")
}

// decorateCitations はコード領域以外の引用記法を番号付き参照へ置換します。
func decorateCitations(fragment string, order map[string]int) (string, error) {
	doc, err := xhtml.Parse(strings.NewReader("<!doctype html><html><body>" + fragment + "</body></html>"))
	if err != nil {
		return "", err
	}
	body := findHTMLBody(doc)
	if body == nil {
		return "", fmt.Errorf("HTML本文を生成できません")
	}
	decorateCitationChildren(body, order, false)
	var output bytes.Buffer
	for child := body.FirstChild; child != nil; child = child.NextSibling {
		if err := xhtml.Render(&output, child); err != nil {
			return "", err
		}
	}
	return output.String(), nil
}

func decorateCitationChildren(node *xhtml.Node, order map[string]int, code bool) {
	for child := node.FirstChild; child != nil; {
		next := child.NextSibling
		if child.Type == xhtml.ElementNode {
			tag := strings.ToLower(child.Data)
			decorateCitationChildren(child, order, code || tag == "code" || tag == "pre")
		} else if child.Type == xhtml.TextNode && !code {
			replaceCitationText(child, order)
		}
		child = next
	}
}

func replaceCitationText(node *xhtml.Node, order map[string]int) {
	matches := citationPattern.FindAllStringSubmatchIndex(node.Data, -1)
	if len(matches) == 0 || node.Parent == nil {
		return
	}
	parent := node.Parent
	offset := 0
	for _, match := range matches {
		if match[0] > offset {
			parent.InsertBefore(&xhtml.Node{Type: xhtml.TextNode, Data: node.Data[offset:match[0]]}, node)
		}
		id := node.Data[match[2]:match[3]]
		label := "引用"
		if number, ok := order[id]; ok {
			label = strconv.Itoa(number)
		}
		sup := &xhtml.Node{Type: xhtml.ElementNode, Data: "sup", Attr: []xhtml.Attribute{{Key: "data-hatonote-citation", Val: id}}}
		link := &xhtml.Node{Type: xhtml.ElementNode, Data: "a", Attr: []xhtml.Attribute{{Key: "href", Val: "#hatonote-references"}}}
		link.AppendChild(&xhtml.Node{Type: xhtml.TextNode, Data: "[" + label + "]"})
		sup.AppendChild(link)
		parent.InsertBefore(sup, node)
		offset = match[1]
	}
	if offset < len(node.Data) {
		parent.InsertBefore(&xhtml.Node{Type: xhtml.TextNode, Data: node.Data[offset:]}, node)
	}
	parent.RemoveChild(node)
}

// Render は安全なMarkdownと許可済みHTML書式だけを変換します。
func (a *App) Render(text string) (string, error) {
	return renderMarkdown(text)
}
