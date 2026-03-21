package indexer

import (
	oaienrich "github.com/sdkim96/indexing/enrich/openai"
)

type OpenAIEnricher struct {
	*oaienrich.OpenAIEnricher
}

func NewOpenAIEnricher(apiKey string) *OpenAIEnricher {
	return &OpenAIEnricher{
		OpenAIEnricher: oaienrich.New(apiKey),
	}
}
