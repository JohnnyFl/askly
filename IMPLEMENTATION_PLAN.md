# Askly Implementation Plan

## Project Overview
Askly is a knowledge extraction platform that allows users to ask questions about YouTube channel content and receive AI-powered answers. The system uses a knowledge base built from channel data to provide instant insights without watching hours of video.

## Technology Stack

### Core Technologies
- **Backend**: Go (Golang)
- **Infrastructure**: Pulumi (Go SDK)
- **Database**: AWS DynamoDB
- **Frontend**: Next.js 13 (TypeScript, React)
- **Cloud Provider**: AWS

### AWS Services Required
1. **Amazon Bedrock** - AI/LLM service for question answering (Claude 3.5 Sonnet or similar)
2. **Amazon Bedrock Knowledge Bases** - RAG (Retrieval Augmented Generation) implementation
3. **Amazon OpenSearch Serverless** - Vector database for knowledge base
4. **Amazon S3** - Store raw data (transcripts, metadata, embeddings)
5. **AWS Cognito** - User authentication and authorization
6. **Amazon DynamoDB** - NoSQL database for application data
7. **AWS Lambda** - Serverless functions for data processing
8. **Amazon API Gateway** - REST API endpoint management
9. **Amazon EventBridge** - Event-driven architecture for processing pipelines
10. **AWS Step Functions** - Orchestrate data ingestion workflows
11. **Amazon CloudWatch** - Logging and monitoring

## System Architecture

### High-Level Flow
```
User Upload → S3 → Lambda (Process) → Knowledge Base → Bedrock
                                    ↓
User Question → API Gateway → Lambda → DynamoDB (history) + Bedrock KB → AI Answer
```

### Detailed Component Architecture

#### 1. Data Ingestion Pipeline
```
Raw Data Upload (S3)
    ↓
EventBridge Trigger
    ↓
Step Functions Orchestration
    ↓
Lambda Functions:
    - Parse/Extract content
    - Chunk text (optimal size for embeddings)
    - Generate metadata
    ↓
S3 (Processed chunks + metadata)
    ↓
Bedrock Knowledge Base (Auto-sync)
    ↓
OpenSearch Serverless (Vector embeddings)
```

#### 2. Question Answering Flow
```
User asks question
    ↓
API Gateway (authenticated via Cognito)
    ↓
Lambda Function:
    - Validate request
    - Query Bedrock Knowledge Base with RAG
    - Store Q&A in DynamoDB
    ↓
Return AI-generated answer to frontend
```

## Database Schema (DynamoDB)

### Table 1: Users
**Primary Key**: `user_id` (String)
```
{
  "user_id": "uuid",
  "email": "string",
  "full_name": "string",
  "password_hash": "string",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "cognito_sub": "string"
}
```

### Table 2: Channels
**Primary Key**: `channel_id` (String)
**GSI**: `status-created_at-index` for querying active channels
```
{
  "channel_id": "uuid",
  "name": "string",
  "description": "string",
  "avatar_url": "string",
  "youtube_url": "string",
  "youtube_channel_id": "string",
  "knowledge_base_id": "string",  // Bedrock KB ID
  "data_source_id": "string",     // Bedrock data source ID
  "s3_bucket": "string",
  "s3_prefix": "string",
  "status": "string",  // active, processing, error
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "metadata": {
    "video_count": "number",
    "total_duration": "number",
    "last_sync": "timestamp"
  }
}
```

### Table 3: Questions
**Primary Key**: `question_id` (String)
**GSI 1**: `channel_id-created_at-index` (for channel questions)
**GSI 2**: `user_id-created_at-index` (for user history)
**GSI 3**: `channel_id-is_faq-faq_order-index` (for FAQs)
```
{
  "question_id": "uuid",
  "channel_id": "string",
  "user_id": "string",
  "question_text": "string",
  "answer_text": "string",
  "created_at": "timestamp",
  "is_faq": "boolean",
  "faq_order": "number",
  "metadata": {
    "response_time_ms": "number",
    "confidence_score": "number",
    "sources": [
      {
        "s3_uri": "string",
        "chunk_id": "string",
        "relevance_score": "number"
      }
    ]
  }
}
```

## Data Upload Format & Processing

### Supported Input Formats

