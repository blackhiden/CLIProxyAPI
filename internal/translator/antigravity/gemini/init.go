package gemini

import (
	. "github.com/blackhiden/CLIProxyAPI/v8/internal/constant"
	"github.com/blackhiden/CLIProxyAPI/v8/internal/interfaces"
	"github.com/blackhiden/CLIProxyAPI/v8/internal/translator/translator"
)

func init() {
	translator.Register(
		Gemini,
		Antigravity,
		ConvertGeminiRequestToAntigravity,
		interfaces.TranslateResponse{
			Stream:     ConvertAntigravityResponseToGemini,
			NonStream:  ConvertAntigravityResponseToGeminiNonStream,
			TokenCount: GeminiTokenCount,
		},
	)
}
