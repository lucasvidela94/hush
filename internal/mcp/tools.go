package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func checkTool() mcp.Tool {
	return mcp.NewTool("hush_check",
		mcp.WithDescription("Report which secret names are present in the hush vault and which are missing. Returns names only, never values."),
		mcp.WithArray("names", mcp.Description("Secret names to check"), mcp.Required()),
	)
}

func listTool() mcp.Tool {
	return mcp.NewTool("hush_list",
		mcp.WithDescription("List secret names stored in the hush vault. Returns names only, never values."),
	)
}

func needTool() mcp.Tool {
	return mcp.NewTool("hush_need",
		mcp.WithDescription("Ensure a secret is stored. If present, reports so without revealing it. If missing, asks the user for the value through the client's native prompt and stores it. Never ask for the value in chat; never pass a pasted value back into any tool."),
		mcp.WithString("name", mcp.Description("Secret name to ensure"), mcp.Required()),
		mcp.WithString("hint", mcp.Description("Where the human finds the value, e.g. Meta dashboard path")),
	)
}

func runTool() mcp.Tool {
	return mcp.NewTool("hush_run",
		mcp.WithDescription("Run a command with named secrets injected into the child process environment. Output is redacted before returning. Injection is deny-by-default (only[] or all=true) and requires human confirmation via prompt; pass confirm=true only when the human already approved and the client cannot prompt."),
		mcp.WithArray("command", mcp.Description("Command and arguments to run"), mcp.Required()),
		mcp.WithArray("only", mcp.Description("Inject only these secret names")),
		mcp.WithBoolean("all", mcp.Description("Inject all secrets. Prefer only[] when possible.")),
		mcp.WithBoolean("confirm", mcp.Description("Human already approved. Only for clients without prompt support.")),
		mcp.WithString("stdin_name", mcp.Description("Write this secret to the command stdin, e.g. wrangler secret put")),
	)
}
