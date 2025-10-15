# Seed Data Script

This script populates the DynamoDB tables with initial data for the Askly POC.

## What it creates:

### 2 YouTube Channels
1. **Y Combinator** - Startup and entrepreneurship content
2. **Startup Grind** - Startup community and interviews

### 1 Admin User
- **Email**: admin@askly.com
- **Password**: admin123 (⚠️ Change after first login!)

## Prerequisites

- AWS credentials configured (via `aws configure` or environment variables)
- DynamoDB tables created (run `pulumi up` in `/infra` first)
- Go 1.23+ installed

## Usage

```bash
# Navigate to seed directory
cd infra/seed

# Install dependencies
go mod tidy

# Run the seed script
go run main.go
```

## Expected Output

```
✅ Created channel: Y Combinator (ID: CHANNEL#xxx)
✅ Created channel: Startup Grind (ID: CHANNEL#xxx)
✅ Created admin user: admin@askly.com
   Password: admin123 (change this after first login!)
   Admin ID: ADMIN#xxx
✅ Seed data created successfully!
```

## Notes

- This script is idempotent - running it multiple times will create duplicate entries
- The admin password is hardcoded for POC purposes only
- Channel IDs are randomly generated UUIDs
- All timestamps use RFC3339 format

## Troubleshooting

### Error: ResourceNotFoundException
**Problem**: DynamoDB tables don't exist

**Solution**: Deploy infrastructure first:
```bash
cd ../
pulumi up
```

### Error: AccessDeniedException
**Problem**: AWS credentials don't have DynamoDB permissions

**Solution**: Ensure your AWS credentials have `dynamodb:PutItem` permission

### Error: ValidationException
**Problem**: Table schema mismatch

**Solution**: Verify tables were created with correct schema (PK, SK structure)
