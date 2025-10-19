# Frontend-Backend Integration Analysis

## ✅ What's Working

### 1. **Admin Login**
- ✅ Endpoint exists: `login-admin` Lambda
- ✅ Request/Response match:
  - Frontend sends: `{ email, password }`
  - Backend expects: `{ email, password }`
  - Backend returns: `{ message, admin_id, token }`
- ✅ JWT token authentication implemented
- ✅ Role-based access control (admin vs user)

### 2. **User Authentication**
- ✅ User login exists: `login-user` Lambda
- ✅ User registration exists: `register-user` Lambda
- ✅ Proper separation between admin and user roles

### 3. **Basic Channel Operations**
- ✅ List channels exists: `list-channels` Lambda
- ✅ Get single channel exists: `get-channel` Lambda
- ✅ Channel model has required fields

---

## ❌ Integration Issues Found

### **CRITICAL: Field Name Mismatch**

**Problem:** Frontend expects `avatar_url` but backend returns `thumbnail_url`

**Location:**
- Backend: `backend/shared/models/channel.go` line 11
- Backend Lambda: `backend/lambdas/list-channels/main.go` line 26
- Frontend: `frontend/app/admin/channels/page.tsx`, `frontend/app/channels/page.tsx`

**Impact:** Channel avatars won't display correctly on frontend

**Fix Options:**
1. **Backend fix (RECOMMENDED):** Add `avatar_url` to Channel model or alias `thumbnail_url` as `avatar_url` in JSON response
2. **Frontend fix:** Change all `avatar_url` references to `thumbnail_url`

```go
// Backend fix - add to ChannelResponse in list-channels/main.go
type ChannelResponse struct {
    ID               string `json:"id"`
    Name             string `json:"name"`
    YoutubeChannelID string `json:"youtube_channel_id"`
    Description      string `json:"description"`
    ThumbnailURL     string `json:"thumbnail_url"`
    AvatarURL        string `json:"avatar_url"` // Add this - duplicate of ThumbnailURL
    YoutubeURL       string `json:"youtube_url"`
}

// Then in response mapping:
channelResponses[i] = ChannelResponse{
    // ... existing fields
    ThumbnailURL: channel.ThumbnailURL,
    AvatarURL:    channel.ThumbnailURL, // Map same value
}
```

---
ONLY TEXT CAN BE PASSED, NO FILES.

---

### **MISSING: Admin-Specific Endpoints**

The following endpoints are called by frontend but don't exist in backend:

#### 1. **Admin Channels List**
- **Frontend calls:** `GET /api/admin/channels`
- **Current:** Only `list-channels` exists but it's for users (checks `IsUser()`)
- **Fix needed:** Modify to allow admin access OR create separate admin endpoint
- **Expected response:**
```json
{
  "channels": [
    {
      "id": "string",
      "name": "string",
      "description": "string",
      "avatar_url": "string",
      "youtube_url": "string",
      "created_at": "2024-10-19T12:00:00Z",
      "transcription_count": 0,
      "status": "active" | "inactive"
    }
  ]
}
```

#### 2. **Get Single Channel (Admin)**
- **Frontend calls:** `GET /api/admin/channels/:id`
- **Current:** `get-channel` Lambda may exist but need to verify it allows admin access
- **Expected response:**
```json
{
  "channel": {
    "id": "string",
    "name": "string",
    "description": "string",
    "avatar_url": "string",
    "youtube_url": "string",
    "created_at": "2024-10-19T12:00:00Z",
    "transcription_count": 0
  }
}
```

#### 3. **List Transcriptions**
- **Frontend calls:** `GET /api/admin/channels/:id/transcriptions`
- **Expected response:**
```json
{
  "transcriptions": [
    {
      "id": "string",
      "video_title": "string",
      "video_url": "string",
      "upload_date": "2024-10-19T12:00:00Z",
      "status": "completed"
    }
  ]
}
```
- **Status:** ❌ Lambda doesn't exist (can reuse existing transcription queries)

#### 4. **Delete Transcription**
- **Frontend calls:** `DELETE /api/admin/channels/:id/transcriptions/:transcriptionId`
- **Status:** ❌ Lambda doesn't exist

---

## 📊 **Data Model Gaps**

### **Transcription Model:**

