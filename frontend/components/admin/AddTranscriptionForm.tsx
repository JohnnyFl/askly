'use client';

import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Loader2, Plus, FileText } from 'lucide-react';

interface AddTranscriptionFormProps {
  channelId: string;
  onSuccess?: () => void;
  onError?: (error: Error) => void;
}

export function AddTranscriptionForm({
  channelId,
  onSuccess,
  onError,
}: AddTranscriptionFormProps) {
  const [videoTitle, setVideoTitle] = useState('');
  const [videoUrl, setVideoUrl] = useState('');
  const [transcriptionText, setTranscriptionText] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!videoTitle.trim() || !transcriptionText.trim()) {
      onError?.(new Error('Video title and transcription text are required'));
      return;
    }

    setIsSubmitting(true);

    try {
      const token = localStorage.getItem('admin_auth_token');
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/admin/channels/${channelId}/transcriptions`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
          },
          body: JSON.stringify({
            video_title: videoTitle,
            video_url: videoUrl,
            transcription_text: transcriptionText,
          }),
        }
      );

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to add transcription');
      }

      // Clear form
      setVideoTitle('');
      setVideoUrl('');
      setTranscriptionText('');

      onSuccess?.();
    } catch (error) {
      console.error('Error adding transcription:', error);
      onError?.(error as Error);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50">
      <CardHeader className="p-6 pb-4">
        <div className="flex items-center gap-3 mb-2">
          <div className="h-10 w-10 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-xl flex items-center justify-center shadow-lg">
            <Plus className="h-5 w-5 text-white" />
          </div>
          <div>
            <CardTitle className="text-2xl font-bold text-slate-900">
              Add New Transcription
            </CardTitle>
            <CardDescription className="text-base mt-1">
              Manually enter video details and transcription text
            </CardDescription>
          </div>
        </div>
      </CardHeader>

      <CardContent className="p-6 pt-0">
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Video Title */}
          <div className="space-y-2">
            <Label htmlFor="video-title" className="text-sm font-semibold text-slate-900">
              Video Title <span className="text-red-500">*</span>
            </Label>
            <Input
              id="video-title"
              type="text"
              placeholder="Enter video title..."
              value={videoTitle}
              onChange={(e) => setVideoTitle(e.target.value)}
              className="h-12 rounded-xl border-2 border-purple-100 focus:border-[#6D28D9]"
              required
            />
          </div>

          {/* Video URL */}
          <div className="space-y-2">
            <Label htmlFor="video-url" className="text-sm font-semibold text-slate-900">
              Video URL <span className="text-slate-400 font-normal">(optional)</span>
            </Label>
            <Input
              id="video-url"
              type="url"
              placeholder="https://www.youtube.com/watch?v=..."
              value={videoUrl}
              onChange={(e) => setVideoUrl(e.target.value)}
              className="h-12 rounded-xl border-2 border-purple-100 focus:border-[#6D28D9]"
            />
          </div>

          {/* Transcription Text */}
          <div className="space-y-2">
            <Label htmlFor="transcription-text" className="text-sm font-semibold text-slate-900">
              Transcription Text <span className="text-red-500">*</span>
            </Label>
            <Textarea
              id="transcription-text"
              placeholder="Paste or type the video transcription here..."
              value={transcriptionText}
              onChange={(e) => setTranscriptionText(e.target.value)}
              className="min-h-[200px] rounded-xl border-2 border-purple-100 focus:border-[#6D28D9] resize-y"
              required
            />
            <p className="text-xs text-slate-500">
              {transcriptionText.length.toLocaleString()} characters
            </p>
          </div>

          {/* Submit Button */}
          <Button
            type="submit"
            disabled={isSubmitting || !videoTitle.trim() || !transcriptionText.trim()}
            className="w-full h-12 rounded-xl bg-gradient-to-r from-[#6D28D9] to-[#5B21B6] hover:from-[#5B21B6] hover:to-[#4C1D95] text-base font-semibold"
          >
            {isSubmitting ? (
              <>
                <Loader2 className="h-5 w-5 mr-2 animate-spin" />
                Adding Transcription...
              </>
            ) : (
              <>
                <FileText className="h-5 w-5 mr-2" />
                Add Transcription
              </>
            )}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
