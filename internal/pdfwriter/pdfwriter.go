package pdfwriter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"

	"github.com/Aakash-Pandit/markdown-generator-mcp/internal/parser"
)

func Write(doc parser.Document, outputDir, filename string) (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("create directory %s: %w", outputDir, err)
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.AddPage()

	pageW, _ := pdf.GetPageSize()
	textW := pageW - 40

	for _, block := range doc.Blocks {
		switch block.Type {

		case parser.BlockHeading:
			size := map[int]float64{1: 22, 2: 17, 3: 14, 4: 12, 5: 11, 6: 10}[block.Level]
			if size == 0 {
				size = 10
			}
			pdf.SetFont("Helvetica", "B", size)
			pdf.SetTextColor(20, 20, 20)
			pdf.Ln(4)
			pdf.MultiCell(textW, size*0.5, block.Content, "", "L", false)
			pdf.Ln(2)

		case parser.BlockParagraph:
			pdf.SetFont("Helvetica", "", 10)
			pdf.SetTextColor(40, 40, 40)
			pdf.Ln(2)
			pdf.MultiCell(textW, 5, block.Content, "", "L", false)
			pdf.Ln(1)

		case parser.BlockBulletList:
			pdf.SetFont("Helvetica", "", 10)
			pdf.SetTextColor(40, 40, 40)
			pdf.Ln(2)
			for _, item := range block.Items {
				pdf.CellFormat(6, 5, "\xe2\x80\xa2", "", 0, "L", false, 0, "")
				pdf.MultiCell(textW-6, 5, item, "", "L", false)
			}
			pdf.Ln(1)

		case parser.BlockNumberedList:
			pdf.SetFont("Helvetica", "", 10)
			pdf.SetTextColor(40, 40, 40)
			pdf.Ln(2)
			for i, item := range block.Items {
				pdf.CellFormat(8, 5, fmt.Sprintf("%d.", i+1), "", 0, "L", false, 0, "")
				pdf.MultiCell(textW-8, 5, item, "", "L", false)
			}
			pdf.Ln(1)

		case parser.BlockCode:
			pdf.SetFont("Courier", "", 9)
			pdf.SetTextColor(30, 30, 30)
			pdf.SetFillColor(240, 240, 240)
			pdf.Ln(2)
			lines := strings.Split(block.Content, "\n")
			for _, line := range lines {
				pdf.CellFormat(textW, 5, line, "", 1, "L", true, 0, "")
			}
			pdf.SetFillColor(255, 255, 255)
			pdf.Ln(2)

		case parser.BlockHRule:
			pdf.Ln(2)
			pdf.SetDrawColor(180, 180, 180)
			x, y := pdf.GetXY()
			pdf.Line(x, y, x+textW, y)
			pdf.Ln(4)
		}
	}

	name := resolveFilename(doc.Title, filename)
	path := filepath.Join(outputDir, name)
	if err := pdf.OutputFileAndClose(path); err != nil {
		return "", fmt.Errorf("write pdf %s: %w", path, err)
	}
	return path, nil
}

func resolveFilename(title, override string) string {
	if override != "" {
		return sanitize(override) + ".pdf"
	}
	if title != "" {
		slug := parser.Slugify(title)
		if slug != "" {
			return slug + ".pdf"
		}
	}
	return "doc-" + time.Now().Format("20060102-150405") + ".pdf"
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
