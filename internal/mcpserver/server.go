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
	s.AddTool(saveMDTool, saveMarkdownHandler)

	savePDFTool := mcp.NewTool("save_as_pdf",
		mcp.WithDescription("Save conversation or text as a formatted PDF file in ./docs/"),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Document title — used as the filename"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The full text or conversation to convert into PDF"),
		),
		mcp.WithString("filename",
			mcp.Description("Optional custom filename without the .pdf extension"),
		),
	)
	s.AddTool(savePDFTool, savePDFHandler)

	listTool := mcp.NewTool("list_docs",
		mcp.WithDescription("List all documents (markdown and PDF) saved in ./docs/"),
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
