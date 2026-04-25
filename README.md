# Markdown Generator MCP

Save any LLM conversation as a clean `.md` file — just by asking Claude Code.

---

## Install

**Requires:** [Claude Code](https://claude.ai/code), macOS or Linux

```sh
curl -sSL https://raw.githubusercontent.com/Aakash-Pandit/markdown-generator-mcp/main/install.sh | sh
```

Restart Claude Code. Done.

---

## How to Use

### Step 1 — Have a conversation

Chat with any LLM — Claude Code, Codex, ChatGPT, anything. When you're done and want to save it:

### Step 2 — Ask Claude Code to save it

Open Claude Code in your project and say:

```
make a markdown
```

Claude will ask you for the conversation content (paste it in), give it a title, and save the file.

Or be more specific:

```
save this conversation as a markdown called api-design-notes
```

```
save this chat as a doc
```

```
make a markdown of everything we discussed about authentication
```

### Step 3 — Find your file

The file is saved to `./docs/` inside whichever project folder you have open in Claude Code.

```
your-project/
└── docs/
    └── api-design-notes.md   ← your saved file
```

---

## All Commands

| What you say | What it does |
|---|---|
| `make a markdown` | Saves with an auto-generated filename |
| `save this as <name>` | Saves as `./docs/<name>.md` |
| `make a markdown called <name>` | Saves as `./docs/<name>.md` |
| `save this conversation as a doc` | Saves with title as filename |
| `what markdown files have I saved?` | Lists all `.md` files in `./docs/` |

---

## Example

You had a conversation with ChatGPT about setting up a database. You want to save it.

1. Copy the conversation text
2. Open Claude Code and say: **"save this as a markdown called db-setup-notes"**
3. Claude asks for the content — paste the conversation
4. File saved to `./docs/db-setup-notes.md`

---

## Where Files Are Saved

Files are always saved to `./docs/` relative to the project folder you have open in Claude Code. The folder is created automatically if it doesn't exist.

---

## Verify It's Running

In Claude Code, run:

```
/mcp
```

You should see `markdown-generator` in the list with status `connected`.

If it's not there, re-run the install command and restart Claude Code.

---

## Uninstall

```sh
curl -sSL https://raw.githubusercontent.com/Aakash-Pandit/markdown-generator-mcp/main/uninstall.sh | sh
```

Restart Claude Code after.
