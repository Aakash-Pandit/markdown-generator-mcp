# Markdown Generator MCP Server — Implementation Plan

## What This Builds

An MCP (Model Context Protocol) server written in Go that installs into Claude Code via a project-level `.mcp.json` file. Once running, you can say **"make a markdown"** or **"save this conversation as a doc"** inside Claude and it will automatically format and save a clean `.md` file to `./docs/` in your project.

---

## Project Structure

Create a new Go project with this layout:

```
mdgen-mcp/
├── internal/
│   ├── mcpserver/
│   │   └── server.go       ← registers MCP tools, starts stdio transport
│   ├── parser/
│   │   └── parser.go       ← converts raw text → structured Document
│   ├── formatter/
│   │   └── formatter.go    ← renders Document → markdown string
│   └── writer/
│       └── writer.go       ← saves .md file, creates ./docs/ if missing
├── main.go                 ← entry point
├── go.mod
└── .mcp.json               ← Claude Code picks this up automatically
```

---

## Step 1: Initialize the Project

```bash
mkdir mdgen-mcp && cd mdgen-mcp
go mod init mdgen-mcp
go get github.com/mark3labs/mcp-go@v0.49.0
go mod tidy
```

---

## Step 2: `.mcp.json` (Project Root)

Claude Code auto-detects this file — no changes to `claude_desktop_config.json` needed.

```json
{
  "mcpServers": {
    "markdown-generator": {
      "command": "/absolute/path/to/mdgen-mcp/bin/mdgen-mcp"
    }
  }
}
```

Replace the path with the absolute path to the binary you'll build in Step 7.

---

## Step 3: `main.go`

```go
package main

import (
	"log"

	"mdgen-mcp/internal/mcpserver"
)

func main() {
	if err := mcpserver.Start(); err != nil {
		log.Fatal(err)
	}
}
```

---

## Step 4: `internal/mcpserver/server.go`

The MCP server exposes two tools: `save_as_markdown` and `list_markdown_files`.

```go
package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"mdgen-mcp/internal/formatter"
	"mdgen-mcp/internal/parser"
	"mdgen-mcp/internal/writer"
)

func Start() error {
	s := server.NewMCPServer(
		"markdown-generator",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Tool 1: save_as_markdown
	saveTool := mcp.NewTool("save_as_markdown",
		mcp.WithDescription("Save conversation or text as a formatted Markdown (.md) file in ./docs/"),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Document title — used as the filename"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The full text or conversation to convert into Markdown"),
		),
		mcp.WithString("filename",
			mcp.Description("Optional custom filename without the .md extension"),
		),
	)
	s.AddTool(saveTool, saveMarkdownHandler)

	// Tool 2: list_markdown_files
	listTool := mcp.NewTool("list_markdown_files",
		mcp.WithDescription("List all Markdown files saved in ./docs/"),
	)
	s.AddTool(listTool, listMarkdownHandler)

	return server.NewStdioServer(s).Listen(context.Background(), nil)
}

func saveMarkdownHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title, _ := req.Params.Arguments["title"].(string)
	content, _ := req.Params.Arguments["content"].(string)
	filename, _ := req.Params.Arguments["filename"].(string)

	if title == "" || content == "" {
		return mcp.NewToolResultError("title and content are required"), nil
	}

	doc := parser.Parse(content, title)
	fmtr := formatter.NewDefault()
	md := fmtr.Format(doc)

	path, err := writer.Write(md, title, "./docs", filename)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to save: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Saved to %s", path)), nil
}

func listMarkdownHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	files, err := writer.ListFiles("./docs")
	if err != nil || len(files) == 0 {
		return mcp.NewToolResultText("No markdown files found in ./docs/"), nil
	}
	return mcp.NewToolResultText(strings.Join(files, "\n")), nil
}
```

---

## Step 5: `internal/parser/parser.go`

Converts raw text into a structured `Document` using a line-by-line state machine.

