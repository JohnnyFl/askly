# Askly POC - Fresh Start

## What We're Building

A working proof-of-concept where:
1. Users can register and login
2. Browse 2 YouTube channels
3. Admin uploads video transcriptions
4. Users ask questions and get AI answers based on uploaded transcriptions

## POC Scope

### In Scope
- User registration and authentication (email/password)
- 2 hardcoded YouTube channels
- Admin panel for uploading transcriptions
- RAG-based Q&A system
- Simple, clean UI

### Out of Scope (For Now)
- Social authentication
- Password reset
- Email verification
- Multiple admin users
- Real-time transcription generation
- Video playback
- Advanced analytics

## Architecture Overview

```
┌──────────────┐
│   Frontend   │  Next.js 13 (App Router)
│   (Next.js)  │  - Landing, Auth, Channels, Chat, Admin
└───────┬──────┘
        │ REST API
        ▼
┌──────────────┐
│ API Gateway  │  Routes to specific Lambda functions
└───────┬──────┘
        │
        ├───► Lambda: Auth (register, login)
        ├───► Lambda: Channels (list, get)
        ├───► Lambda: Admin (upload transcriptions)
        └───► Lambda: Chat (ask questions - with in-memory vectors)
                │
                ├───► DynamoDB (all data + embeddings as binary)
                └───► Bedrock (Titan Embed + Claude 3 Haiku)
```

### Data Flow

**Admin uploads transcription:**
1. Admin logs in → Admin panel
2. Selects channel → Uploads .txt file
3. Lambda chunks text → Calls Bedrock Titan Embed
4. Stores chunks + embeddings (binary) in DynamoDB
5. Total cost: ~$0.0001 per video

**User asks question:**
1. User types question → API Gateway → Chat Lambda
2. **Lambda cold start**: Load all vectors from DynamoDB into memory (~1-2s, one-time)
3. **Lambda warm**: Vectors already in memory
4. Generate question embedding (Bedrock Titan)
5. In-memory brute-force cosine similarity search (~100ms)
6. Retrieve top 3-5 chunks
7. Call Bedrock Claude 3 Haiku with context
8. Return answer + sources (~2-3s total)
9. Store message in DynamoDB

## Tech Stack

### Frontend
- Next.js 13 (App Router)
- TypeScript
- Tailwind CSS + shadcn/ui
- React Hook Form + Zod
- Fetch API for backend calls

### Backend
- **Language**: Go 1.23+
- **Architecture**: Multiple Lambda functions (no HTTP framework needed)
- **Deployment**: AWS Lambda (serverless)
- **API Gateway**: AWS API Gateway (REST API)
- **Authentication**: JWT (using golang-jwt/jwt)
- **AWS SDK**: aws-sdk-go-v2

### Lambda Functions
```
1. lambda-auth-register    - User registration
2. lambda-auth-login       - User/admin login
3. lambda-channels-list    - List channels
4. lambda-channels-get     - Get channel details
5. lambda-admin-upload     - Upload transcriptions (chunk + embed)
6. lambda-admin-list       - List transcriptions
7. lambda-chat-ask         - Answer questions (in-memory vector search)
8. lambda-chat-history     - Get chat history
```

### Infrastructure (AWS)
- **Compute**: AWS Lambda functions (8 separate functions)
- **API**: API Gateway (REST API with Lambda integrations)
- **Database**: DynamoDB (all data including embeddings)
- **Vector Search**: In-memory cosine similarity in Lambda
- **Frontend**: S3 (static files) + CloudFront (CDN)
- **IaC**: Pulumi (Go)

### AI Components (AWS Bedrock)
- **Embeddings**: amazon.titan-embed-text-v2:0
  - 1536 dimensions
  - $0.0001 per 1k tokens
- **Chat**: anthropic.claude-3-haiku-20240307-v1:0
  - Input: $0.00025/1k tokens
  - Output: $0.00125/1k tokens
  - 12x cheaper than Claude 3.5 Sonnet
  - Fast and good quality for Q&A
- **Vector Search**: In-memory brute-force cosine similarity (no additional service needed)

## DynamoDB Schema

### Table: Users
**Primary Key**: PK (String), SK (String)
**GSI1**: email-index (email as PK)

