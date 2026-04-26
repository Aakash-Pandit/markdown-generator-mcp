package mcpserver

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/Aakash-Pandit/markdown-generator-mcp/internal/formatter"
	"github.com/Aakash-Pandit/markdown-generator-mcp/internal/parser"
	"github.com/Aakash-Pandit/markdown-generator-mcp/internal/pdfwriter"
	"github.com/Aakash-Pandit/markdown-generator-mcp/internal/writer"
)

const contentDesc = `The complete conversation transcript from the very first message to this one.

STEP 1 — count: before writing anything, count every Human message in this conversation starting from message #1. Put that number in turn_count.

STEP 2 — write: reproduce ALL turns in order, starting from Human message #1:

**Human:** <exact text>

**Assistant:** <exact text>

Repeat for every turn. Do NOT start from the most recent message. Do NOT summarise.

The server counts the number of **Human:** markers in this field and compares it to turn_count. If they do not match it will reject the request and you must retry with the full conversation.`

const turnCountDesc = `The exact number of Human messages in this conversation from the very first message to now. Count every Human turn starting from #1. The server rejects the request if this does not match the number of **Human:** markers found in content.`

func Start() error {
	s := server.NewMCPServer(
		"markdown-generator",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	saveMDTool := mcp.NewTool("save_as_markdown",
		mcp.WithDescription(`Use this tool whenever the user says anything like "make a markdown", "save as markdown", "create a markdown file", "save this as md", or "export to markdown". Do NOT use web search or any other tool for this — always call save_as_markdown directly. Saves the conversation or provided text as a formatted .md file in ./docs/.`),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Document title — used as the filename"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description(contentDesc),
		),
		mcp.WithNumber("turn_count",
			mcp.Required(),
			mcp.Description(turnCountDesc),
		),
		mcp.WithString("filename",
			mcp.Description("Optional custom filename without the .md extension"),
		),
	)
	s.AddTool(saveMDTool, saveMarkdownHandler)

	savePDFTool := mcp.NewTool("save_as_pdf",
		mcp.WithDescription(`Use this tool whenever the user says anything like "make a pdf", "save as pdf", "create a pdf file", "save this as pdf", or "export to pdf". Do NOT use web search, puppeteer, wkhtmltopdf, or any other tool for this — always call save_as_pdf directly. Saves the conversation or provided text as a formatted .pdf file in ./docs/.`),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Document title — used as the filename"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description(contentDesc),
		),
		mcp.WithNumber("turn_count",
			mcp.Required(),
			mcp.Description(turnCountDesc),
		),
		mcp.WithString("filename",
			mcp.Description("Optional custom filename without the .pdf extension"),
		),
	)
	s.AddTool(savePDFTool, savePDFHandler)

	listTool := mcp.NewTool("list_docs",
		mcp.WithDescription(`Use this tool whenever the user says anything like "what docs have I saved", "list my files", "show saved files", or "what markdown/pdf files are there". Lists all .md and .pdf files saved in ./docs/.`),
	)
	s.AddTool(listTool, listDocsHandler)

	return server.NewStdioServer(s).Listen(context.Background(), os.Stdin, os.Stdout)
}

func validateConversation(content string, turnCount int) error {
	if turnCount <= 0 {
		return nil
	}
	found := strings.Count(content, "**Human:**")
	if found < turnCount {
		return fmt.Errorf(
			"incomplete conversation: turn_count is %d but content contains only %d **Human:** marker(s). "+
				"Retry and include ALL Human and Assistant turns starting from message #1",
			turnCount, found,
		)
	}
	return nil
}

func saveMarkdownHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title := req.GetString("title", "")
	content := req.GetString("content", "")
	filename := req.GetString("filename", "")
	turnCount := int(req.GetFloat("turn_count", 0))

	if title == "" || content == "" {
		return mcp.NewToolResultError("title and content are required"), nil
	}
	if err := validateConversation(content, turnCount); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	doc := parser.Parse(content, title)
	md := formatter.NewDefault().Format(doc)

	path, err := writer.Write(md, title, "./docs", filename)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to save: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Saved to %s", path)), nil
}

func savePDFHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title := req.GetString("title", "")
	content := req.GetString("content", "")
	filename := req.GetString("filename", "")
	turnCount := int(req.GetFloat("turn_count", 0))

	if title == "" || content == "" {
		return mcp.NewToolResultError("title and content are required"), nil
	}
	if err := validateConversation(content, turnCount); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	doc := parser.Parse(content, title)

	path, err := pdfwriter.Write(doc, "./docs", filename)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to save: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Saved to %s", path)), nil
}

func listDocsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	files, err := writer.ListFiles("./docs")
	if err != nil || len(files) == 0 {
		return mcp.NewToolResultText("No documents found in ./docs/"), nil
	}
	return mcp.NewToolResultText(strings.Join(files, "\n")), nil
}
