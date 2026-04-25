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
	Level   int
	Content string
	Items   []string
	Lang    string
}

type Document struct {
	Title  string
	Blocks []Block
}

var (
	allCapsRe  = regexp.MustCompile(`^[A-Z][A-Z\s\d\-\_]{3,}:?$`)
	bulletRe   = regexp.MustCompile(`^[-*•]\s+`)
	numberedRe = regexp.MustCompile(`^\d+\.\s+`)
	hRuleRe    = regexp.MustCompile(`^[-*_]{3,}$`)
	slugRe     = regexp.MustCompile(`[^a-z0-9]+`)
)

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

		if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
			flushParagraph()
			flushList()
			codeLines = append(codeLines, strings.TrimPrefix(strings.TrimPrefix(line, "\t"), "    "))
			inCode = true
			prevBlank = false
			continue
		}

		if inCode && !inFencedCode && strings.TrimSpace(line) != "" {
			flushCode()
		}

		if strings.TrimSpace(line) == "" {
			flushParagraph()
			flushList()
			prevBlank = true
			continue
		}

		if hRuleRe.MatchString(strings.TrimSpace(line)) {
			flushParagraph()
			flushList()
			blocks = append(blocks, Block{Type: BlockHRule})
			prevBlank = true
			continue
		}

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

		trimmed := strings.TrimSpace(line)
		if allCapsRe.MatchString(trimmed) && hasLetter(trimmed) {
			flushParagraph()
			flushList()
			content := toTitleCase(strings.TrimSuffix(trimmed, ":"))
			blocks = append(blocks, Block{Type: BlockHeading, Level: 2, Content: content})
			prevBlank = false
			continue
		}

		if prevBlank && strings.HasSuffix(trimmed, ":") && strings.Contains(trimmed, " ") {
			flushParagraph()
			flushList()
			content := strings.TrimSuffix(trimmed, ":")
			blocks = append(blocks, Block{Type: BlockHeading, Level: 3, Content: content})
			prevBlank = false
			continue
		}

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

		flushList()
		paraLines = append(paraLines, trimmed)
		prevBlank = false
	}

	flushParagraph()
	flushList()
	flushCode()

	title := titleOverride
	if title == "" && len(blocks) > 0 && blocks[0].Type == BlockHeading {
		title = blocks[0].Content
	}

	return Document{Title: title, Blocks: blocks}
}

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