```json
{
  "PK": "USER#<uuid>",
  "SK": "PROFILE",
  "email": "user@example.com",
  "password_hash": "bcrypt-hash",
  "created_at": "2025-10-15T10:00:00Z",
  "updated_at": "2025-10-15T10:00:00Z"
}
```

### Table: Admins
**Primary Key**: PK (String), SK (String)
**GSI1**: email-index (email as PK)

```json
{
  "PK": "ADMIN#<uuid>",
  "SK": "PROFILE",
  "email": "admin@example.com",
  "password_hash": "bcrypt-hash",
  "created_at": "2025-10-15T10:00:00Z"
}
```

### Table: Channels
**Primary Key**: PK (String), SK (String)

```json
{
  "PK": "CHANNEL#<uuid>",
  "SK": "METADATA",
  "name": "AI & Technology Channel",
  "description": "Deep dives into AI and tech",
  "thumbnail_url": "https://...",
  "youtube_url": "https://youtube.com/@channel",
  "youtube_channel_id": "UC...",
  "created_at": "2025-10-15T10:00:00Z"
}
```

### Table: Transcriptions
**Primary Key**: PK (String), SK (String)
**GSI1**: channel_id-uploaded_at-index

```json
{
  "PK": "TRANSCRIPTION#<uuid>",
  "SK": "METADATA",
  "channel_id": "CHANNEL#<uuid>",
  "video_title": "Introduction to RAG Systems",
  "video_url": "https://youtube.com/watch?v=...",
  "video_id": "abc123",
  "full_text": "Full transcription text...",
  "uploaded_by": "ADMIN#<uuid>",
  "uploaded_at": "2025-10-15T10:00:00Z"
}
```

### Table: TranscriptionChunks
**Primary Key**: PK (String), SK (String)
**GSI1**: channel_id-created_at-index
**GSI2**: transcription_id-chunk_index-index

```json
{
  "PK": "CHUNK#<uuid>",
  "SK": "METADATA",
  "transcription_id": "TRANSCRIPTION#<uuid>",
  "channel_id": "CHANNEL#<uuid>",
  "chunk_text": "This section discusses...",
  "chunk_index": 0,
  "embedding": "<binary data - 1536 floats as []byte>",
  "created_at": "2025-10-15T10:00:00Z"
}
```

**Note**: Embedding is stored as binary (DynamoDB Binary type) to save space:
- 1536 float32 values = 6,144 bytes per chunk
- 6,000 chunks = ~37 MB total
- Lambda can load all embeddings into memory on cold start

### Table: ChatMessages
**Primary Key**: PK (String), SK (String)
**GSI1**: user_id-created_at-index
**GSI2**: channel_id-created_at-index

```json
{
  "PK": "MESSAGE#<uuid>",
  "SK": "METADATA",
  "user_id": "USER#<uuid>",
  "channel_id": "CHANNEL#<uuid>",
  "question": "What is RAG?",
  "answer": "RAG stands for Retrieval Augmented Generation...",
  "sources": [
    {
      "video_title": "Intro to RAG",
      "video_url": "https://...",
      "chunk_index": 2,
      "relevance_score": 0.89
    }
  ],
  "created_at": "2025-10-15T10:00:00Z"
}
```

### Vector Search Strategy

**In-Memory Search in Lambda:**

The `lambda-chat-ask` function implements vector search:

1. **Cold Start (first invocation)**:
   - Scan DynamoDB TranscriptionChunks table
   - Load all embeddings into memory (~37 MB for 6,000 chunks)
   - Cache in Lambda global scope
   - Duration: 1-2 seconds

2. **Warm Start (subsequent invocations)**:
   - Embeddings already in memory
   - Duration: < 100ms

3. **Search Algorithm**:
   ```go
   // Pseudocode
   func cosineSimilarity(vec1, vec2 []float32) float32 {
       // Dot product / (magnitude1 * magnitude2)
   }

   func searchTopK(queryEmbedding []float32, allChunks []Chunk, k int) []Match {
       scores := []Match{}
       for _, chunk := range allChunks {
           score := cosineSimilarity(queryEmbedding, chunk.Embedding)
           scores = append(scores, Match{chunk, score})
       }
       sort.Desc(scores)
       return scores[:k]
   }
   ```

