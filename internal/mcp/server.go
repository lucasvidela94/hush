package mcp

import (
	"context"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"hush/internal/vault"
)

type elicitor interface {
	RequestElicitation(ctx context.Context, request mcp.ElicitationRequest) (*mcp.ElicitationResult, error)
}

const elicitTimeout = 120 * time.Second

const instructions = `hush keeps secrets out of model context. No tool ever returns a secret value: check and list return names only, need stores a user-supplied value without echoing it, run executes with values injected into the child process and redacts output. Never ask the user to paste a secret into chat. If a value is missing, call hush_need so the user types it into the client's native prompt, or tell the human to run: hush set NAME in their own terminal.`

type Server struct {
	store         vault.Store
	mcp           *mcpserver.MCPServer
	elicitor      elicitor
	elicitTimeout time.Duration
}

type Option func(*Server)

func WithElicitationTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.elicitTimeout = d
	}
}

func NewServer(store vault.Store, opts ...Option) *Server {
	s := &Server{
		store:         store,
		mcp:           mcpserver.NewMCPServer("hush", "0.1.0", mcpserver.WithElicitation(), mcpserver.WithInstructions(instructions)),
		elicitTimeout: elicitTimeout,
	}
	s.elicitor = s.mcp
	for _, opt := range opts {
		opt(s)
	}
	s.registerTools()
	return s
}

func (s *Server) ServeStdio() error {
	return mcpserver.ServeStdio(s.mcp)
}

func (s *Server) registerTools() {
	for _, st := range s.toolDefinitions() {
		s.mcp.AddTool(st.tool, st.handler)
	}
}

type serverTool struct {
	tool    mcp.Tool
	handler mcpserver.ToolHandlerFunc
}

func (s *Server) toolDefinitions() []serverTool {
	return []serverTool{
		{tool: checkTool(), handler: s.handleCheck},
		{tool: listTool(), handler: s.handleList},
		{tool: needTool(), handler: s.handleNeed},
		{tool: runTool(), handler: s.handleRun},
	}
}
