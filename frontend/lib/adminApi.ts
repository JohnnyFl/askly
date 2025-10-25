/**
 * Admin API Service Layer
 * Centralized service for all admin-related API calls
 */

const API_URL = process.env.NEXT_PUBLIC_API_URL;

// Helper function to get auth token
const getAuthToken = (): string | null => {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('admin_auth_token');
};

// Helper function to create headers
const createHeaders = (includeAuth = true): HeadersInit => {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (includeAuth) {
    const token = getAuthToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
  }

  return headers;
};

// Error handler
class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

const handleResponse = async (response: Response) => {
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
    throw new ApiError(response.status, errorData.error || `HTTP ${response.status}`);
  }
  return response.json();
};

// ==================== Authentication ====================

export interface AdminLoginRequest {
  email: string;
  password: string;
}

export interface AdminLoginResponse {
  token: string;
  admin: {
    id: string;
    email: string;
    name?: string;
  };
}

export const adminLogin = async (credentials: AdminLoginRequest): Promise<AdminLoginResponse> => {
  const response = await fetch(`${API_URL}/admin/login`, {
    method: 'POST',
    headers: createHeaders(false),
    body: JSON.stringify(credentials),
  });
  return handleResponse(response);
};

// ==================== Dashboard Stats ====================

export interface DashboardStats {
  totalChannels: number;
  totalUsers: number;
  totalQuestions: number;
  totalTranscriptions: number;
}

export const getDashboardStats = async (): Promise<DashboardStats> => {
  const response = await fetch(`${API_URL}/admin/stats`, {
    headers: createHeaders(),
  });
  return handleResponse(response);
};

// ==================== Activity ====================
// Note: Activity endpoint not yet implemented in backend
// Keeping types for future implementation

export interface Activity {
  id: string;
  type: 'channel' | 'user' | 'question' | 'transcription';
  description: string;
  timestamp: string;
}

export interface ActivityResponse {
  activities: Activity[];
}

// Placeholder - returns empty activities until backend endpoint is implemented
export const getRecentActivity = async (): Promise<ActivityResponse> => {
  return { activities: [] };
};

// ==================== Channels ====================

export interface Channel {
  id: string;
  name: string;
  description: string;
  avatar_url: string;
  youtube_url: string;
  created_at: string;
  transcription_count?: number;
  status?: 'active' | 'inactive';
}

export interface ChannelsResponse {
  channels: Channel[];
}

export const getChannels = async (): Promise<ChannelsResponse> => {
  const response = await fetch(`${API_URL}/admin/channels`, {
    headers: createHeaders(),
  });
  return handleResponse(response);
};

export interface ChannelResponse {
  channel: Channel;
}

export const getChannel = async (channelId: string): Promise<ChannelResponse> => {
  const response = await fetch(`${API_URL}/admin/channels/${channelId}`, {
    headers: createHeaders(),
  });
  return handleResponse(response);
};

// ==================== Transcriptions ====================

export interface Transcription {
  id: string;
  file_name: string;
  file_size: number;
  upload_date: string;
  status: 'processing' | 'completed' | 'failed';
  error_message?: string;
}

export interface TranscriptionsResponse {
  transcriptions: Transcription[];
}

export const getTranscriptions = async (channelId: string): Promise<TranscriptionsResponse> => {
  const response = await fetch(`${API_URL}/admin/channels/${channelId}/transcriptions`, {
    headers: createHeaders(),
  });
  return handleResponse(response);
};

export interface UploadTranscriptionResponse {
  transcription: Transcription;
  message: string;
}

export const uploadTranscription = async (
  channelId: string,
  file: File,
  onProgress?: (progress: number) => void
): Promise<UploadTranscriptionResponse> => {
  return new Promise((resolve, reject) => {
    const formData = new FormData();
    formData.append('file', file);

    const xhr = new XMLHttpRequest();

    // Progress tracking
    if (onProgress) {
      xhr.upload.addEventListener('progress', (e) => {
        if (e.lengthComputable) {
          const progress = Math.round((e.loaded / e.total) * 100);
          onProgress(progress);
        }
      });
    }

    // Success handler
    xhr.addEventListener('load', () => {
      if (xhr.status === 200 || xhr.status === 201) {
        try {
          const data = JSON.parse(xhr.responseText);
          resolve(data);
        } catch (error) {
          reject(new ApiError(xhr.status, 'Invalid response format'));
        }
      } else {
        try {
          const errorData = JSON.parse(xhr.responseText);
          reject(new ApiError(xhr.status, errorData.error || 'Upload failed'));
        } catch {
          reject(new ApiError(xhr.status, 'Upload failed'));
        }
      }
    });

    // Error handler
    xhr.addEventListener('error', () => {
      reject(new ApiError(0, 'Network error occurred'));
    });

    // Abort handler
    xhr.addEventListener('abort', () => {
      reject(new ApiError(0, 'Upload cancelled'));
    });

    // Send request
    const token = getAuthToken();
    xhr.open('POST', `${API_URL}/admin/channels/${channelId}/transcriptions`);
    if (token) {
      xhr.setRequestHeader('Authorization', `Bearer ${token}`);
    }
    xhr.send(formData);
  });
};

export const deleteTranscription = async (
  channelId: string,
  transcriptionId: string
): Promise<{ message: string }> => {
  const response = await fetch(
    `${API_URL}/admin/channels/${channelId}/transcriptions/${transcriptionId}`,
    {
      method: 'DELETE',
      headers: createHeaders(),
    }
  );
  return handleResponse(response);
};

export const downloadTranscription = async (
  channelId: string,
  transcriptionId: string
): Promise<Blob> => {
  const token = getAuthToken();
  const headers: HeadersInit = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(
    `${API_URL}/admin/channels/${channelId}/transcriptions/${transcriptionId}/download`,
    {
      headers,
    }
  );

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: 'Download failed' }));
    throw new ApiError(response.status, errorData.error || 'Download failed');
  }

  return response.blob();
};

// ==================== Utility Functions ====================

export const isAuthenticated = (): boolean => {
  return !!getAuthToken();
};

export const logout = (): void => {
  if (typeof window !== 'undefined') {
    localStorage.removeItem('admin_auth_token');
  }
};