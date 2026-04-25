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
			mcp.Description("The full text or conversation to convert into Markdown. If the user does not provide content, use the current conversation."),
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
			mcp.Description("The full text or conversation to convert into PDF. If the user does not provide content, use the current conversation."),
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

func saveMarkdownHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title := req.GetString("title", "")
	content := req.GetString("content", "")
	filename := req.GetString("filename", "")

	if title == "" || content == "" {
		return mcp.NewToolResultError("title and content are required"), nil
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

	if title == "" || content == "" {
		return mcp.NewToolResultError("title and content are required"), nil
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