**Performance**:
- 6,000 vectors × 1536 dimensions: ~100-200ms search time
- Acceptable for POC
- Lambda memory: 512 MB (plenty of room)

## API Endpoints

### Authentication
```
POST /api/auth/register
Body: { email, password }
Response: { token, user: { id, email } }

POST /api/auth/login
Body: { email, password }
Response: { token, user: { id, email } }

POST /api/auth/me
Headers: Authorization: Bearer <token>
Response: { user: { id, email } }
```

### Admin
```
POST /api/admin/login
Body: { email, password }
Response: { token, admin: { id, email } }

POST /api/admin/channels/:channelId/transcriptions
Headers: Authorization: Bearer <admin-token>
Body: FormData with file + metadata
Response: { transcription: {...}, chunksCreated: number }

GET /api/admin/channels/:channelId/transcriptions
Response: [ { id, video_title, video_url, uploaded_at, ... } ]
```

### Channels
```
GET /api/channels
Headers: Authorization: Bearer <token>
Response: [ { id, name, description, thumbnail_url } ]

GET /api/channels/:id
Headers: Authorization: Bearer <token>
Response: { id, name, description, thumbnail_url, youtube_url }
```

### Chat
```
POST /api/channels/:channelId/ask
Headers: Authorization: Bearer <token>
Body: { question: "What is...?" }
Response: {
  messageId: "uuid",
  answer: "Based on the videos...",
  sources: [
    { video_title, video_url, relevance }
  ]
}

GET /api/channels/:channelId/messages
Headers: Authorization: Bearer <token>
Response: [ { id, question, answer, sources, created_at } ]
```

## Go Dependencies

### Required Go Packages

```go
// Each Lambda function has its own go.mod
// Example: backend/lambdas/chat-ask/go.mod

module lambda-chat-ask

go 1.23

require (
    // AWS Lambda
    github.com/aws/aws-lambda-go v1.47.0

    // AWS SDK v2
    github.com/aws/aws-sdk-go-v2 v1.30.0
    github.com/aws/aws-sdk-go-v2/config v1.27.0
    github.com/aws/aws-sdk-go-v2/service/dynamodb v1.34.0
    github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.14.0
    github.com/aws/aws-sdk-go-v2/service/bedrockruntime v1.16.0

    // Authentication (for lambdas that need JWT)
    github.com/golang-jwt/jwt/v5 v5.2.1
    golang.org/x/crypto v0.25.0  // bcrypt for auth lambdas

    // Utilities
    github.com/google/uuid v1.6.0
)
```

**Note**: No HTTP framework (Gin/Echo) needed - Lambda handles HTTP via API Gateway events.

## Project Structure

```
askly/
├── frontend/
│   ├── app/
│   │   ├── page.tsx                     # Landing
│   │   ├── login/page.tsx               # User login
│   │   ├── signup/page.tsx              # User registration
│   │   ├── channels/page.tsx            # Channel list
│   │   ├── channels/[id]/page.tsx       # Chat interface
│   │   └── admin/
│   │       ├── login/page.tsx           # Admin login
│   │       ├── page.tsx                 # Admin dashboard
│   │       └── channels/[id]/page.tsx   # Upload transcriptions
│   ├── components/
│   ├── lib/
│   └── package.json
├── backend/
│   ├── lambdas/
│   │   ├── auth-register/
│   │   │   ├── main.go                  # User registration Lambda
│   │   │   └── go.mod
│   │   ├── auth-login/
│   │   │   ├── main.go                  # User/admin login Lambda
│   │   │   └── go.mod
│   │   ├── channels-list/
│   │   │   ├── main.go                  # List channels Lambda
│   │   │   └── go.mod
│   │   ├── channels-get/
│   │   │   ├── main.go                  # Get channel Lambda
│   │   │   └── go.mod
│   │   ├── admin-upload/
│   │   │   ├── main.go                  # Upload transcription Lambda
│   │   │   └── go.mod
│   │   ├── admin-list/
│   │   │   ├── main.go                  # List transcriptions Lambda
│   │   │   └── go.mod
│   │   ├── chat-ask/
│   │   │   ├── main.go                  # Ask question Lambda (w/ vector search)
│   │   │   └── go.mod
│   │   └── chat-history/
│   │       ├── main.go                  # Get chat history Lambda
│   │       └── go.mod
│   └── shared/
│       ├── models/
│       │   ├── user.go
│       │   ├── channel.go
│       │   ├── transcription.go
│       │   └── message.go
│       ├── repository/
│       │   └── dynamodb_repo.go         # Shared DynamoDB operations
│       ├── services/
│       │   ├── bedrock_service.go       # Bedrock SDK wrapper
│       │   ├── chunking_service.go      # Text chunking logic
│       │   └── vector_search.go         # In-memory cosine similarity
│       └── auth/
│           └── jwt.go                   # JWT utilities
└── infra/
    └── main.go                          # Pulumi (DynamoDB, Lambda, API Gateway)
```

