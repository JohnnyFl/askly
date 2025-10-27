'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Loader2, Plus, Video, ExternalLink, CheckCircle2 } from 'lucide-react';

interface AddChannelDialogProps {
  onSuccess?: () => void;
  onError?: (error: Error) => void;
}

// Helper function to extract channel ID from YouTube URL
function extractChannelId(url: string): string | null {
  if (!url) return null;

  try {
    const urlObj = new URL(url);

    // Format: youtube.com/channel/CHANNEL_ID
    if (urlObj.pathname.startsWith('/channel/')) {
      return urlObj.pathname.split('/channel/')[1].split('/')[0];
    }

    // Format: youtube.com/@username - we'll need the actual channel ID
    // In this case, we return null and ask user to provide it
    if (urlObj.pathname.startsWith('/@')) {
      return null;
    }

    // Format: youtube.com/c/username - also needs manual entry
    if (urlObj.pathname.startsWith('/c/')) {
      return null;
    }

    return null;
  } catch {
    return null;
  }
}

export function AddChannelDialog({ onSuccess, onError }: AddChannelDialogProps) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [thumbnailUrl, setThumbnailUrl] = useState('');
  const [youtubeUrl, setYoutubeUrl] = useState('');
  const [youtubeChannelId, setYoutubeChannelId] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [autoExtracted, setAutoExtracted] = useState(false);

  // Auto-extract channel ID from YouTube URL
  useEffect(() => {
    const channelId = extractChannelId(youtubeUrl);
    if (channelId) {
      setYoutubeChannelId(channelId);
      setAutoExtracted(true);
    } else if (youtubeUrl && !youtubeChannelId) {
      setAutoExtracted(false);
    }
  }, [youtubeUrl]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!name.trim() || !youtubeChannelId.trim()) {
      onError?.(new Error('Channel name and YouTube channel ID are required'));
      return;
    }

    setIsSubmitting(true);

    try {
      const token = localStorage.getItem('admin_auth_token');
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/admin/channels`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          name: name.trim(),
          description: description.trim(),
          thumbnail_url: thumbnailUrl.trim(),
          youtube_url: youtubeUrl.trim(),
          youtube_channel_id: youtubeChannelId.trim(),
        }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || 'Failed to create channel');
      }

      // Clear form and close dialog
      setName('');
      setDescription('');
      setThumbnailUrl('');
      setYoutubeUrl('');
      setYoutubeChannelId('');
      setOpen(false);

      onSuccess?.();
    } catch (error) {
      console.error('Error creating channel:', error);
      onError?.(error as Error);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button className="gap-2 rounded-xl bg-gradient-to-r from-[#6D28D9] to-[#5B21B6] hover:from-[#5B21B6] hover:to-[#4C1D95] shadow-lg h-14 px-6 text-base">
          <Plus className="h-5 w-5" />
          Add New Channel
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl rounded-3xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <div className="flex items-center gap-3 mb-2">
            <div className="h-12 w-12 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-xl flex items-center justify-center shadow-lg">
              <Video className="h-6 w-6 text-white" />
            </div>
            <div>
              <DialogTitle className="text-2xl font-bold">Add New Channel</DialogTitle>
              <DialogDescription className="text-base mt-1">
                Add a YouTube channel to the platform
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6 mt-4">
          {/* Instructions */}
          <div className="bg-purple-50 border border-purple-200 rounded-xl p-4">
            <h4 className="text-sm font-semibold text-slate-900 mb-2 flex items-center gap-2">
              <Video className="h-4 w-4 text-[#6D28D9]" />
              How to find the Channel ID:
            </h4>
            <ol className="text-xs text-slate-600 space-y-1 ml-6 list-decimal">
              <li>Go to the YouTube channel page</li>
              <li>Look at the URL - if it contains <code className="bg-white px-1 rounded">/channel/</code>, copy the ID after it</li>
              <li>Or right-click → View Page Source → search for <code className="bg-white px-1 rounded">channelId</code></li>
              <li>Example: <code className="bg-white px-1 rounded">UCsBjURrPoezykLs9EqgamOA</code></li>
            </ol>
          </div>

          {/* YouTube URL */}
          <div className="space-y-2">
            <Label htmlFor="youtube-url" className="text-sm font-semibold text-slate-900">
              YouTube Channel URL
            </Label>
            <Input
              id="youtube-url"
              type="url"
              placeholder="https://youtube.com/channel/UCsBjURrPoezykLs9EqgamOA"
              value={youtubeUrl}
              onChange={(e) => setYoutubeUrl(e.target.value)}
              className="h-12 rounded-xl border-2 border-purple-100 focus:border-[#6D28D9] text-base"
            />
            <p className="text-xs text-slate-500">
              If URL contains <code className="bg-slate-100 px-1 rounded">/channel/</code>, the ID will be auto-extracted
            </p>
          </div>

          {/* YouTube Channel ID */}
          <div className="space-y-2">
            <Label htmlFor="youtube-channel-id" className="text-sm font-semibold text-slate-900">
              YouTube Channel ID <span className="text-red-500">*</span>
            </Label>
            <div className="relative">
              <Input
                id="youtube-channel-id"
                type="text"
                placeholder="e.g., UCsBjURrPoezykLs9EqgamOA"
                value={youtubeChannelId}
                onChange={(e) => {
                  setYoutubeChannelId(e.target.value);
                  setAutoExtracted(false);
                }}
                required
                className="h-12 rounded-xl border-2 border-purple-100 focus:border-[#6D28D9] text-base font-mono pr-10"
              />
              {autoExtracted && (
                <CheckCircle2 className="absolute right-3 top-1/2 -translate-y-1/2 h-5 w-5 text-green-500" />
              )}
            </div>
            {autoExtracted && (
              <p className="text-xs text-green-600 flex items-center gap-1">
                <CheckCircle2 className="h-3 w-3" />
                Auto-extracted from URL
              </p>
            )}
          </div>

          {/* Channel Name */}
          <div className="space-y-2">
            <Label htmlFor="name" className="text-sm font-semibold text-slate-900">
              Channel Name <span className="text-red-500">*</span>
            </Label>
            <Input
              id="name"
              type="text"
              placeholder="e.g., Fireship, ThePrimeagen, Web Dev Simplified"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              className="h-12 rounded-xl border-2 border-purple-100 focus:border-[#6D28D9] text-base"
            />
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label htmlFor="description" className="text-sm font-semibold text-slate-900">
              Description
            </Label>
            <Textarea
              id="description"
              placeholder="Brief description of the channel..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
              className="rounded-xl border-2 border-purple-100 focus:border-[#6D28D9] resize-none text-base"
            />
          </div>

          {/* Thumbnail URL */}
          <div className="space-y-2">
            <Label htmlFor="thumbnail-url" className="text-sm font-semibold text-slate-900">
              Thumbnail URL
            </Label>
            <Input
              id="thumbnail-url"
              type="url"
              placeholder="https://..."
              value={thumbnailUrl}
              onChange={(e) => setThumbnailUrl(e.target.value)}
              className="h-12 rounded-xl border-2 border-purple-100 focus:border-[#6D28D9] text-base"
            />
            <p className="text-xs text-slate-500">
              Optional: Channel avatar/logo image URL
            </p>
          </div>

          {/* Submit Button */}
          <div className="flex gap-3 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => setOpen(false)}
              disabled={isSubmitting}
              className="flex-1 h-12 rounded-xl border-2 border-purple-200 hover:bg-purple-50 text-base"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isSubmitting}
              className="flex-1 h-12 rounded-xl bg-gradient-to-r from-[#6D28D9] to-[#5B21B6] hover:from-[#5B21B6] hover:to-[#4C1D95] text-base shadow-lg"
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="mr-2 h-5 w-5 animate-spin" />
                  Creating...
                </>
              ) : (
                <>
                  <Plus className="mr-2 h-5 w-5" />
                  Create Channel
                </>
              )}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
