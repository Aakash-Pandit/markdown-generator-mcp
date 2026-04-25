# Markdown Generator MCP

Say **"make a markdown"** in Claude Code to save any conversation as a `.md` file in `./docs/`.

Works with any source — Claude Code chats, Codex sessions, or any text you paste.

## Install

**Requires:** Go 1.21+, [Claude Code](https://claude.ai/code)

```sh
curl -sSL https://raw.githubusercontent.com/Aakash-Pandit/markdown-generator-mcp/main/install.sh | sh
```

Restart Claude Code. Done.

## Usage

Say any of these in Claude Code:

| What you say | What happens |
|---|---|
| "make a markdown" | Saves conversation to `./docs/` |
| "save this chat as api-notes" | Saves as `./docs/api-notes.md` |
| "save this conversation as a doc" | Saves with auto-generated filename |
| "what markdown files have I saved?" | Lists all files in `./docs/` |

## How It Works

The MCP server exposes two tools:

- **`save_as_markdown`** — formats text as clean Markdown and writes it to `./docs/`
- **`list_markdown_files`** — lists all `.md` files saved in `./docs/`

Claude calls these automatically when you ask it to save something.

## Uninstall

```sh
claude mcp remove markdown-generator
```