```go
package parser

import (
	"regexp"
	"strings"
	"unicode"
)

type BlockType int

const (
	BlockHeading BlockType = iota
	BlockParagraph
	BlockBulletList
	BlockNumberedList
	BlockCode
	BlockHRule
)

type Block struct {
	Type    BlockType
	Level   int      // 1, 2, 3 for headings
	Content string   // for heading/paragraph
	Items   []string // for list blocks
	Lang    string   // for code blocks
}

type Document struct {
	Title  string
	Blocks []Block
}

var (
	allCapsRe   = regexp.MustCompile(`^[A-Z][A-Z\s\d\-\_]{3,}:?$`)
	bulletRe    = regexp.MustCompile(`^[-*•]\s+`)
	numberedRe  = regexp.MustCompile(`^\d+\.\s+`)
	hRuleRe     = regexp.MustCompile(`^[-*_]{3,}$`)
	slugRe      = regexp.MustCompile(`[^a-z0-9]+`)
)

// Parse converts raw text into a Document. titleOverride sets the title
// directly; if empty, the first H1 heading is used as the title.
func Parse(text, titleOverride string) Document {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")

	var blocks []Block
	var paraLines []string
	var codeLines []string
	var listItems []string
	var listType BlockType
	var codeLang string

	inCode := false
	inFencedCode := false
	prevBlank := true

	flushParagraph := func() {
		if len(paraLines) == 0 {
			return
		}
		blocks = append(blocks, Block{
			Type:    BlockParagraph,
			Content: strings.Join(paraLines, " "),
		})
		paraLines = nil
	}

	flushList := func() {
		if len(listItems) == 0 {
			return
		}
		blocks = append(blocks, Block{
			Type:  listType,
			Items: listItems,
		})
		listItems = nil
	}

	flushCode := func() {
		if len(codeLines) == 0 {
			return
		}
		blocks = append(blocks, Block{
			Type:    BlockCode,
			Content: strings.Join(codeLines, "\n"),
			Lang:    codeLang,
		})
		codeLines = nil
		codeLang = ""
		inCode = false
	}

	for _, line := range lines {
		// Handle fenced code blocks
		if strings.HasPrefix(line, "```") {
			if inFencedCode {
				flushCode()
				inFencedCode = false
				prevBlank = false
				continue
			}
			flushParagraph()
			flushList()
			codeLang = strings.TrimPrefix(line, "```")
			inFencedCode = true
			inCode = true
			prevBlank = false
			continue
		}

		if inFencedCode {
			codeLines = append(codeLines, line)
			continue
		}

		// Indented code (4 spaces or tab)
		if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
			flushParagraph()
			flushList()
			codeLines = append(codeLines, strings.TrimPrefix(strings.TrimPrefix(line, "\t"), "    "))
			inCode = true
			prevBlank = false
			continue
		}

		// Exit indented code on non-indented, non-blank line
		if inCode && !inFencedCode && strings.TrimSpace(line) != "" {
			flushCode()
		}

		// Blank line
		if strings.TrimSpace(line) == "" {
			flushParagraph()
			flushList()
			prevBlank = true
			continue
		}

		// Horizontal rule
		if hRuleRe.MatchString(strings.TrimSpace(line)) {
			flushParagraph()
			flushList()
			blocks = append(blocks, Block{Type: BlockHRule})
			prevBlank = true
			continue
		}

		// ATX heading (# ## ###)
		if strings.HasPrefix(line, "#") {
			flushParagraph()
			flushList()
			level := 0
			for _, ch := range line {
				if ch == '#' {
					level++
				} else {
					break
				}
			}
			if level > 6 {
				level = 6
			}
			content := strings.TrimSpace(strings.TrimLeft(line, "# "))
			blocks = append(blocks, Block{Type: BlockHeading, Level: level, Content: content})
			prevBlank = false
			continue
		}

		// ALL CAPS → inferred H2
		trimmed := strings.TrimSpace(line)
		if allCapsRe.MatchString(trimmed) && hasLetter(trimmed) {
			flushParagraph()
			flushList()
			content := toTitleCase(strings.TrimSuffix(trimmed, ":"))
			blocks = append(blocks, Block{Type: BlockHeading, Level: 2, Content: content})
			prevBlank = false
			continue
		}

		// Line ending with ":" after a blank → inferred H3
		if prevBlank && strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, " ") == false {
			flushParagraph()
			flushList()
			content := strings.TrimSuffix(trimmed, ":")
			blocks = append(blocks, Block{Type: BlockHeading, Level: 3, Content: content})
			prevBlank = false
			continue
		}

		// Bullet list
		if bulletRe.MatchString(line) {
			flushParagraph()
			if len(listItems) > 0 && listType != BlockBulletList {
				flushList()
			}
			listType = BlockBulletList
			item := bulletRe.ReplaceAllString(line, "")
			listItems = append(listItems, strings.TrimSpace(item))
			prevBlank = false
			continue
		}

		// Numbered list
		if numberedRe.MatchString(line) {
			flushParagraph()
			if len(listItems) > 0 && listType != BlockNumberedList {
				flushList()
			}
			listType = BlockNumberedList
			item := numberedRe.ReplaceAllString(line, "")
			listItems = append(listItems, strings.TrimSpace(item))
			prevBlank = false
			continue
		}

		// Paragraph
		flushList()
		paraLines = append(paraLines, trimmed)
		prevBlank = false
	}

	// Flush remaining accumulators
	flushParagraph()
	flushList()
	flushCode()

	// Resolve title
	title := titleOverride
	if title == "" && len(blocks) > 0 && blocks[0].Type == BlockHeading {
		title = blocks[0].Content
	}

	return Document{Title: title, Blocks: blocks}
}

