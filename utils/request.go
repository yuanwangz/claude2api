// utils/chat_utils.go
package utils

import (
	"claude2api/config"
	"claude2api/logger"
	"fmt"
	"strings"
)

// ChatRequestProcessor handles common chat request processing logic
type ChatRequestProcessor struct {
	Prompt          strings.Builder
	RootPrompt      strings.Builder
	ImgDataList     []string
	LastUserMessage string
}

// NewChatRequestProcessor creates a new processor instance
func NewChatRequestProcessor() *ChatRequestProcessor {
	return &ChatRequestProcessor{
		Prompt:          strings.Builder{},
		RootPrompt:      strings.Builder{},
		ImgDataList:     []string{},
		LastUserMessage: "",
	}
}

// ProcessMessages processes the messages array into a prompt and extracts images
func (p *ChatRequestProcessor) ProcessMessages(messages []map[string]interface{}) {
	if config.ConfigInstance.PromptDisableArtifacts {
		p.Prompt.WriteString("System: Forbidden to use <antArtifac> </antArtifac> to wrap code blocks, use markdown syntax instead, which means wrapping code blocks with ``` ```\n\n")
	}

	for _, msg := range messages {
		role, roleOk := msg["role"].(string)
		if !roleOk {
			continue // Skip invalid format
		}

		content, exists := msg["content"]
		if !exists {
			continue
		}

		rolePrefix := GetRolePrefix(role)

		p.Prompt.WriteString(rolePrefix)

		switch v := content.(type) {
		case string: // If content is directly a string
			p.Prompt.WriteString(v + "\n\n")
			if role == "user" {
				p.LastUserMessage = rolePrefix + v + "\n\n"
			}
		case []interface{}: // If content is an array of []interface{} type
			for _, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if itemType, ok := itemMap["type"].(string); ok {
						if itemType == "text" {
							if text, ok := itemMap["text"].(string); ok {
								p.Prompt.WriteString(text + "\n\n")
								if role == "user" {
									p.LastUserMessage = rolePrefix + text + "\n\n"
								}
							}
						} else if itemType == "image_url" {
							if imageUrl, ok := itemMap["image_url"].(map[string]interface{}); ok {
								if url, ok := imageUrl["url"].(string); ok {
									p.ImgDataList = append(p.ImgDataList, url)
								}
							}
						}
					}
				}
			}
		}
		logger.Debug(fmt.Sprintf("LastUserMessage: %s", p.LastUserMessage))
	}
	p.RootPrompt.WriteString(p.Prompt.String())
	// Debug output
	logger.Debug(fmt.Sprintf("Processed prompt: %s", p.Prompt.String()))
	logger.Debug(fmt.Sprintf("Image data list: %v", p.ImgDataList))
}

// ResetForBigContext resets the prompt for big context usage
func (p *ChatRequestProcessor) ResetForBigContext() {
	// 重置提示词
	p.Prompt.Reset()

	if config.ConfigInstance.PromptDisableArtifacts {
		p.Prompt.WriteString("System: Forbidden to use <antArtifac> </antArtifac> to wrap code blocks, use markdown syntax instead, which means wrapping code blocks with ``` ```\n\n")
	}

	// 添加大型上下文提示词
	p.Prompt.WriteString(config.ConfigInstance.BigContextPrompt + "\n\n")

	// 添加最后一个用户消息
	// if p.LastUserMessage != "" {
	// 	p.Prompt.WriteString(p.LastUserMessage)
	// }
	// logger.Debug(fmt.Sprintf("ResetForBigContext: %s", p.Prompt.String()))
}