#### Option 1: YouTube Transcript JSON (Recommended Initial)
```json
{
  "channel_id": "youtube_channel_id",
  "channel_name": "Channel Name",
  "videos": [
    {
      "video_id": "youtube_video_id",
      "title": "Video Title",
      "description": "Video description",
      "published_at": "timestamp",
      "duration": "seconds",
      "transcript": "Full transcript text...",
      "metadata": {
        "views": "number",
        "likes": "number",
        "tags": ["tag1", "tag2"]
      }
    }
  ]
}
```

#### Option 2: Raw Transcript Files
- Upload individual `.txt` or `.srt` files per video
- Metadata in separate JSON file
- Organized in S3: `channels/{channel_id}/videos/{video_id}/transcript.txt`

#### Option 3: Structured Knowledge Documents
```json
{
  "channel_id": "uuid",
  "documents": [
    {
      "id": "doc_id",
      "title": "Topic or Video Title",
      "content": "Content text for embedding",
      "metadata": {
        "video_url": "string",
        "timestamp": "string",
        "category": "string"
      }
    }
  ]
}
```

### Text Chunking Strategy
- **Chunk Size**: 500-1000 tokens (optimal for Claude embeddings)
- **Overlap**: 100-200 tokens between chunks
- **Preservation**: Keep semantic boundaries (paragraphs, sentences)
- **Metadata**: Attach video title, timestamp, channel info to each chunk

## Backend Implementation (Go)

### Project Structure
```
backend/
├── cmd/
│   └── api/
│       └── main.go                 # API server entry point
├── internal/
│   ├── api/
│   │   ├── handlers/               # HTTP handlers
│   │   │   ├── auth.go
│   │   │   ├── channels.go
│   │   │   └── questions.go
│   │   ├── middleware/             # Auth, CORS, logging
│   │   └── routes.go
│   ├── domain/
│   │   ├── user.go                 # Domain models
│   │   ├── channel.go
│   │   └── question.go
│   ├── repository/
│   │   ├── dynamodb/               # DynamoDB implementations
│   │   │   ├── user_repo.go
│   │   │   ├── channel_repo.go
│   │   │   └── question_repo.go
│   │   └── interfaces.go
│   ├── service/
│   │   ├── auth_service.go         # Business logic
│   │   ├── channel_service.go
│   │   ├── question_service.go
│   │   └── bedrock_service.go      # AI interactions
│   └── config/
│       └── config.go               # Configuration
├── lambda/
│   ├── ingest/                     # Data ingestion Lambda
│   │   └── main.go
│   ├── process/                    # Data processing Lambda
│   │   └── main.go
│   └── sync/                       # KB sync Lambda
│       └── main.go
├── pkg/
│   ├── bedrock/                    # Bedrock client wrapper
│   ├── auth/                       # JWT, Cognito helpers
│   └── utils/                      # Shared utilities
└── go.mod
```

### Key Go Packages Needed
```go
// AWS SDK
github.com/aws/aws-sdk-go-v2/service/dynamodb
github.com/aws/aws-sdk-go-v2/service/bedrockagent
github.com/aws/aws-sdk-go-v2/service/bedrockruntime
github.com/aws/aws-sdk-go-v2/service/s3
github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider

// Lambda
github.com/aws/aws-lambda-go/lambda

// HTTP & API
github.com/gin-gonic/gin  // or chi, fiber
github.com/go-chi/cors

// Auth
github.com/golang-jwt/jwt/v5
golang.org/x/crypto/bcrypt

// Utilities
github.com/google/uuid
github.com/joho/godotenv
```

## Infrastructure as Code (Pulumi)

### Pulumi Stack Structure
```
infra/
├── main.go                         # Main entry point
├── stacks/
│   ├── network.go                  # VPC (if needed)
│   ├── storage.go                  # S3 buckets
│   ├── database.go                 # DynamoDB tables
│   ├── auth.go                     # Cognito user pool
│   ├── api.go                      # API Gateway + Lambda
│   ├── ai.go                       # Bedrock Knowledge Base + OpenSearch
│   ├── processing.go               # Step Functions + EventBridge
│   └── monitoring.go               # CloudWatch
├── config/
│   ├── dev.yaml
│   ├── staging.yaml
│   └── prod.yaml
└── Pulumi.yaml
```

### Resources to Provision

