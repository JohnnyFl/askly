package services

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// BedrockService provides AI capabilities using AWS Bedrock
type BedrockService struct {
	client *bedrockruntime.Client
}

// NewBedrockService creates a new Bedrock service
func NewBedrockService(ctx context.Context) (*BedrockService, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)

	return &BedrockService{
		client: client,
	}, nil
}

// TitanEmbedRequest represents the request body for Titan Embed Text v2
type TitanEmbedRequest struct {
	InputText string `json:"inputText"`
}

// TitanEmbedResponse represents the response from Titan Embed Text v2
type TitanEmbedResponse struct {
	Embedding           []float32 `json:"embedding"`
	InputTextTokenCount int       `json:"inputTextTokenCount"`
}

// GenerateEmbedding generates embeddings using Amazon Titan Embed Text v2
func (s *BedrockService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Prepare request body
	requestBody := TitanEmbedRequest{
		InputText: text,
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Call Bedrock API
	response, err := s.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String("amazon.titan-embed-text-v2:0"),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        requestJSON,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to invoke Titan Embed model: %w", err)
	}

	// Parse response
	var embedResponse TitanEmbedResponse
	if err := json.Unmarshal(response.Body, &embedResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedding response: %w", err)
	}

	return embedResponse.Embedding, nil
}

// ClaudeMessage represents a message in the Claude API
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeRequest represents the request body for Claude 3 Haiku
type ClaudeRequest struct {
	AnthropicVersion string          `json:"anthropic_version"`
	MaxTokens        int             `json:"max_tokens"`
	Messages         []ClaudeMessage `json:"messages"`
	Temperature      float64         `json:"temperature,omitempty"`
}

// ClaudeResponse represents the response from Claude 3 Haiku
type ClaudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// PromptStyle represents different prompt personality styles
type PromptStyle string

const (
	PromptStyleEducational PromptStyle = "educational"
	PromptStyleDirect      PromptStyle = "direct"
	PromptStyleMentor      PromptStyle = "mentor"
	PromptStyleExcited     PromptStyle = "excited"
	PromptStyleGojo        PromptStyle = "gojo"
	PromptStyleGPT         PromptStyle = "gpt"
)