// Slugify converts a string to a URL-safe slug.
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func toTitleCase(s string) string {
	s = strings.ToLower(s)
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(string(w[0])) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
```

---

## Step 6: `internal/formatter/formatter.go`

Renders a `Document` into a formatted Markdown string.

```go
package formatter

import (
	"fmt"
	"strings"

	"mdgen-mcp/internal/parser"
)

type Formatter interface {
	Format(doc parser.Document) string
}

type defaultFormatter struct{}

func NewDefault() Formatter {
	return &defaultFormatter{}
}

func (f *defaultFormatter) Format(doc parser.Document) string {
	var sb strings.Builder

	for i, block := range doc.Blocks {
		switch block.Type {
		case parser.BlockHeading:
			sb.WriteString(strings.Repeat("#", block.Level))
			sb.WriteString(" ")
			sb.WriteString(block.Content)
			sb.WriteString("\n")

		case parser.BlockParagraph:
			sb.WriteString(wrapText(block.Content, 80))

		case parser.BlockBulletList:
			for _, item := range block.Items {
				sb.WriteString("- ")
				sb.WriteString(item)
				sb.WriteString("\n")
			}

		case parser.BlockNumberedList:
			for j, item := range block.Items {
				sb.WriteString(fmt.Sprintf("%d. %s\n", j+1, item))
			}

		case parser.BlockCode:
			sb.WriteString("```")
			sb.WriteString(block.Lang)
			sb.WriteString("\n")
			sb.WriteString(block.Content)
			sb.WriteString("\n```\n")

		case parser.BlockHRule:
			sb.WriteString("---\n")
		}

		// Add blank line between blocks (except last)
		if i < len(doc.Blocks)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func wrapText(content string, width int) string {
	if width <= 0 || len(content) <= width {
		return content + "\n"
	}
	words := strings.Fields(content)
	var lines []string
	var current strings.Builder
	for _, w := range words {
		if current.Len()+len(w)+1 > width && current.Len() > 0 {
			lines = append(lines, current.String())
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteByte(' ')
		}
		current.WriteString(w)
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return strings.Join(lines, "\n") + "\n"
}
```

---

## Step 7: `internal/writer/writer.go`

Saves the formatted markdown to disk.

```go
package writer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mdgen-mcp/internal/parser"
)

// Write saves markdown content to outputDir/<resolvedName>.md.
// Returns the relative path to the written file.
func Write(content, title, outputDir, filename string) (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("create directory %s: %w", outputDir, err)
	}
	name := resolveFilename(title, filename)
	path := filepath.Join(outputDir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write file %s: %w", path, err)
	}
	return path, nil
}

// ListFiles returns a formatted list of .md files in the given directory.
func ListFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var results []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			info, _ := e.Info()
			results = append(results, fmt.Sprintf("%s (%d bytes)", e.Name(), info.Size()))
		}
	}
	return results, nil
}

func resolveFilename(title, override string) string {
	if override != "" {
		return sanitize(override) + ".md"
	}
	if title != "" {
		slug := parser.Slugify(title)
		if slug != "" {
			return slug + ".md"
		}
	}
	return "doc-" + time.Now().Format("20060102-150405") + ".md"
}

func sanitize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else if r == ' ' {
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
```

---

## Step 8: Build & Install

```bash
# Build the binary
go build -o bin/mdgen-mcp .

# Update .mcp.json with the absolute path to the binary
# Then restart Claude Code
```

Verify in Claude Code by running `/mcp` — you should see `markdown-generator` listed.

---

## Usage in Claude

| You say | Claude calls |
|---|---|
| "Make a markdown of this conversation" | `save_as_markdown(title, content)` |
| "Save this as a doc called API Notes" | `save_as_markdown(title="API Notes", content)` |
| "Save with filename my-notes" | `save_as_markdown(..., filename="my-notes")` |
| "What markdown files have I saved?" | `list_markdown_files()` |

Files are saved to `./docs/` relative to where the binary runs (your project root).

---

## Sample Transformation

**Input (conversation text):**
```
GETTING STARTED

This tool converts your chats into markdown docs automatically.

Installation:
Run the following to install:

    go build -o bin/mdgen-mcp .

Features:
- Saves full conversations
- Auto-formats headings and lists
- Detects code blocks
```

**Output (`./docs/getting-started.md`):**
```markdown
## Getting Started

This tool converts your chats into markdown docs automatically.

### Installation

Run the following to install:

```
go build -o bin/mdgen-mcp .
```

### Features

- Saves full conversations
- Auto-formats headings and lists
- Detects code blocks
```

---

## Dependencies Summary

| Package | Version | Purpose |
|---|---|---|
| `github.com/mark3labs/mcp-go` | v0.49.0 | MCP server + stdio transport |
| Go stdlib only | — | parser, formatter, writer, logger |

Go version required: **1.21+** (for `log/slog`).