#### 1. S3 Buckets
- `askly-{env}-raw-data` - Uploaded raw data
- `askly-{env}-processed-data` - Processed chunks for KB
- `askly-{env}-assets` - User avatars, channel images

#### 2. DynamoDB Tables
- `askly-{env}-users` (with email GSI)
- `askly-{env}-channels` (with status-created_at GSI)
- `askly-{env}-questions` (with 3 GSIs as defined above)

#### 3. Cognito
- User Pool with email verification
- User Pool Client for frontend
- Identity Pool for AWS credentials (if needed)

#### 4. Bedrock Knowledge Base
- One KB per channel (or shared KB with metadata filtering)
- OpenSearch Serverless collection (vector engine)
- Data sources pointing to S3 processed buckets

#### 5. Lambda Functions
- `ingest-handler` - Triggered by S3 upload
- `process-handler` - Process and chunk data
- `sync-kb-handler` - Sync to Bedrock KB
- `api-handler` - Main API (or API Gateway direct integration)

#### 6. API Gateway
- REST API with Cognito authorizer
- CORS configuration for frontend
- Rate limiting and throttling

#### 7. Step Functions
- Orchestrate: Upload → Process → Chunk → Embed → Sync workflow

#### 8. EventBridge
- S3 upload events → trigger Step Functions

## Implementation Phases

### Phase 1: Infrastructure & Authentication (Week 1-2)
**Goal**: Set up core infrastructure and user authentication

**Tasks**:
1. ✅ Set up Pulumi project structure
2. Provision S3 buckets for data storage
3. Create DynamoDB tables (Users, Channels, Questions)
4. Set up Cognito User Pool and configure
5. Implement user registration/login API (Go backend)
6. Integrate JWT authentication
7. Connect frontend to real authentication endpoints
8. Test user flow end-to-end

**Deliverables**:
- Working user signup/login
- DynamoDB tables created
- S3 buckets ready
- API Gateway with authentication

### Phase 2: Channel Management & Data Ingestion (Week 3-4)
**Goal**: Enable channel creation and data upload

**Tasks**:
1. Implement Channels CRUD API (Go)
2. Build data upload endpoint (accept JSON format)
3. Create S3 upload Lambda trigger
4. Implement text chunking logic
5. Store processed chunks in S3
6. Update frontend to fetch real channel data
7. Build admin interface for channel/data upload (optional MVP)

**Deliverables**:
- Channels can be created and listed
- Data upload pipeline functional
- Chunks stored in S3

### Phase 3: Bedrock Knowledge Base Integration (Week 5-6)
**Goal**: Set up AI question answering

**Tasks**:
1. Provision OpenSearch Serverless collection
2. Create Bedrock Knowledge Base per channel
3. Configure data sources (S3 → KB)
4. Implement KB sync Lambda function
5. Create Bedrock service layer in Go
6. Implement question answering API endpoint
7. Test RAG retrieval with sample questions
8. Store Q&A history in DynamoDB

**Deliverables**:
- Working knowledge base with embeddings
- Question answering returns AI responses
- Q&A history saved

### Phase 4: Frontend Integration & UX (Week 7)
**Goal**: Connect frontend to real backend

**Tasks**:
1. Update all API calls to real endpoints
2. Implement error handling and loading states
3. Display AI answers in chat interface
4. Show user question history
5. Implement FAQ selection logic
6. Add search functionality for channels
7. Polish UI/UX based on real data

**Deliverables**:
- Fully functional frontend
- Real-time question answering
- User history tracking

### Phase 5: Advanced Features & Optimization (Week 8-9)
**Goal**: Enhance functionality and performance

**Tasks**:
1. Implement FAQ auto-generation (popular questions)
2. Add answer confidence scoring
3. Implement caching for common questions (ElastiCache optional)
4. Add source citations to answers
5. Optimize chunking strategy based on testing
6. Implement rate limiting per user
7. Add CloudWatch monitoring and alerts
8. Performance testing and optimization

**Deliverables**:
- FAQ system working
- Source citations visible
- Monitoring dashboards
- Optimized performance

### Phase 6: Production Readiness (Week 10)
**Goal**: Prepare for production deployment

