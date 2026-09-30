package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	stdhtml "html"
	"mime"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"time"

	"github.com/hatolife/HatoNote/internal/bundle"
	"github.com/hatolife/HatoNote/internal/process"
	"github.com/hatolife/HatoNote/internal/settings"
	"github.com/hatolife/HatoNote/internal/workspace"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	xhtml "golang.org/x/net/html"
)

func cloneExportDocument(doc *bundle.Document) *bundle.Document {
	files := make(map[string][]byte, len(doc.Files))
	for name, data := range doc.Files {
		files[name] = append([]byte(nil), data...)
	}
	manifest := make(map[string]json.RawMessage, len(doc.Manifest))
	for key, value := range doc.Manifest {
		manifest[key] = append(json.RawMessage(nil), value...)
	}
	return &bundle.Document{Files: files, Manifest: manifest, Entry: doc.Entry}
}

func exportDocumentTitle(filename string, doc *bundle.Document) string {
	if doc != nil && doc.HasSlides() {
		if deck, err := doc.Deck(); err == nil && strings.TrimSpace(deck.Title) != "" {
			return deck.Title
		}
	}
	if filename != "" {
		name := filepath.Base(filename)
		if ext := filepath.Ext(name); ext != "" {
			name = strings.TrimSuffix(name, ext)
		}
		if name != "" {
			return name
		}
	}
	return "HatoNote Document"
}

func exportPageOrder(doc *bundle.Document) []string {
	if doc == nil {
		return nil
	}
	if src, err := detectBook(doc); err == nil && src != "" {
		summary := path.Join(src, "SUMMARY.md")
		entries := parseContents(doc.Files[summary], src, doc.Files)
		seen := map[string]bool{}
		pages := []string{}
		for _, entry := range entries {
			if entry.Name == "" || entry.Missing || seen[entry.Name] {
				continue
			}
			seen[entry.Name] = true
			pages = append(pages, entry.Name)
		}
		for _, name := range doc.Pages() {
			if name == summary || seen[name] {
				continue
			}
			seen[name] = true
			pages = append(pages, name)
		}
		return pages
	}
	return doc.Pages()
}

func exportAnchor(index int) string {
	return "hatonote-page-" + strconv.Itoa(index+1)
}

func exportResolveReference(reference, source string) (string, string, bool) {
	parsed, err := url.Parse(reference)
	if err != nil || parsed.Scheme != "" || parsed.Host != "" || strings.HasPrefix(parsed.Path, "/") {
		return "", "", false
	}
	decoded, err := url.PathUnescape(parsed.Path)
	if err != nil {
		return "", "", false
	}
	if decoded == "" {
		return source, parsed.Fragment, true
	}
	target := path.Clean(path.Join(path.Dir(source), decoded))
	if !bundle.ValidPath(target) {
		return "", "", false
	}
	fragment, _ := url.PathUnescape(parsed.Fragment)
	return target, fragment, true
}

func htmlAttribute(node *xhtml.Node, key string) (string, int) {
	for index := range node.Attr {
		if node.Attr[index].Key == key {
			return node.Attr[index].Val, index
		}
	}
	return "", -1
}

func setHTMLAttribute(node *xhtml.Node, key, value string) {
	if _, index := htmlAttribute(node, key); index >= 0 {
		node.Attr[index].Val = value
		return
	}
	node.Attr = append(node.Attr, xhtml.Attribute{Key: key, Val: value})
}

