'use client';

import { useState, useCallback } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Progress } from '@/components/ui/progress';
import { Upload, X, FileText, AlertCircle, CheckCircle2 } from 'lucide-react';
import { Alert, AlertDescription } from '@/components/ui/alert';

type FileWithProgress = {
  file: File;
  progress: number;
  status: 'pending' | 'uploading' | 'success' | 'error';
  error?: string;
};

type FileUploadProps = {
  channelId: string;
  onUploadComplete?: () => void;
};

const ALLOWED_FILE_TYPES = ['.txt', '.srt', '.vtt', '.json'];
const MAX_FILE_SIZE = 50 * 1024 * 1024; // 50MB

export function FileUpload({ channelId, onUploadComplete }: FileUploadProps) {
  const [files, setFiles] = useState<FileWithProgress[]>([]);
  const [isDragging, setIsDragging] = useState(false);
  const [globalError, setGlobalError] = useState('');

  const validateFile = (file: File): string | null => {
    const extension = '.' + file.name.split('.').pop()?.toLowerCase();
    if (!ALLOWED_FILE_TYPES.includes(extension)) {
      return `Invalid file type. Allowed: ${ALLOWED_FILE_TYPES.join(', ')}`;
    }
    if (file.size > MAX_FILE_SIZE) {
      return `File too large. Max size: 50MB`;
    }
    return null;
  };

  const handleFiles = useCallback((newFiles: FileList | File[]) => {
    setGlobalError('');
    const filesArray = Array.from(newFiles);
    const validFiles: FileWithProgress[] = [];

    filesArray.forEach((file) => {
      const error = validateFile(file);
      if (error) {
        setGlobalError((prev) => prev + (prev ? '\n' : '') + `${file.name}: ${error}`);
      } else {
        validFiles.push({
          file,
          progress: 0,
          status: 'pending',
        });
      }
    });

    setFiles((prev) => [...prev, ...validFiles]);
  }, []);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  }, []);

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setIsDragging(false);
      handleFiles(e.dataTransfer.files);
    },
    [handleFiles]
  );

  const handleFileInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      if (e.target.files) {
        handleFiles(e.target.files);
      }
    },
    [handleFiles]
  );

  const removeFile = (index: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== index));
  };

  const uploadFile = async (fileWithProgress: FileWithProgress, index: number) => {
    try {
      const token = localStorage.getItem('admin_auth_token');

      // Update status to uploading
      setFiles((prev) =>
        prev.map((f, i) => (i === index ? { ...f, status: 'uploading', progress: 10 } : f))
      );

      // Read file content as text
      const fileContent = await fileWithProgress.file.text();

      // Extract video title from filename (remove extension)
      const videoTitle = fileWithProgress.file.name.replace(/\.[^/.]+$/, '');

      // Update progress to 30% after reading file
      setFiles((prev) =>
        prev.map((f, i) => (i === index ? { ...f, progress: 30 } : f))
      );

      // Send as JSON to match backend expectations
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/admin/channels/${channelId}/transcriptions`,
        {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            video_title: videoTitle,
            video_id: '', // Optional
            video_url: '', // Optional
            transcription: fileContent,
          }),
        }
      );

      // Update progress to 90% after sending
      setFiles((prev) =>
        prev.map((f, i) => (i === index ? { ...f, progress: 90 } : f))
      );

      if (response.ok) {
        setFiles((prev) =>
          prev.map((f, i) => (i === index ? { ...f, status: 'success', progress: 100 } : f))
        );
        if (onUploadComplete) {
          onUploadComplete();
        }
      } else {
        const errorData = await response.json();
        const errorMessage = errorData.error || 'Upload failed';
        setFiles((prev) =>
          prev.map((f, i) =>
            i === index ? { ...f, status: 'error', error: errorMessage } : f
          )
        );
      }
    } catch (error: any) {
      setFiles((prev) =>
        prev.map((f, i) =>
          i === index ? { ...f, status: 'error', error: error.message || 'Upload failed' } : f
        )
      );
    }
  };

  const uploadAll = () => {
    files.forEach((file, index) => {
      if (file.status === 'pending') {
        uploadFile(file, index);
      }
    });
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
  };

  return (
    <div className="space-y-6">
      {/* Drop Zone */}
      <Card
        className={`bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border-2 border-dashed transition-all ${
          isDragging
            ? 'border-[#6D28D9] bg-purple-50/50 scale-[1.02]'
            : 'border-purple-200 hover:border-[#6D28D9]/50'
        }`}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
      >
        <CardContent className="p-12">
          <div className="flex flex-col items-center justify-center text-center space-y-4">
            <div className="h-20 w-20 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-2xl flex items-center justify-center shadow-lg animate-float">
              <Upload className="h-10 w-10 text-white" />
            </div>
            <div>
              <h3 className="text-2xl font-bold text-slate-900 mb-2">Upload Transcriptions</h3>
              <p className="text-slate-600 mb-4">
                Drag and drop files here or click to browse
              </p>
              <p className="text-sm text-slate-500">
                Supported formats: {ALLOWED_FILE_TYPES.join(', ')} • Max size: 50MB
              </p>
            </div>
            <input
              type="file"
              multiple
              accept={ALLOWED_FILE_TYPES.join(',')}
              onChange={handleFileInputChange}
              className="hidden"
              id="file-upload"
            />
            <label htmlFor="file-upload">
              <Button
                type="button"
                className="rounded-xl bg-gradient-to-r from-[#6D28D9] to-[#5B21B6] hover:from-[#5B21B6] hover:to-[#4C1D95] shadow-lg cursor-pointer"
                asChild
              >
                <span>Browse Files</span>
              </Button>
            </label>
          </div>
        </CardContent>
      </Card>

      {/* Global Error */}
      {globalError && (
        <Alert variant="destructive" className="rounded-xl">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription className="whitespace-pre-line">{globalError}</AlertDescription>
        </Alert>
      )}

      {/* File List */}
      {files.length > 0 && (
        <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50">
          <CardContent className="p-6">
            <div className="flex items-center justify-between mb-6">
              <h4 className="text-lg font-bold text-slate-900">
                Files ({files.length})
              </h4>
              <Button
                onClick={uploadAll}
                disabled={files.every((f) => f.status !== 'pending')}
                className="rounded-xl bg-gradient-to-r from-[#6D28D9] to-[#5B21B6] hover:from-[#5B21B6] hover:to-[#4C1D95] shadow-lg gap-2"
              >
                <Upload className="h-4 w-4" />
                Upload All
              </Button>
            </div>

            <div className="space-y-4">
              {files.map((fileWithProgress, index) => (
                <div
                  key={index}
                  className="p-4 rounded-2xl bg-gradient-to-br from-purple-50 to-pink-50 border border-purple-100"
                >
                  <div className="flex items-start gap-4">
                    <div className="h-10 w-10 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-xl flex items-center justify-center shrink-0">
                      <FileText className="h-5 w-5 text-white" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-start justify-between gap-2 mb-2">
                        <div className="flex-1 min-w-0">
                          <p className="font-medium text-slate-900 truncate">
                            {fileWithProgress.file.name}
                          </p>
                          <p className="text-sm text-slate-500">
                            {formatFileSize(fileWithProgress.file.size)}
                          </p>
                        </div>
                        <div className="flex items-center gap-2 shrink-0">
                          {fileWithProgress.status === 'success' && (
                            <CheckCircle2 className="h-5 w-5 text-green-500" />
                          )}
                          {fileWithProgress.status === 'error' && (
                            <AlertCircle className="h-5 w-5 text-red-500" />
                          )}
                          {fileWithProgress.status === 'pending' && (
                            <Button
                              size="icon"
                              variant="ghost"
                              onClick={() => removeFile(index)}
                              className="h-8 w-8 rounded-lg hover:bg-purple-100"
                            >
                              <X className="h-4 w-4" />
                            </Button>
                          )}
                        </div>
                      </div>

                      {fileWithProgress.status === 'uploading' && (
                        <div className="space-y-1">
                          <Progress value={fileWithProgress.progress} className="h-2" />
                          <p className="text-xs text-slate-500">{fileWithProgress.progress}%</p>
                        </div>
                      )}

                      {fileWithProgress.status === 'error' && (
                        <p className="text-sm text-red-600">{fileWithProgress.error}</p>
                      )}

                      {fileWithProgress.status === 'success' && (
                        <p className="text-sm text-green-600">Upload complete</p>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}