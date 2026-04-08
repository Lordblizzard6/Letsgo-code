package tools

import "github.com/user/go-claude-code/internal/api"

type Tool interface {
	Definition() api.Tool
	Execute(input interface{}) (string, error)
}
