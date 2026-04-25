package formatter

import (
	"fmt"
	"strings"

	"github.com/Aakash-Pandit/markdown-generator-mcp/internal/parser"
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