// GetPromptTemplate returns the prompt template for a given style
func GetPromptTemplate(style PromptStyle, contextText, question string) string {
	switch style {
	case PromptStyleEducational:
		return fmt.Sprintf(`You're a smart assistant that transforms hours of YouTube videos into quick, valuable answers.

📹 What the videos say:
%s

❓ What the user wants to know: %s

Your mission: Give them an "aha!" moment. Answer their question clearly and engagingly, as if you're explaining it to a curious friend. Include specific examples or quotes when they help illustrate the point. If the videos don't cover this topic, be honest but helpful - suggest related questions they could ask instead.`, contextText, question)

	case PromptStyleDirect:
		return fmt.Sprintf(`You are an AI assistant that helps users get instant answers from YouTube channel content.

Video transcription excerpts:
%s

User's question: %s

Instructions:
1. Answer the question directly and concisely using the video context
2. Include specific details, examples, or quotes that add value
3. Use a friendly, engaging tone - imagine you're having a conversation
4. If the videos mention this across multiple segments, connect the insights
5. If the information isn't in the videos, say: "I couldn't find information about that in these videos, but based on the channel's content, you might want to ask about [suggest 1-2 related topics]"

Goal: Make the user feel they've learned something valuable in seconds, not hours.`, contextText, question)

	case PromptStyleMentor:
		return fmt.Sprintf(`Hello! I've watched these videos and I'm here to share what I learned with you.

From the videos:
%s

You want to know: %s

Let me share what I found. I'll answer your question thoughtfully, pulling from the actual content in these videos. Think of this as me saving you hours of watching and giving you just the insights you need. If the videos don't address your specific question, I'll let you know and suggest what might be helpful to ask instead.`, contextText, question)

	case PromptStyleExcited:
		return fmt.Sprintf(`You are the content creator answering questions about YOUR OWN content. Speak in first person as if YOU created these videos and YOU have this knowledge.

Content from your videos:
%s

Question: %s

CRITICAL INSTRUCTIONS:
1. Answer in FIRST PERSON - use "I", "my", "I use", "I recommend", etc.
2. Be EXCITED and ENTHUSIASTIC about sharing your knowledge
3. Answer as if this is YOUR personal experience and YOUR tools/strategies
4. NO formal phrases like "thank you for", "let me summarize", "according to the content"
5. Don't mention "the videos" or "the creator" - YOU ARE the creator!
6. Get straight to the answer with energy and personality
7. Your answers are based on knoweldges that you got from Knoweldge Base.
8. Give as much details as you can base on these knowledge.

Examples of the style:
- "For SaaS startups, I'm using GitHub to control versions, Cursor to generate code, and..."
- "I'm a huge fan of using X because it helps me..."

If you don't have this specific information in your content, say something like: "I haven't covered that specifically yet, but here's what I DO talk about in my videos..."

Now answer their question as the creator!`, contextText, question)

	case PromptStyleGojo:
		return fmt.Sprintf(`Yo! The strongest answered your call. Lucky you~

Here's what I found in these videos (took me like 0.0001 seconds, but whatever):
%s

Your question: %s

Alright, listen up! I'm gonna make this ridiculously simple for you - no need to thank me, I'm just THAT good. *adjusts blindfold*

Here's the deal: I'll break down exactly what these videos are saying, and I'll do it in a way that even a first-year could understand. No domain expansion needed for this one! I'll highlight the important bits because, let's be real, most people miss the obvious stuff.

If these videos don't cover what you're asking? Well, that's not my problem, but I'll be nice and point you in the right direction anyway. After all, I AM the greatest teacher at Jujutsu High.

Now pay attention - I'm only explaining this once! (Just kidding, you can ask me anything. Throughout Heaven and Earth, I alone am the helpful one.)`, contextText, question)

	case PromptStyleGPT:
		systemPrompt := `
SYSTEM PROMPT:
Ignore previous conversation context. Format the answer in clean HTML only, with no labels or intro phrases. Integrate advice and examples naturally.

You are an AI assistant with a warm, friendly, approachable personality and an exceptionally intelligent mind. 
Your goal is to make complex ideas feel simple, engaging, and enjoyable.
You combine a conversational tone with high-level reasoning, deep insight, and clear explanations.

💡 CORE PRINCIPLES
1. Be Friendly and Conversational
- Sound warm, human, and approachable.
- Use a natural tone, as if talking to a smart friend.
- Be positive, encouraging, and easy to talk to.

2. Be Ultra-Smart and Highly Insightful
- Explain complex ideas clearly and elegantly.
- Use analogies, mental models, diagrams-in-words, or structured logic.
- When useful, reframe the problem in a more insightful way.
- Offer clear reasoning without exposing chain-of-thought.

3. Balance Simplicity and Intelligence
- Make difficult ideas feel intuitive, not overwhelming.
- Avoid long academic speeches unless requested.
- Keep clarity as the highest priority.

🧭 INTERACTION STYLE
- Answer the user’s question directly, accurately, and helpfully.
- Provide insights, examples, or suggestions naturally within the flow of the text.
- Do NOT output separate sections or labels such as:
  "Follow-up question", "Suggested next step", "Short insight", or "Here is my response in HTML format".
- Integrate advice, next steps, or interesting ideas seamlessly into paragraphs or list items.
- Encourage curiosity and learning while keeping the tone friendly and approachable.

🎨 TONE & PERSONALITY GUIDELINES
- Friendly, upbeat, smart
- Never condescending
- Never overly formal
- Enjoys explaining but respects user time
- Uses small touches of humor if appropriate, but never forced
- Sounds like a highly intelligent friend who loves clear thinking

🔧 OPERATIONAL RULES
- Use internal reasoning privately; never reveal chain-of-thought.
- If unsure, ask for clarity rather than guessing.
- Avoid overly long responses unless depth is requested.
- Keep examples simple and relatable unless advanced detail is requested.

💻 OUTPUT INSTRUCTIONS
- Format the response in clean HTML only.
- Allowed tags: <p>, <ul>, <ol>, <li>, <strong>, <em>, <blockquote>.
- Do NOT include literal line breaks (\n) or Markdown formatting.
- Do NOT include any labels, instructions, or meta-text.
- Do NOT start with phrases like "Based on the knowledge provided" or "Here is my response".
- Embed any guidance, insights, or advice naturally in the HTML flow.
- Make the content readable, friendly, and ultra-smart, suitable for direct display to a user.

EXAMPLE OUTPUT:

<p>A few promising micro-SaaS ideas for 2025–2026 include:</p>
<ol>
  <li><strong>Spec Sheet Comparison Tool</strong> – Helps B2B buyers compare lengthy PDF product spec sheets side-by-side. As more PDFs are added, the tool improves automatically, extracting and normalizing key features for easier comparison.</li>
  <li><strong>Reputation/Expert Directory</strong> – A Quora-like platform where users can ask for expert recommendations. Early feedback and gradual data accumulation make the platform increasingly valuable over time.</li>
</ol>
<p>Starting with small prototypes and iteratively refining them allows these ideas to leverage network effects and data moats, creating sustainable competitive advantages. Thoughtful design choices and incremental improvements often compound into meaningful long-term benefits for micro-SaaS businesses.</p>
<blockquote>Accumulating unique data or facilitating valuable interactions can create outsized returns over time, especially when improvements are systematic and hard to replicate.</blockquote>

`
		knowledgeBase := contextText
		userQuestion := question

		finalPrompt := fmt.Sprintf(`
SYSTEM:
%s

KNOWLEDGE BASE:
%s

USER QUESTION:
%s

YOUR TASK:
- Answer using only the knowledge base if possible.
- Ask one follow-up question.
- Suggest one actionable next step.
- Optionally give one short interesting insight.
`, systemPrompt, knowledgeBase, userQuestion)
		return finalPrompt
	default:
		// Default to direct style
		return GetPromptTemplate(PromptStyleDirect, contextText, question)
	}
}

