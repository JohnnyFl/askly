'use client';

import { useState } from 'react';
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
import { Loader2, Plus, Video } from 'lucide-react';

interface AddChannelDialogProps {
  onSuccess?: () => void;
  onError?: (error: Error) => void;
}

export function AddChannelDialog({ onSuccess, onError }: AddChannelDialogProps) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [thumbnailUrl, setThumbnailUrl] = useState('');
  const [youtubeUrl, setYoutubeUrl] = useState('');
  const [youtubeChannelId, setYoutubeChannelId] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

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

          {/* YouTube Channel ID */}
          <div className="space-y-2">
            <Label htmlFor="youtube-channel-id" className="text-sm font-semibold text-slate-900">
              YouTube Channel ID <span className="text-red-500">*</span>
            </Label>
            <Input
              id="youtube-channel-id"
              type="text"
              placeholder="e.g., UCsBjURrPoezykLs9EqgamOA"
              value={youtubeChannelId}
              onChange={(e) => setYoutubeChannelId(e.target.value)}
              required
              className="h-12 rounded-xl border-2 border-purple-100 focus:border-[#6D28D9] text-base font-mono"
            />
            <p className="text-xs text-slate-500">
              Find this in the channel URL after &quot;@username&quot; or in the channel page source
            </p>
          </div>

          {/* YouTube URL */}
          <div className="space-y-2">
            <Label htmlFor="youtube-url" className="text-sm font-semibold text-slate-900">
              YouTube Channel URL
            </Label>
            <Input
              id="youtube-url"
              type="url"
              placeholder="https://youtube.com/@channelname"
              value={youtubeUrl}
              onChange={(e) => setYoutubeUrl(e.target.value)}
              className="h-12 rounded-xl border-2 border-purple-100 focus:border-[#6D28D9] text-base"
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
