package llm

import (
	"context"
	"log/slog"
	"time"

	"github.com/arslanovdi/Gist/core/internal/domain/model"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

const flowTimeout = time.Second * 30

// GetChatGist выполняет запрос к LLM - сценарий getChatGistFlow.
func (s *GenkitService) GetChatGist(ctx context.Context, messages []model.Message) (string, error) {

	log := slog.With("func", "llm.GetChatGist")
	log.Debug("get chat gist start", slog.Any("messages", messages))

	ctxFlow, cancel := context.WithTimeout(ctx, flowTimeout)
	defer cancel()

	resp, err := s.getChatGistFlow.Run(ctxFlow, &chat{messages})
	if err != nil {
		return "", err
	}

	log.Debug("get chat gist success", slog.String("chat gist", resp))

	return resp, nil
}

// defineGetChatGistFlow определяет сценарий для генерации краткого пересказа чата.
func (s *GenkitService) defineGetChatGistFlow() {

	prompt := `Тебе дана история сообщений из чата в Telegram. 
Каждое сообщение содержит текст, время отправки, ID отправителя и служебную информацию (например, было ли оно отредактировано, переслано или является ответом). 
Составь краткий и связный пересказ этой беседы. Сосредоточься на главных событиях, важных решениях, вопросах и ответах. 
Технические детали (например, метки «переслано» или «отредактировано») учитывай только если они влияют на смысл. 
Пересказ должен быть нейтральным, легко читаемым и помогать человеку быстро понять суть обсуждения. 
История сообщений (в хронологическом порядке): {{messages}} 
Приведи только пересказ, без пояснений и дополнительного форматирования.`

	// Старайся уложиться в 150 слов, если беседа не слишком объёмная.
	// Определяем простой запрос(prompt) getChatGistPrompt
	getChatGistPrompt := genkit.DefinePrompt(s.g, "getChatGistPrompt",
		ai.WithPrompt(prompt),                    // запрос
		ai.WithInputType(chat{}),                 // входные данные
		ai.WithOutputFormat(ai.OutputFormatText), // выходные данные	TODO возвращать текст + аудио
	)

	// Определяем сценарий(flow) getChatGistFlow
	s.getChatGistFlow = genkit.DefineFlow(s.g, "getChatGistFlow", func(ctx context.Context, input *chat) (string, error) {
		// выполняем простой запрос
		resp, err := getChatGistPrompt.Execute(ctx,
			ai.WithInput(input), // входные данные
		)

		// TODO генерация аудио

		if err != nil {
			return "", err
		}
		return resp.Text(), nil
	})
}