**Tasks**:
1. Security audit (IAM roles, encryption at rest/transit)
2. Set up staging environment
3. Implement CI/CD pipeline (GitHub Actions)
4. Load testing
5. Cost optimization review
6. Documentation (API docs, deployment guide)
7. Backup and disaster recovery setup
8. Production deployment

**Deliverables**:
- Production-ready infrastructure
- Documented deployment process
- Monitoring and alerting configured

## API Endpoints Specification

### Authentication
- `POST /api/auth/signup` - Register new user
- `POST /api/auth/login` - Login and get JWT token
- `POST /api/auth/refresh` - Refresh JWT token
- `POST /api/auth/logout` - Invalidate token

### Channels
- `GET /api/channels` - List all active channels (public)
- `GET /api/channels/:id` - Get channel details (public)
- `POST /api/channels` - Create new channel (admin only)
- `PUT /api/channels/:id` - Update channel (admin only)
- `DELETE /api/channels/:id` - Delete channel (admin only)
- `POST /api/channels/:id/upload` - Upload channel data (admin only)

### Questions
- `GET /api/channels/:id/questions` - Get user's questions for channel (authenticated)
- `POST /api/channels/:id/questions` - Ask new question (authenticated)
- `GET /api/channels/:id/faqs` - Get FAQs for channel (public)
- `PUT /api/questions/:id/faq` - Mark question as FAQ (admin only)

### Admin (Future)
- `GET /api/admin/stats` - Platform statistics
- `GET /api/admin/users` - List users
- `POST /api/admin/sync/:channel_id` - Manually trigger KB sync

## Environment Variables

### Backend (.env)
```bash
# AWS
AWS_REGION=us-east-1
AWS_ACCOUNT_ID=xxxxxxxxxxxx

# DynamoDB
DYNAMODB_USERS_TABLE=askly-prod-users
DYNAMODB_CHANNELS_TABLE=askly-prod-channels
DYNAMODB_QUESTIONS_TABLE=askly-prod-questions

# S3
S3_RAW_BUCKET=askly-prod-raw-data
S3_PROCESSED_BUCKET=askly-prod-processed-data
S3_ASSETS_BUCKET=askly-prod-assets

# Cognito
COGNITO_USER_POOL_ID=us-east-1_xxxxxxxxx
COGNITO_CLIENT_ID=xxxxxxxxxxxxxxxxxxxxxxxxxx
COGNITO_REGION=us-east-1

# Bedrock
BEDROCK_MODEL_ID=anthropic.claude-3-5-sonnet-20241022-v2:0
BEDROCK_REGION=us-east-1

# API
API_PORT=8080
JWT_SECRET=your-secret-key
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://askly.com

# Feature Flags
ENABLE_RATE_LIMITING=true
MAX_QUESTIONS_PER_HOUR=20
```

### Frontend (.env.local)
```bash
NEXT_PUBLIC_API_URL=https://api.askly.com
NEXT_PUBLIC_AWS_REGION=us-east-1
NEXT_PUBLIC_COGNITO_USER_POOL_ID=us-east-1_xxxxxxxxx
NEXT_PUBLIC_COGNITO_CLIENT_ID=xxxxxxxxxxxxxxxxxxxxxxxxxx
```

## Cost Estimation (Monthly - MVP Scale)

### AWS Services
- **DynamoDB**: ~$5-10 (On-demand pricing, low traffic)
- **S3**: ~$5-15 (Data storage + requests)
- **Cognito**: Free tier (50,000 MAUs)
- **API Gateway**: ~$3.50 per million requests
- **Lambda**: Free tier covers MVP (1M free requests)
- **Bedrock Knowledge Bases**: ~$0.10 per 1K queries + storage
- **OpenSearch Serverless**: ~$700/month (OCU-based, can optimize)
- **CloudWatch**: ~$5-10 (logs + metrics)

**Estimated MVP Total**: $750-850/month

### Optimization Strategies
- Use OpenSearch Serverless minimally (1-2 OCUs)
- Implement caching to reduce Bedrock calls
- Use S3 Intelligent-Tiering
- Reserved capacity for DynamoDB if usage is predictable
- Consider Bedrock Batch for bulk processing

## Security Considerations

1. **Authentication & Authorization**
   - JWT tokens with short expiry (15 min access, 7 day refresh)
   - Cognito integration for user management
   - Role-based access control (user, admin)

