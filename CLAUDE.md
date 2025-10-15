# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Askly is a platform that allows users to extract knowledge from YouTube channels and get instant answers to questions about the content without watching hours of video. The project consists of:

- **Frontend**: Next.js 13 application with TypeScript, Radix UI components, and Tailwind CSS
- **Infrastructure**: Pulumi-based infrastructure as code using Go to provision AWS resources (S3 buckets, Cognito user pools)
- **Backend**: Currently empty placeholder directory (future backend implementation)

## Architecture

### Frontend (Next.js App Router)

The frontend uses the Next.js 13 App Router pattern with client-side components. Key pages:

- `/app/page.tsx` - Landing page with product marketing
- `/app/login/page.tsx` - User authentication
- `/app/signup/page.tsx` - User registration
- `/app/channels/page.tsx` - Browse available YouTube channels
- `/app/channels/[id]/page.tsx` - Chat interface for specific channel

**Authentication**: Uses localStorage to store auth tokens (`auth_token` key) with Bearer token authentication to backend API.

**API Communication**: Fetches from `process.env.NEXT_PUBLIC_API_URL` (must be configured).

**UI Components**: All UI components are from shadcn/ui (Radix UI + Tailwind) located in `frontend/components/ui/`. Import pattern: `@/components/ui/*`

**Path Alias**: TypeScript configured with `@/*` mapping to frontend root directory.

### Infrastructure (Pulumi + Go)

Located in `infra/` directory. Uses Pulumi to provision AWS resources:

- S3 bucket for storage
- Cognito User Pool for authentication

**Module**: `askly` (defined in `infra/go.mod`)
**Pulumi Version**: v3 SDK with AWS provider v7

## Development Commands

### Frontend

All commands run from `frontend/` directory:

```bash
npm run dev          # Start development server
npm run build        # Build production bundle
npm start            # Run production server
npm run lint         # Run ESLint
npm run typecheck    # Run TypeScript compiler check (no emit)
```

### Infrastructure

All commands run from `infra/` directory:

```bash
pulumi preview       # Preview infrastructure changes
pulumi up            # Deploy infrastructure
pulumi stack output  # View stack outputs (bucketName, userPoolId)
pulumi config get    # Get configuration values
pulumi config set    # Set configuration values
```

**Prerequisites for infra**:
- Go 1.23.11+
- AWS credentials configured
- Pulumi CLI installed

## Environment Variables

### Frontend

Required in `.env.local` or environment:

- `NEXT_PUBLIC_API_URL` - Backend API base URL

## Key Technical Details

### Frontend Stack
- Next.js 13.5.1 (App Router)
- React 18.2.0
- TypeScript 5.2.2
- Tailwind CSS + shadcn/ui components
- React Hook Form + Zod for form validation
- Lucide React for icons

### Design System
- Color scheme: Pink/Rose (`#FB7185`, `#EF4444`) and Purple (`#6D28D9`, `#5B21B6`)
- Rounded, modern UI with backdrop blur effects
- Gradient accents throughout

### Data Flow
1. Users authenticate via `/login` or `/signup` (stores JWT in localStorage)
2. Browse channels at `/channels` (fetches from API with Bearer token)
3. Click channel to chat at `/channels/[id]` (chat interface for Q&A)

## Important Notes

- Backend directory exists but is currently empty
- Frontend expects backend API to be running and accessible via `NEXT_PUBLIC_API_URL`
- Infrastructure provisions AWS resources but is not yet connected to application code
- All frontend routes except landing/login/signup require authentication (checked via localStorage token)