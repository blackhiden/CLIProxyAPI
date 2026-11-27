package chat_completions

import (
	. "github.com/blackhiden/CLIProxyAPI/v8/internal/constant"
	"github.com/blackhiden/CLIProxyAPI/v8/internal/interfaces"
	"github.com/blackhiden/CLIProxyAPI/v8/internal/translator/translator"
)

func init() {
	translator.Register(
		OpenAI,
		Gemini,
		ConvertOpenAIRequestToGemini,
		interfaces.TranslateResponse{
			Stream:    ConvertGeminiResponseToOpenAI,
			NonStream: ConvertGeminiResponseToOpenAINonStream,
		},
	)
}