## Implementation Plan

### Phase 0: Infrastructure Setup (1-2 days)

#### Day 1: Pulumi + AWS Setup
- [x] Update infra/main.go for POC infrastructure
- [x] Create 6 DynamoDB tables with on-demand billing:
  - askly-users (PK, SK, GSI: email)
  - askly-admins (PK, SK, GSI: email)
  - askly-channels (PK, SK)
  - askly-transcriptions (PK, SK, GSI: channel_id-uploaded_at)
  - askly-chunks (PK, SK, GSI: channel_id, GSI: transcription_id-chunk_index)
  - askly-messages (PK, SK, GSI: user_id-created_at, GSI: channel_id-created_at)
- [x] Create S3 bucket for frontend static files (website hosting)
- [x] Create IAM roles for Lambda functions (DynamoDB + Bedrock permissions)
- [ ] Set up API Gateway (REST API) with 8 routes (will add with Lambda functions in Phase 5)
- [ ] Deploy infrastructure with `pulumi up` (requires user to set PULUMI_CONFIG_PASSPHRASE)
- [ ] Note down all resource ARNs/endpoints after deployment
- [ ] Seed 2 channels and 1 admin user in DynamoDB

### Phase 1: Backend Foundation (2-3 days)

#### Day 2: Shared Code Setup
- [ ] Create backend/shared/ directory
- [ ] Implement shared models (User, Channel, Transcription, Message)
- [ ] Implement DynamoDB repository layer (CRUD operations)
- [ ] Implement JWT utilities (generate, verify)
- [ ] Implement Bedrock service wrapper (embed, chat)
- [ ] Implement text chunking service (600 words/chunk)
- [ ] Implement vector search service (cosine similarity)
- [ ] Test locally with AWS credentials

#### Day 3: Auth Lambdas
- [ ] Create lambda-auth-register
  - Accept email/password
  - Hash password with bcrypt
  - Store in DynamoDB Users table
  - Generate JWT token
  - Return token + user info
- [ ] Create lambda-auth-login
  - Support both user and admin login
  - Verify password
  - Generate JWT token
  - Return token + user/admin info
- [ ] Test auth flow with AWS SAM Local or direct invocation

#### Day 4: Channel & Admin List Lambdas
- [ ] Create lambda-channels-list
  - Query DynamoDB Channels table
  - Return array of channels
- [ ] Create lambda-channels-get
  - Get single channel by ID
  - Return channel details
- [ ] Create lambda-admin-list
  - List transcriptions for a channel
  - Return array from Transcriptions table
- [ ] Test with sample data

### Phase 2: RAG System (2-3 days)

#### Day 5: Admin Upload Lambda
- [ ] Create lambda-admin-upload
  - Parse multipart form data from API Gateway
  - Extract video title, URL, transcription text
  - Chunk transcription (600 words per chunk)
  - Call Bedrock Titan Embed for each chunk
  - Convert embeddings to binary ([]byte)
  - Store chunks + embeddings in DynamoDB TranscriptionChunks table
  - Store transcription metadata in Transcriptions table
  - Return success + chunk count
- [ ] Test upload with sample video transcription

#### Day 6: Chat Lambda with Vector Search
- [ ] Create lambda-chat-ask
  - **Cold start logic**: Scan all chunks from DynamoDB, load into memory
  - Parse question from request
  - Call Bedrock Titan Embed for question
  - Run in-memory cosine similarity search (top 5 matches)
  - Construct RAG prompt with context chunks
  - Call Bedrock Claude 3 Haiku
  - Parse answer
  - Store question + answer in ChatMessages table
  - Return answer + sources
