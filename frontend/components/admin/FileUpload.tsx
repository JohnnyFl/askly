'use client'

import { useEffect, useMemo } from 'react'
import Uppy from '@uppy/core'
import { Dashboard } from '@uppy/react'
import XHRUpload from '@uppy/xhr-upload'

// Import Uppy styles
import '@uppy/core/css/style.min.css';
import '@uppy/dashboard/css/style.min.css';

interface FileUploadProps {
  endpoint: string
  channelId?: string
  onUploadComplete?: (files: any[]) => void
  onUploadError?: (error: Error) => void
  allowedFileTypes?: string[]
  maxFileSize?: number
  maxNumberOfFiles?: number
  authTokenKey?: string // Key for retrieving auth token from localStorage
}

export default function FileUpload({
  endpoint,
  channelId,
  onUploadComplete,
  onUploadError,
  allowedFileTypes = ['.json', '.txt', '.csv', '.pdf'],
  maxFileSize = 10 * 1024 * 1024, // 10MB default
  maxNumberOfFiles = 10,
  authTokenKey = 'auth_token'
}: FileUploadProps) {
  // Create Uppy instance with useMemo to persist across renders
  const uppy = useMemo(() => {
    return new Uppy({
      restrictions: {
        maxFileSize,
        maxNumberOfFiles,
        allowedFileTypes: allowedFileTypes.length > 0 ? allowedFileTypes : undefined
      },
      autoProceed: false
    }).use(XHRUpload, {
      endpoint,
      method: 'POST',
      formData: true,
      fieldName: 'file',
      headers: () => {
        const token = localStorage.getItem(authTokenKey)
        return {
          Authorization: token ? `Bearer ${token}` : ''
        }
      },
      ...(channelId && {
        meta: { channelId }
      })
    })
  }, [endpoint, channelId, allowedFileTypes, maxFileSize, maxNumberOfFiles, authTokenKey])

  // Setup event handlers
  useEffect(() => {
    // Success handler
    uppy.on('complete', (result) => {
      if (result.successful && result.successful.length > 0) {
        onUploadComplete?.(result.successful)
      }

      if (result.failed && result.failed.length > 0) {
        const error = new Error(`${result.failed.length} file(s) failed to upload`)
        onUploadError?.(error)
      }
    })

    // Error handler
    uppy.on('upload-error', (_file, error) => {
      console.error('Upload error:', error)
      onUploadError?.(error as Error)
    })

    // Cleanup
    return () => {
      uppy.cancelAll()
    }
  }, [uppy, onUploadComplete, onUploadError])

  return (
    <div className="w-full">
      <Dashboard
        uppy={uppy}
        proudlyDisplayPoweredByUppy={false}
        height={350}
        note={`Max file size: ${maxFileSize / 1024 / 1024}MB, Max files: ${maxNumberOfFiles}`}
      />
    </div>
  )
}
