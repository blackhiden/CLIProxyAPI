package gemini

import (
	. "github.com/blackhiden/CLIProxyAPI/v8/internal/constant"
	"github.com/blackhiden/CLIProxyAPI/v8/internal/interfaces"
	"github.com/blackhiden/CLIProxyAPI/v8/internal/translator/translator"
)

func init() {
	translator.Register(
		Gemini,
		OpenAI,
		ConvertGeminiRequestToOpenAI,
		interfaces.TranslateResponse{
			Stream:     ConvertOpenAIResponseToGemini,
			NonStream:  ConvertOpenAIResponseToGeminiNonStream,
			TokenCount: GeminiTokenCount,
		},
	)
}