Current backend model has all required fields. Frontend needs to adapt to use:
- `video_title` (already exists) instead of `file_name`
- Calculate size from `full_text` length if needed
- `uploaded_at` (already exists) maps to `upload_date`
- Add `status` field: `"completed"` (default) or `"failed"`

---

## 🔧 **Quick Fixes**

### **Fix 1: Update list-channels to support admins**

`backend/lambdas/list-channels/main.go` line 51-54:
```go
// CURRENT (line 52):
if !claims.IsUser() {
    return response(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
}

// CHANGE TO:
if !claims.IsUser() && !claims.IsAdmin() {
    return response(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
}
```

### **Fix 2: Add avatar_url to response**

`backend/lambdas/list-channels/main.go`:
```go
type ChannelResponse struct {
    ID               string `json:"id"`
    Name             string `json:"name"`
    YoutubeChannelID string `json:"youtube_channel_id"`
    Description      string `json:"description"`
    ThumbnailURL     string `json:"thumbnail_url"`
    AvatarURL        string `json:"avatar_url"` // ADD THIS
    YoutubeURL       string `json:"youtube_url"`
    CreatedAt        string `json:"created_at"`  // ADD THIS for admin panel
}

// In mapping (line 70-78):
channelResponses[i] = ChannelResponse{
    ID:               strings.TrimPrefix(channel.PK, "CHANNEL#"),
    Name:             channel.Name,
    YoutubeChannelID: channel.YoutubeChannelID,
    Description:      channel.Description,
    ThumbnailURL:     channel.ThumbnailURL,
    AvatarURL:        channel.ThumbnailURL, // ADD THIS
    YoutubeURL:       channel.YoutubeURL,
    CreatedAt:        channel.CreatedAt.Format(time.RFC3339), // ADD THIS
}
```

---

## 📋 **Missing Lambda Functions Needed**

Create these new Lambda functions:

1. ✅ `login-admin` - EXISTS
2. ✅ `upload-transcription` - EXISTS (accepts text via JSON)
3. ❌ `admin-stats` - Get dashboard statistics
4. ❌ `admin-activity` - Get recent activity
5. ❌ `list-transcriptions` - List transcriptions for a channel
6. ❌ `delete-transcription` - Delete a transcription

---

## 🌐 **Environment Setup**

### **Frontend needs:**

Create `frontend/.env.local`:
```env
NEXT_PUBLIC_API_URL=https://your-api-gateway-url.amazonaws.com/prod
```

Or for local development:
```env
NEXT_PUBLIC_API_URL=http://localhost:3000
```

---

## 🚦 **Priority Action Items**

### **P0 - CRITICAL (Blocks Basic Functionality)**
1. Fix field name mismatch: `thumbnail_url` → `avatar_url`
2. Update list-channels to allow admin access
3. Add `created_at` to ChannelResponse

### **P1 - HIGH (Admin Panel Features)**
4. Create `admin-stats` Lambda
5. Create `list-transcriptions` Lambda
6. Create `delete-transcription` Lambda
7. Add `status` field to Transcription model

### **P2 - MEDIUM (Nice to Have)**
8. Create `admin-activity` Lambda
9. Add `transcription_count` to Channel queries

---

## 📝 **Testing Checklist**

Once fixes are implemented:

- [ ] Admin login with valid credentials
- [ ] Admin dashboard loads with stats
- [ ] Channels list displays with avatars
- [ ] Click channel → navigate to upload page
- [ ] Upload transcription text → success
- [ ] View transcriptions list
- [ ] Delete transcription
- [ ] View recent activity

---

## 💡 **Recommendations**

1. **Add TypeScript types** to backend (consider using Protocol Buffers or shared types)
2. **API Gateway configuration** - ensure CORS is properly configured
3. **Error handling** - standardize error response format across all lambdas
4. **Logging** - backend already has good logging, ensure CloudWatch is configured
5. **Text size limits** - backend should validate reasonable text size limits
6. **Rate limiting** - consider adding rate limits to upload endpoint

---

## 🎯 **Next Steps**

1. Review this report with backend team
2. Prioritize P0 fixes (field naming, admin access)
3. Create GitHub issues for each missing Lambda
4. Update frontend FileUpload component to send text via JSON instead of multipart/form-data
5. Test integration after fixes
6. Deploy and verify in staging environment