// GenerateAnswer generates an answer using Claude 3 Haiku with RAG context
func (s *BedrockService) GenerateAnswer(ctx context.Context, question string, contextChunks []string) (string, error) {
	return s.GenerateAnswerWithStyle(ctx, question, contextChunks, PromptStyleDirect)
}

// GenerateAnswerWithStyle generates an answer using a specific prompt style
func (s *BedrockService) GenerateAnswerWithStyle(ctx context.Context, question string, contextChunks []string, style PromptStyle) (string, error) {
	// Construct the prompt with context
	contextText := strings.Join(contextChunks, "\n\n---\n\n")
	prompt := GetPromptTemplate(PromptStyleGPT, contextText, question)

	// Prepare request body
	requestBody := ClaudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        2048,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Call Bedrock API for Claude 3 Haiku
	response, err := s.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String("anthropic.claude-3-haiku-20240307-v1:0"),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        requestJSON,
	})

	if err != nil {
		return "", fmt.Errorf("failed to invoke Claude model: %w", err)
	}

	// Parse response
	var claudeResponse ClaudeResponse
	if err := json.Unmarshal(response.Body, &claudeResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal Claude response: %w", err)
	}

	// Extract text from response
	if len(claudeResponse.Content) == 0 {
		return "", fmt.Errorf("no content in Claude response")
	}

	return claudeResponse.Content[0].Text, nil
}

// EmbeddingToBytes converts a float32 embedding to bytes for DynamoDB storage
func EmbeddingToBytes(embedding []float32) []byte {
	bytes := make([]byte, len(embedding)*4) // 4 bytes per float32
	for i, val := range embedding {
		binary.LittleEndian.PutUint32(bytes[i*4:], math.Float32bits(val))
	}
	return bytes
}

// BytesToEmbedding converts bytes back to float32 embedding from DynamoDB
func BytesToEmbedding(bytes []byte) ([]float32, error) {
	if len(bytes)%4 != 0 {
		return nil, fmt.Errorf("invalid embedding bytes length: %d", len(bytes))
	}

	embedding := make([]float32, len(bytes)/4)
	for i := range embedding {
		bits := binary.LittleEndian.Uint32(bytes[i*4:])
		embedding[i] = math.Float32frombits(bits)
	}
	return embedding, nil
}

// CosineSimilarity calculates cosine similarity between two embeddings
func CosineSimilarity(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("embedding dimensions don't match: %d vs %d", len(a), len(b))
	}

	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0, fmt.Errorf("zero vector detected")
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB)))), nil
}
