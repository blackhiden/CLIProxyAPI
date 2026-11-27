package claude

import (
	. "github.com/blackhiden/CLIProxyAPI/v8/internal/constant"
	"github.com/blackhiden/CLIProxyAPI/v8/internal/interfaces"
	"github.com/blackhiden/CLIProxyAPI/v8/internal/translator/translator"
)

func init() {
	translator.Register(
		Claude,
		Interactions,
		ConvertClaudeRequestToInteractions,
		interfaces.TranslateResponse{
			Stream:    ConvertInteractionsResponseToClaude,
			NonStream: ConvertInteractionsResponseToClaudeNonStream,
		},
	)
}