- [ ] Create lambda-chat-history
  - Query ChatMessages by channel_id
  - Return chat history
- [ ] Test with sample questions

#### Day 7: Testing & Refinement
- [ ] Upload 3-5 sample transcriptions for each channel
- [ ] Test various question types (factual, comparison, summary)
- [ ] Measure Lambda performance (cold/warm start times)
- [ ] Tune chunking (size, overlap if needed)
- [ ] Tune RAG prompt template
- [ ] Add error handling and logging
- [ ] Optimize Lambda memory settings (512 MB recommended)

### Phase 3: Admin Frontend (1-2 days)

#### Day 7: Admin Panel
- [ ] Create /admin/login page
- [ ] Create /admin dashboard page
- [ ] Create /admin/channels/[id] upload page
- [ ] File upload UI
- [ ] Display uploaded transcriptions list
- [ ] Connect to backend APIs
- [ ] Test admin flow end-to-end

### Phase 4: User Frontend (1-2 days)

#### Day 8-9: Chat Interface
- [ ] Update /channels page to fetch from API
- [ ] Update /channels/[id] chat page
- [ ] Connect question submission to API
- [ ] Display AI responses
- [ ] Show source videos
- [ ] Add loading states
- [ ] Add error handling
- [ ] Test full user flow

### Phase 5: Deployment (1-2 days)

#### Day 10-11: Deploy to AWS

**Backend Deployment:**
- [ ] Update Pulumi to deploy all 8 Lambda functions
  - Build Go binaries for each lambda
  - Zip handlers with dependencies
  - Create Lambda resources in Pulumi
- [ ] Configure Lambda environment variables (DynamoDB tables, Bedrock models, JWT secret)
- [ ] Set up API Gateway with 8 routes:
  - POST /api/auth/register → lambda-auth-register
  - POST /api/auth/login → lambda-auth-login
  - GET /api/channels → lambda-channels-list
  - GET /api/channels/{id} → lambda-channels-get
  - POST /api/admin/channels/{id}/transcriptions → lambda-admin-upload
  - GET /api/admin/channels/{id}/transcriptions → lambda-admin-list
  - POST /api/channels/{id}/ask → lambda-chat-ask
  - GET /api/channels/{id}/messages → lambda-chat-history
- [ ] Deploy infrastructure: `pulumi up`
- [ ] Test each Lambda endpoint with curl/Postman

**Frontend Deployment:**
- [ ] Configure Next.js for static export
  - Update next.config.js: `output: 'export'`
  - Set NEXT_PUBLIC_API_URL to API Gateway URL
- [ ] Build static frontend: `npm run build`
- [ ] Deploy to S3:
  - Upload `out/` directory to S3 bucket
  - Set proper content types for HTML, CSS, JS
  - Configure index.html as default document
- [ ] Invalidate CloudFront cache: `aws cloudfront create-invalidation`
- [ ] Test frontend via CloudFront URL
- [ ] Test full deployed flow end-to-end
- [ ] Monitor Lambda cold start times (optimize if needed)
- [ ] Set up CloudWatch dashboards for monitoring

## User Flows

### User Registration & Chat Flow
1. User lands on homepage
2. Clicks "Sign Up" → Fills form → Gets JWT token
3. Redirected to /channels
4. Sees 2 channels with thumbnails
5. Clicks channel A → Redirected to /channels/A
6. Types question: "What tools are recommended?"
7. AI responds with answer + sources
8. User can ask follow-up questions

### Admin Upload Flow
1. Admin navigates to /admin/login
2. Logs in with admin credentials
3. Sees dashboard with 2 channels
4. Clicks "Upload Transcription" for Channel A
5. Fills form:
   - Video title: "Marketing Tools 101"
   - Video URL: youtube.com/watch?v=abc
   - Upload .txt file with transcription
6. Clicks "Upload"
7. Backend chunks, embeds, stores
8. Admin sees success message
9. Transcription appears in list

## Environment Variables

### Backend Environment Variables (Lambda)

Each Lambda function gets these environment variables via Pulumi:

```env
AWS_REGION=us-east-1

# DynamoDB Tables
DYNAMODB_USERS_TABLE=askly-users
DYNAMODB_ADMINS_TABLE=askly-admins
DYNAMODB_CHANNELS_TABLE=askly-channels
DYNAMODB_TRANSCRIPTIONS_TABLE=askly-transcriptions
DYNAMODB_CHUNKS_TABLE=askly-chunks
DYNAMODB_MESSAGES_TABLE=askly-messages

# Bedrock Models
BEDROCK_EMBEDDING_MODEL=amazon.titan-embed-text-v2:0
BEDROCK_CHAT_MODEL=anthropic.claude-3-haiku-20240307-v1:0

# Auth
JWT_SECRET=<fetch-from-secrets-manager>

# Lambda Config
LAMBDA_MEMORY_MB=512
LAMBDA_TIMEOUT_SEC=30
```

### Local Development .env
For testing Lambdas locally with AWS SAM:

```env
AWS_REGION=us-east-1
AWS_PROFILE=default

# Same as above for Lambda
DYNAMODB_USERS_TABLE=askly-users
...
JWT_SECRET=local-dev-secret-change-in-prod
```

**Note**: Store JWT_SECRET in AWS Secrets Manager for production

### Frontend .env.local
```env
# Local development
NEXT_PUBLIC_API_URL=http://localhost:8080

# Production (after deployment)
# NEXT_PUBLIC_API_URL=https://xxx.execute-api.us-east-1.amazonaws.com/prod
```

### Frontend Configuration (next.config.js)

For static export to S3 + CloudFront:

```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',  // Enable static HTML export
  images: {
    unoptimized: true,  // Disable image optimization for static export
  },
  trailingSlash: true,  // Add trailing slashes to URLs
}

module.exports = nextConfig
```

**Build and Deploy:**
```bash
# Build static export
npm run build
# Output: out/ directory with static HTML, CSS, JS

# Deploy to S3
aws s3 sync out/ s3://askly-frontend --delete

# Invalidate CloudFront cache
aws cloudfront create-invalidation \
  --distribution-id XXXXX \
  --paths "/*"
```

## Success Criteria

POC is complete when:
- [ ] Users can register and login
- [ ] Users see 2 channels
- [ ] Users can ask questions on any channel
- [ ] AI returns accurate, relevant answers
- [ ] Answers include source video links
- [ ] Admin can login
- [ ] Admin can upload transcriptions for each channel
- [ ] System chunks and embeds transcriptions automatically
- [ ] Deployed and accessible online

## Estimated Timeline

**Total: 11 days**

- Phase 0 (Infrastructure): 1-2 days
- Phase 1 (Backend Foundation): 2-3 days
- Phase 2 (RAG System): 2-3 days
- Phase 3 (Admin UI): 1-2 days
- Phase 4 (User UI): 1-2 days
- Phase 5 (Deployment): 1-2 days

## Cost Estimate (AWS) - Ultra-Cheap Configuration

### POC Development Costs (~2 weeks, 400 videos)

#### One-Time Costs (Embedding Generation)
```
400 videos × 9,000 words = 3.6M words ≈ 4.8M tokens
Bedrock Titan Embed: $0.0001 per 1k tokens
Cost: $0.48
```

#### Ongoing Costs (Queries & Storage)
- **DynamoDB**:
  - Storage: ~100 MB (with embeddings) - **$0.025/month**
  - Reads/Writes: On-demand, ~1,000 queries - **$1-2**
- **Lambda**:
  - 8 functions × 1,000 invocations each = **$0** (free tier covers 1M/month)
  - Duration charges: ~200ms avg × 8k invocations = **$0.20**
- **API Gateway**:
  - 10,000 API calls = **$0** (free tier covers 1M/month)
- **S3** (Frontend):
  - Storage: ~50 MB static files - **$0.001**
  - Requests: ~1,000 GET requests - **$0.00**
- **CloudFront**:
  - Data transfer: 10 GB free tier - **$0**
  - Requests: 10M requests free tier - **$0**
- **Bedrock Claude 3 Haiku** (1,000 queries):
  - Input: 1,000 queries × 3k tokens × $0.00025 = **$0.75**
  - Output: 1,000 queries × 500 tokens × $0.00125 = **$0.625**
  - Total: **$1.38**