2. **Data Protection**
   - Encrypt S3 buckets (SSE-S3 or SSE-KMS)
   - DynamoDB encryption at rest (enabled by default)
   - HTTPS/TLS for all API communication
   - VPC for sensitive resources (optional for MVP)

3. **API Security**
   - Rate limiting per user/IP
   - Input validation and sanitization
   - CORS properly configured
   - API keys for admin operations

4. **IAM Best Practices**
   - Least privilege principle
   - Separate roles for each Lambda function
   - No hardcoded credentials
   - Use IAM roles for EC2/Lambda

5. **Monitoring & Auditing**
   - CloudWatch logs for all operations
   - CloudTrail for API calls
   - Alerts for suspicious activity
   - Regular security audits

## Testing Strategy

### Unit Tests (Go)
- Repository layer tests (mocked DynamoDB)
- Service layer tests (mocked dependencies)
- Utility function tests
- Target: 80%+ coverage

### Integration Tests
- API endpoint tests (local DynamoDB)
- Cognito integration tests
- S3 upload/download tests
- Bedrock API tests

### End-to-End Tests
- User registration → login → ask question flow
- Data upload → processing → KB sync flow
- Frontend Cypress/Playwright tests

### Load Testing
- Apache JMeter or k6
- Simulate 100-1000 concurrent users
- Test API response times
- Monitor AWS service limits

## Monitoring & Observability

### Metrics to Track
- API response times (p50, p95, p99)
- Error rates by endpoint
- Bedrock query latency
- DynamoDB read/write capacity
- Lambda execution duration
- User signup/login success rate
- Questions asked per hour
- Knowledge base sync success rate

### Alerts
- API error rate > 5%
- Bedrock query failures
- Lambda timeouts
- DynamoDB throttling
- S3 bucket policy violations
- Cognito authentication failures

### Dashboards
- CloudWatch dashboard for key metrics
- Real-time user activity
- Cost tracking dashboard
- System health overview

## Future Enhancements

1. **Multi-modal Support**
   - Video timestamp linking
   - Image/screenshot analysis
   - Audio clips in responses

2. **Advanced AI Features**
   - Conversational follow-up questions
   - Multi-turn dialogue
   - Cross-channel knowledge synthesis

3. **Social Features**
   - Upvote/downvote answers
   - Share questions publicly
   - User profiles and reputation

4. **Analytics**
   - Popular topics per channel
   - User engagement metrics
   - Creator insights dashboard

5. **Scalability**
   - Multi-region deployment
   - CDN for static assets (CloudFront)
   - ElastiCache for caching
   - GraphQL API option

## Success Metrics

### Technical KPIs
- API uptime: 99.9%
- Average response time: < 2s
- Question answer accuracy: > 85%
- Knowledge base sync success: > 95%

### Business KPIs
- Monthly active users
- Questions asked per user
- User retention rate
- Channel growth rate
- Answer satisfaction rating

## Risks & Mitigation

### Technical Risks
1. **Bedrock costs higher than expected**
   - Mitigation: Implement aggressive caching, monitor usage

2. **OpenSearch Serverless expensive**
   - Mitigation: Consider alternatives (Pinecone, Weaviate)

3. **Poor answer quality**
   - Mitigation: Optimize chunking, fine-tune prompts, add human review

4. **DynamoDB hot partition**
   - Mitigation: Proper key design, use composite keys

### Business Risks
1. **Low user adoption**
   - Mitigation: Start with 2-3 popular channels, gather feedback

2. **Data sourcing challenges**
   - Mitigation: Build automated YouTube transcript fetching

3. **Creator/copyright concerns**
   - Mitigation: Clear ToS, attribution, option for creators to opt-out

## Getting Started Checklist

- [ ] Set up AWS account and configure credentials
- [ ] Install Go 1.23+, Pulumi CLI, Node.js 20+
- [ ] Clone repository and install dependencies
- [ ] Configure environment variables
- [ ] Deploy initial infrastructure (Pulumi)
- [ ] Set up local development environment
- [ ] Create first channel and upload test data
- [ ] Test question answering locally
- [ ] Deploy to staging environment
- [ ] Run integration tests
- [ ] Document setup process
- [ ] Deploy to production

---

**Document Version**: 1.0
**Last Updated**: 2025-10-14
**Author**: Development Team
**Status**: Draft - Ready for Implementation