func transformExportHTML(fragment, source, anchor string, anchors map[string]string, doc *bundle.Document) (string, error) {
	root, err := xhtml.Parse(strings.NewReader("<!doctype html><html><body>" + fragment + "</body></html>"))
	if err != nil {
		return "", err
	}
	var body *xhtml.Node
	var walk func(*xhtml.Node)
	walk = func(node *xhtml.Node) {
		if node.Type == xhtml.ElementNode && node.Data == "body" {
			body = node
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	if body == nil {
		return "", fmt.Errorf("HTML本文を生成できません")
	}
	var transform func(*xhtml.Node)
	transform = func(node *xhtml.Node) {
		if node.Type == xhtml.ElementNode {
			if id, index := htmlAttribute(node, "id"); index >= 0 && id != "" {
				node.Attr[index].Val = anchor + "-" + id
			}
			switch node.Data {
			case "img":
				if sourceURL, _ := htmlAttribute(node, "src"); sourceURL != "" {
					if target, _, ok := exportResolveReference(sourceURL, source); ok {
						if data, exists := doc.Files[target]; exists {
							kind := mime.TypeByExtension(strings.ToLower(path.Ext(target)))
							if kind == "" {
								kind = "application/octet-stream"
							}
							setHTMLAttribute(node, "src", "data:"+kind+";base64,"+base64.StdEncoding.EncodeToString(data))
						}
					}
				}
			case "a":
				if href, _ := htmlAttribute(node, "href"); href != "" {
					if target, fragment, ok := exportResolveReference(href, source); ok {
						if targetAnchor, exists := anchors[target]; exists {
							next := "#" + targetAnchor
							if fragment != "" {
								next += "-" + fragment
							}
							setHTMLAttribute(node, "href", next)
						} else if target == source && fragment != "" {
							setHTMLAttribute(node, "href", "#"+anchor+"-"+fragment)
						}
					}
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			transform(child)
		}
	}
	transform(body)
	var output bytes.Buffer
	for child := body.FirstChild; child != nil; child = child.NextSibling {
		if err := xhtml.Render(&output, child); err != nil {
			return "", err
		}
	}
	return output.String(), nil
}

func exportFontFamily(value string) string {
	switch value {
	case "gothic":
		return "'Yu Gothic UI','Yu Gothic',Meiryo,sans-serif"
	case "mincho":
		return "'Yu Mincho','MS Mincho',serif"
	case "monospace":
		return "'Cascadia Mono',Consolas,monospace"
	default:
		return "'Segoe UI','Yu Gothic UI','Yu Gothic',Meiryo,sans-serif"
	}
}

func buildDocumentHTML(doc *bundle.Document, title string) ([]byte, error) {
	pages := exportPageOrder(doc)
	if len(pages) == 0 {
		return nil, fmt.Errorf("エクスポートするMarkdownページがありません")
	}
	anchors := map[string]string{}
	for index, pageName := range pages {
		anchors[pageName] = exportAnchor(index)
	}
	var body strings.Builder
	body.WriteString("<nav class=\"hatonote-toc\"><h1>" + stdhtml.EscapeString(title) + "</h1><ol>")
	for _, pageName := range pages {
		body.WriteString("<li><a href=\"#" + anchors[pageName] + "\">" + stdhtml.EscapeString(pageName) + "</a></li>")
	}
	body.WriteString("</ol></nav>")
	for _, pageName := range pages {
		text, ok := doc.Files[pageName]
		if !ok {
			continue
		}
		rendered, err := renderMarkdown(string(text))
		if err != nil {
			return nil, err
		}
		rendered, err = transformExportHTML(rendered, pageName, anchors[pageName], anchors, doc)
		if err != nil {
			return nil, err
		}
		body.WriteString("<article class=\"hatonote-page\" id=\"" + anchors[pageName] + "\"><header class=\"hatonote-page-path\">" + stdhtml.EscapeString(pageName) + "</header>" + rendered + "</article>")
	}
	css := `
:root{color-scheme:light dark}*{box-sizing:border-box}body{margin:0;background:#f5f7fa;color:#202938;font:16px/1.75 "Segoe UI","Yu Gothic",Meiryo,sans-serif}.hatonote-toc,.hatonote-page{max-width:920px;margin:28px auto;padding:42px 54px;background:#fff;color:#202938;box-shadow:0 2px 16px #1f293712}.hatonote-toc ol{columns:2;padding-left:1.5em}.hatonote-page-path{color:#718096;font-size:12px;border-bottom:1px solid #d8e0ea;padding-bottom:10px;margin-bottom:28px}h1,h2,h3,h4,h5{line-height:1.3}pre{overflow:auto;padding:16px;background:#eef2f7;border-radius:7px}code{font-family:Consolas,monospace}img{max-width:100%;height:auto}table{border-collapse:collapse;width:100%}th,td{padding:7px 10px;border:1px solid #ccd5e0;text-align:left}blockquote{margin-left:0;padding-left:16px;border-left:4px solid #8aa4ce;color:#526176}a{color:#2563c7}.math-render{display:flex;justify-content:center;overflow:auto;margin:1em 0}.math-render math{font-size:1.2em}@media print{@page{size:A4;margin:14mm}body{background:#fff}.hatonote-toc,.hatonote-page{max-width:none;margin:0;padding:0;box-shadow:none}.hatonote-toc{break-after:page}.hatonote-page{break-after:page}.hatonote-page:last-child{break-after:auto}}`
	html := "<!doctype html><html lang=\"ja\"><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\"><title>" + stdhtml.EscapeString(title) + "</title><style>" + css + "</style></head><body>" + body.String() + "</body></html>"
	return []byte(html), nil
}

func buildSlidesHTML(doc *bundle.Document, title string) ([]byte, error) {
	deck, err := doc.Deck()
	if err != nil {
		return nil, err
	}
	anchors := map[string]string{}
	for index, slide := range deck.Slides {
		anchors[slide.File] = exportAnchor(index)
	}
	height := 540
	pageWidth := "13.333in"
	if deck.Aspect == "4:3" {
		height = 720
		pageWidth = "10in"
	}
	marginX, marginY := deck.ContentMarginX, deck.ContentMarginY
	if marginX == 0 {
		marginX = 60
	}
	if marginY == 0 {
		marginY = 48
	}
	bodySize := deck.BodyFontSize
	if bodySize == 0 {
		bodySize = 20
	}
	var body strings.Builder
	for index, slide := range deck.Slides {
		rendered, err := renderSlideMarkdown(string(doc.Files[slide.File]), slide.Layout)
		if err != nil {
			return nil, err
		}
		rendered, err = transformExportHTML(rendered, slide.File, anchors[slide.File], anchors, doc)
		if err != nil {
			return nil, err
		}
		background := "#ffffff"
		color := "#202b3e"
		if deck.Theme == "dark" {
			background, color = "#182333", "#e9eff8"
		}
		if slide.Background != "" {
			background = slide.Background
		}
		style := fmt.Sprintf("--slide-margin-x:%dpx;--slide-margin-y:%dpx;font-size:%dpx;font-family:%s;background:%s;color:%s", marginX, marginY, bodySize, exportFontFamily(deck.FontFamily), background, color)
		for level, size := range []int{deck.H1FontSize, deck.H2FontSize, deck.H3FontSize, deck.H4FontSize, deck.H5FontSize} {
			if size > 0 {
				style += fmt.Sprintf(";--slide-h%d-size:%dpx", level+1, size)
			}
		}
		body.WriteString("<section class=\"export-slide\" id=\"" + anchors[slide.File] + "\" data-layout=\"" + stdhtml.EscapeString(slide.Layout) + "\" style=\"" + stdhtml.EscapeString(style) + "\"><div class=\"slide-content\">" + rendered + "</div>")
		if deck.PageNumberEnabled && !slide.HidePageNumber {
			position := deck.PageNumberPosition
			if position == "" {
				position = "bottom-right"
			}
			body.WriteString("<div class=\"slide-page-number\" data-position=\"" + stdhtml.EscapeString(position) + "\">" + strconv.Itoa(deck.PageNumberStart+index) + "</div>")
		}
		body.WriteString("</section>")
	}
	css := fmt.Sprintf(`
*{box-sizing:border-box}html,body{margin:0;background:%s;color:#202b3e}body{font-family:"Segoe UI","Yu Gothic",Meiryo,sans-serif;padding:24px}.export-slide{position:relative;width:960px;height:%dpx;margin:0 auto 28px;padding:var(--slide-margin-y) var(--slide-margin-x);overflow:hidden;overflow-wrap:anywhere;box-shadow:0 3px 18px #0002}.slide-content{height:100%%;min-height:0}h1,h2,h3,h4,h5{font-weight:700;line-height:1.25;margin:0 0 .65em}h1{font-size:var(--slide-h1-size,1.65em)}h2{font-size:var(--slide-h2-size,1.3em)}h3{font-size:var(--slide-h3-size,1.12em)}h4{font-size:var(--slide-h4-size,1em)}h5{font-size:var(--slide-h5-size,.9em)}p{margin:.65em 0}ul,ol{margin:.7em 0;padding-left:1.4em}li{margin:.35em 0}img{display:block;max-width:100%%;max-height:340px;object-fit:contain;margin:16px auto}pre{font-size:.68em;line-height:1.5;overflow:hidden;padding:20px;background:#8497b522;border-radius:8px}table{border-collapse:collapse;width:100%%;font-size:.8em}td,th{border-bottom:1px solid #8497b566;padding:.35em .5em;text-align:left}.slide-columns{display:grid;grid-template-columns:1fr 1fr;gap:40px;height:100%%}.export-slide[data-layout=cover] .slide-content{display:flex;flex-direction:column;justify-content:center;text-align:center}.export-slide[data-layout=image] .slide-content{display:flex;flex-direction:column}.export-slide[data-layout=image] .slide-content>p:has(img){flex:1;min-height:0;margin:0;display:flex;align-items:center;justify-content:center}.export-slide[data-layout=image] img{max-height:100%%;height:100%%;margin:0}.slide-page-number{position:absolute;z-index:4;font:14px/1.2 "Segoe UI",sans-serif;opacity:.7}.slide-page-number[data-position=top-left]{top:18px;left:24px}.slide-page-number[data-position=top-center]{top:18px;left:50%%;transform:translateX(-50%%)}.slide-page-number[data-position=top-right]{top:18px;right:24px}.slide-page-number[data-position=bottom-left]{bottom:18px;left:24px}.slide-page-number[data-position=bottom-center]{bottom:18px;left:50%%;transform:translateX(-50%%)}.slide-page-number[data-position=bottom-right]{bottom:18px;right:24px}.math-render{display:flex;justify-content:center;overflow:hidden;margin:.7em 0}.math-render math{font-size:1.2em}@media(max-width:1000px){body{padding:0}.export-slide{width:100vw;height:auto;aspect-ratio:%s;margin:0 0 16px}}@media print{@page{size:%s 7.5in;margin:0}body{padding:0;background:#fff}.export-slide{width:100vw;height:100vh;margin:0;box-shadow:none;break-after:page}.export-slide:last-child{break-after:auto}}`, deck.MarginColor, height, strings.ReplaceAll(deck.Aspect, ":", "/"), pageWidth)
	html := "<!doctype html><html lang=\"ja\"><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\"><title>" + stdhtml.EscapeString(title) + "</title><style>" + css + "</style></head><body>" + body.String() + "</body></html>"
	return []byte(html), nil
}

func buildSingleHTML(doc *bundle.Document, title string) ([]byte, error) {
	if doc == nil {
		return nil, fmt.Errorf("文書がありません")
	}
	if doc.HasSlides() {
		return buildSlidesHTML(doc, title)
	}
	return buildDocumentHTML(doc, title)
}

func exportDiagramKind(code *xhtml.Node) string {
	if code == nil || code.Type != xhtml.ElementNode || code.Data != "code" {
		return ""
	}
	className, _ := htmlAttribute(code, "class")
	for _, class := range strings.Fields(className) {
		switch class {
		case "language-mermaid":
			return "mermaid"
		case "language-plantuml":
			return "plantuml"
		case "language-puml":
			return "puml"
		}
	}
	return ""
}

func exportNodeText(node *xhtml.Node) string {
	var output strings.Builder
	var walk func(*xhtml.Node)
	walk = func(current *xhtml.Node) {
		if current.Type == xhtml.TextNode {
			output.WriteString(current.Data)
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return output.String()
}

func exportMathCode(code *xhtml.Node) bool {
	if code == nil || code.Type != xhtml.ElementNode || code.Data != "code" {
		return false
	}
	className, _ := htmlAttribute(code, "class")
	for _, class := range strings.Fields(className) {
		switch class {
		case "language-math", "language-tex", "language-latex":
			return true
		}
	}
	return false
}

func exportMathElement(markup string) (*xhtml.Node, error) {
	root, err := xhtml.Parse(strings.NewReader("<!doctype html><html><body><div class=\"math-render math-display\">" + markup + "</div></body></html>"))
	if err != nil {
		return nil, err
	}
	var result *xhtml.Node
	var walk func(*xhtml.Node)
	walk = func(node *xhtml.Node) {
		if result == nil && node.Type == xhtml.ElementNode && node.Data == "div" {
			if className, _ := htmlAttribute(node, "class"); strings.Contains(className, "math-render") {
				result = node
				return
			}
		}
		for child := node.FirstChild; child != nil && result == nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	if result == nil {
		return nil, fmt.Errorf("MathML要素を生成できません")
	}
	if parent := result.Parent; parent != nil {
		parent.RemoveChild(result)
	}
	return result, nil
}

func exportElementOnlyText(node *xhtml.Node) bool {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == xhtml.ElementNode {
			return false
		}
	}
	return true
}

func enhanceExportMath(data []byte, cfg settings.Settings) ([]byte, error) {
	root, err := xhtml.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	var blocks []*xhtml.Node
	var paragraphs []*xhtml.Node
	var walk func(*xhtml.Node)
	walk = func(node *xhtml.Node) {
		if node.Type == xhtml.ElementNode && node.Data == "pre" {
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if exportMathCode(child) {
					blocks = append(blocks, node)
					break
				}
			}
		}
		if node.Type == xhtml.ElementNode && node.Data == "p" && exportElementOnlyText(node) {
			raw := strings.TrimSpace(exportNodeText(node))
			if len(raw) > 4 && strings.HasPrefix(raw, "$") && strings.HasSuffix(raw, "$") {
				paragraphs = append(paragraphs, node)
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	for _, pre := range blocks {
		var code *xhtml.Node
		for child := pre.FirstChild; child != nil; child = child.NextSibling {
			if exportMathCode(child) {
				code = child
				break
			}
		}
		if code == nil {
			continue
		}
		markup, renderErr := renderMath(exportNodeText(code), true, cfg)
		if renderErr != nil {
			setHTMLAttribute(pre, "title", renderErr.Error())
			setHTMLAttribute(pre, "data-math-error", "true")
			continue
		}
		element, parseErr := exportMathElement(markup)
		if parseErr != nil {
			return nil, parseErr
		}
		if parent := pre.Parent; parent != nil {
			parent.InsertBefore(element, pre)
			parent.RemoveChild(pre)
		}
	}
	for _, paragraph := range paragraphs {
		raw := strings.TrimSpace(exportNodeText(paragraph))
		source := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(raw, "$"), "$"))
		markup, renderErr := renderMath(source, true, cfg)
		if renderErr != nil {
			setHTMLAttribute(paragraph, "title", renderErr.Error())
			setHTMLAttribute(paragraph, "data-math-error", "true")
			continue
		}
		element, parseErr := exportMathElement(markup)
		if parseErr != nil {
			return nil, parseErr
		}
		if parent := paragraph.Parent; parent != nil {
			parent.InsertBefore(element, paragraph)
			parent.RemoveChild(paragraph)
		}
	}
	var output bytes.Buffer
	if err := xhtml.Render(&output, root); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func enhanceExportDiagrams(data []byte, cfg settings.Settings) ([]byte, error) {
	root, err := xhtml.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	var blocks []*xhtml.Node
	var walk func(*xhtml.Node)
	walk = func(node *xhtml.Node) {
		if node.Type == xhtml.ElementNode && node.Data == "pre" {
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if exportDiagramKind(child) != "" {
					blocks = append(blocks, node)
					break
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	for _, pre := range blocks {
		var code *xhtml.Node
		for child := pre.FirstChild; child != nil; child = child.NextSibling {
			if exportDiagramKind(child) != "" {
				code = child
				break
			}
		}
		kind := exportDiagramKind(code)
		if kind == "" {
			continue
		}
		value, renderErr := renderDiagram(kind, exportNodeText(code), cfg)
		if renderErr != nil {
			setHTMLAttribute(pre, "title", renderErr.Error())
			setHTMLAttribute(pre, "data-diagram-error", "true")
			continue
		}
		figure := &xhtml.Node{Type:xhtml.ElementNode, Data:"figure", Attr:[]xhtml.Attribute{{Key:"class", Val:"diagram-render"}}}
		image := &xhtml.Node{Type:xhtml.ElementNode, Data:"img", Attr:[]xhtml.Attribute{{Key:"src", Val:value}, {Key:"alt", Val:"diagram"}}}
		figure.AppendChild(image)
		if parent := pre.Parent; parent != nil {
			parent.InsertBefore(figure, pre)
			parent.RemoveChild(pre)
		}
	}
	var output bytes.Buffer
	if err := xhtml.Render(&output, root); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func findEdgeExecutable() string {
	if executable, err := exec.LookPath("msedge"); err == nil {
		return executable
	}
	candidates := []string{
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "Edge", "Application", "msedge.exe"),
		filepath.Join(os.Getenv("ProgramFiles"), "Microsoft", "Edge", "Application", "msedge.exe"),
		filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Edge", "Application", "msedge.exe"),
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}

func exportPDF(data []byte, filename string) error {
	if goruntime.GOOS != "windows" {
		return fmt.Errorf("PDFエクスポートはWindowsで使用できます")
	}
	edge := findEdgeExecutable()
	if edge == "" {
		return fmt.Errorf("Microsoft Edgeが見つからないためPDFを生成できません")
	}
	dir, err := os.MkdirTemp("", "HatoNote-export-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	htmlFile := filepath.Join(dir, "document.html")
	if err := os.WriteFile(htmlFile, data, 0600); err != nil {
		return err
	}
	profile := filepath.Join(dir, "edge-profile")
	absolute, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	filePath := filepath.ToSlash(htmlFile)
	if !strings.HasPrefix(filePath, "/") {
		filePath = "/" + filePath
	}
	fileURL := (&url.URL{Scheme:"file", Path:filePath}).String()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := process.CommandContext(ctx, edge,
		"--headless",
		"--disable-gpu",
		"--no-pdf-header-footer",
		"--user-data-dir="+profile,
		"--print-to-pdf="+absolute,
		fileURL,
	)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("PDF生成に失敗しました: %w\n%s", err, strings.TrimSpace(string(output)))
	}
	if info, err := os.Stat(absolute); err != nil || info.Size() == 0 {
		return fmt.Errorf("PDFファイルを生成できませんでした")
	}
	return nil
}

// Export は現在の作業内容を単一HTMLまたはPDFへ出力します。
func (a *App) Export(format string) (bool, error) {
	a.mu.Lock()
	if a.session == nil {
		a.mu.Unlock()
		return false, fmt.Errorf("文書を開いてください")
	}
	if err := a.syncNativeLocked(); err != nil {
		a.mu.Unlock()
		return false, err
	}
	if err := a.session.Capture(); err != nil {
		a.mu.Unlock()
		return false, err
	}
	doc := cloneExportDocument(a.session.Doc)
	source := a.session.Filename
	cfg := a.cfg
	a.mu.Unlock()

	title := exportDocumentTitle(source, doc)
	data, err := buildSingleHTML(doc, title)
	if err != nil {
		return false, err
	}
	data, err = enhanceExportDiagrams(data, cfg)
	if err != nil {
		return false, err
	}
	data, err = enhanceExportMath(data, cfg)
	if err != nil {
		return false, err
	}
	stem := "document"
	if source != "" {
		stem = strings.TrimSuffix(filepath.Base(source), filepath.Ext(source))
		if stem == "" {
			stem = "document"
		}
	}
	var extension, display string
	switch format {
	case "html":
		extension, display = ".html", "単一HTML"
	case "pdf":
		extension, display = ".pdf", "PDF"
	default:
		return false, fmt.Errorf("未対応のエクスポート形式です")
	}
	filename, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title: display + "へエクスポート",
		DefaultFilename: stem + extension,
		Filters: []wailsruntime.FileFilter{{DisplayName:display, Pattern:"*"+extension}},
	})
	if err != nil {
		return false, err
	}
	if filename == "" {
		return false, nil
	}
	if !strings.EqualFold(filepath.Ext(filename), extension) {
		filename += extension
	}
	if format == "html" {
		if err := workspace.AtomicWrite(filename, data); err != nil {
			return false, err
		}
	} else if err := exportPDF(data, filename); err != nil {
		return false, err
	}
	return true, nil
}