- **CloudWatch Logs**: **$0.50**

**Total POC Cost: ~$5-7 for 2 weeks** 🎉

### Cost Breakdown by Component
| Service | POC Cost | Production (10k users) |
|---------|----------|------------------------|
| DynamoDB | $2 | $10-20 |
| Lambda | $0.20 | $5-10 |
| API Gateway | $0 | $3-5 |
| S3 + CloudFront | $0 | $5-10 |
| Bedrock (Haiku) | $1.38 | $50-100 |
| Embeddings | $0.48 (one-time) | - |
| **Total** | **$5-7** | **$75-145/month** |

### Cost Optimization Strategy
1. **Use Claude 3 Haiku** instead of Sonnet (12x cheaper)
2. **DynamoDB on-demand billing** (no idle costs)
3. **Lambda for compute** (no idle costs, free tier)
4. **In-memory vector search** (no OpenSearch $700/month cost!)
5. **S3 + CloudFront** for frontend (free tier for POC)
6. **Static export** for Next.js (no SSR compute costs)
7. **Keep Lambda warm** with provisioned concurrency (optional, $5/month)
8. **Monitor with free tier CloudWatch** (1M metrics free)

### Why This is 20x Cheaper
- **No OpenSearch**: Saves $700/month minimum
- **No RDS**: Saves $15-30/month
- **S3 + CloudFront**: Free tier covers POC vs Vercel Pro $20/month
- **Static export**: No SSR compute costs
- **Bedrock Haiku**: 12x cheaper than Sonnet
- **Lambda free tier**: Covers POC completely
- **DynamoDB on-demand**: Pay only for what you use

## Next Steps

1. Review this plan
2. Set up development environment
3. Start with Phase 1 (backend foundation)
4. Work through phases sequentially
5. Test thoroughly at each phase
6. Deploy and demo

## Key Decisions Made

### Technology Choices
- **Go backend**: Type safety, AWS Lambda native support, fast compilation
- **Multiple Lambda functions**: Each endpoint is a separate Lambda (no HTTP framework)
- **AWS-native infrastructure**: DynamoDB + Lambda + Bedrock only
- **DynamoDB**: Stores all data including embeddings as binary
- **In-memory vector search**: Brute-force cosine similarity in Lambda (no OpenSearch!)
- **API Gateway + Lambda**: Serverless, pay-per-request, no idle costs
- **Pulumi**: Infrastructure as Code in Go

### AI Provider
- **AWS Bedrock Claude 3 Haiku**: 12x cheaper than Sonnet, fast, good quality
- **AWS Bedrock Titan Embed v2**: 1536 dimensions, $0.0001/1k tokens
- **Why Bedrock over OpenAI**: No egress costs, AWS-native, similar pricing

### Architecture Decisions
- **No OpenSearch**: In-memory search saves $700/month, works for 6k vectors
- **No RDS/PostgreSQL**: DynamoDB handles everything
- **No Kubernetes**: Lambda provides serverless compute
- **No HTTP framework**: Lambda handles API Gateway events natively
- **S3 + CloudFront**: Static frontend hosting (no SSR needed)
- **Next.js static export**: All pages are client-side (call backend API)
- **Simple JWT auth**: No Cognito for POC (can add later)
- **2 channels, 400 videos**: Perfect scale for POC
- **Manual upload via admin panel**: No automated pipeline yet

### Cost Optimization
- **DynamoDB on-demand**: Pay per request ($2/month for POC)
- **Lambda free tier**: Covers 1M requests/month
- **Haiku over Sonnet**: 12x cheaper ($1.38 vs $20 for 1k queries)
- **No vector database**: Saves $700/month minimum
- **Total: $5-7 for POC** vs $60-130 with OpenSearch

### Deployment Strategy
- **Backend**: Multiple Lambda functions via Pulumi
- **Frontend**: Next.js static export → S3 + CloudFront
- **Infrastructure**: Single `pulumi up` deploys all AWS resources
- **Frontend Deploy**: `npm run build` + `aws s3 sync` + CloudFront invalidation

---

**Status**: Ready to implement
**Last Updated**: 2025-10-15
