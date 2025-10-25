'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { ArrowLeft, Video, FileText, Loader2, Upload as UploadIcon, AlertCircle } from 'lucide-react';
import  FileUpload  from '@/components/admin/FileUpload';
import { TranscriptionList, Transcription } from '@/components/admin/TranscriptionList';
import { FadeInWhenVisible } from '@/components/animations/FadeInWhenVisible';
import { useToast } from '@/hooks/use-toast';

type Channel = {
  id: string;
  name: string;
  description: string;
  thumbnail_url: string;
  youtube_url: string;
  youtube_channel_id: string;
  transcription_count?: number;
};

export default function AdminChannelUploadPage() {
  const params = useParams();
  const router = useRouter();
  const { toast } = useToast();
  const channelId = params.id as string;

  const [channel, setChannel] = useState<Channel | null>(null);
  const [transcriptions, setTranscriptions] = useState<Transcription[]>([]);
  const [loading, setLoading] = useState(true);
  const [transcriptionsLoading, setTranscriptionsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (channelId) {
      fetchChannel();
      fetchTranscriptions();
    }
  }, [channelId]);

  const fetchChannel = async () => {
    try {
      const token = localStorage.getItem('admin_auth_token');
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/admin/channels/${channelId}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to fetch channel');
      }

      const data = await response.json();
      // Backend returns channel directly, not wrapped in { channel: ... }
      setChannel(data);
      setError(null);
    } catch (error) {
      console.error('Error fetching channel:', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to load channel information';
      setError(errorMessage);
      toast({
        title: 'Error',
        description: errorMessage,
        variant: 'destructive',
      });
    } finally {
      setLoading(false);
    }
  };

  const fetchTranscriptions = async () => {
    setTranscriptionsLoading(true);
    try {
      const token = localStorage.getItem('admin_auth_token');
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/admin/channels/${channelId}/transcriptions`,
        {
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        }
      );

      if (!response.ok) {
        throw new Error('Failed to fetch transcriptions');
      }

      const data = await response.json();
      setTranscriptions(data.transcriptions || []);
    } catch (error) {
      console.error('Error fetching transcriptions:', error);
      toast({
        title: 'Error',
        description: 'Failed to load transcriptions',
        variant: 'destructive',
      });
    } finally {
      setTranscriptionsLoading(false);
    }
  };

  const handleUploadComplete = () => {
    toast({
      title: 'Success',
      description: 'Transcription uploaded successfully',
    });
    fetchTranscriptions();
  };

  const handleUploadError = (error: Error) => {
    toast({
      title: 'Upload Failed',
      description: error.message || 'Failed to upload transcription',
      variant: 'destructive',
    });
  };

  const handleDelete = async (transcriptionId: string) => {
    try {
      const token = localStorage.getItem('admin_auth_token');
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/admin/channels/${channelId}/transcriptions/${transcriptionId}`,
        {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        }
      );

      if (!response.ok) {
        throw new Error('Failed to delete transcription');
      }

      toast({
        title: 'Success',
        description: 'Transcription deleted successfully',
      });
      fetchTranscriptions();
    } catch (error) {
      console.error('Error deleting transcription:', error);
      toast({
        title: 'Error',
        description: 'Failed to delete transcription',
        variant: 'destructive',
      });
      throw error;
    }
  };

  const handleDownload = async (transcriptionId: string) => {
    try {
      const token = localStorage.getItem('admin_auth_token');
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/admin/channels/${channelId}/transcriptions/${transcriptionId}/download`,
        {
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        }
      );

      if (!response.ok) {
        throw new Error('Failed to download transcription');
      }

      // Get filename from Content-Disposition header or use default
      const contentDisposition = response.headers.get('Content-Disposition');
      const filenameMatch = contentDisposition?.match(/filename="(.+)"/);
      const filename = filenameMatch ? filenameMatch[1] : `transcription-${transcriptionId}.txt`;

      // Create blob and download
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      toast({
        title: 'Success',
        description: 'Transcription downloaded successfully',
      });
    } catch (error) {
      console.error('Error downloading transcription:', error);
      toast({
        title: 'Error',
        description: 'Failed to download transcription',
        variant: 'destructive',
      });
      throw error;
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-[#6D28D9]" />
      </div>
    );
  }

  if (error || !channel) {
    return (
      <div className="space-y-8">
        {/* Back Button */}
        <Button
          variant="ghost"
          onClick={() => router.push('/admin/channels')}
          className="gap-2 rounded-xl hover:bg-purple-50 hover:text-[#6D28D9]"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Channels
        </Button>

        {/* Error Card */}
        <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-red-200/50 p-8">
          <div className="flex flex-col items-center justify-center text-center py-12">
            <div className="h-20 w-20 bg-red-100 rounded-full flex items-center justify-center mb-6">
              <AlertCircle className="h-10 w-10 text-red-600" />
            </div>
            <h2 className="text-2xl font-bold text-slate-900 mb-2">Channel Not Found</h2>
            <p className="text-slate-600 mb-6 max-w-md">
              {error || 'The channel you are looking for could not be found or you do not have permission to access it.'}
            </p>
            <div className="flex gap-3">
              <Button
                onClick={() => fetchChannel()}
                variant="outline"
                className="gap-2 rounded-xl border-2 border-purple-200 hover:bg-purple-50"
              >
                <Loader2 className="h-4 w-4" />
                Try Again
              </Button>
              <Button
                onClick={() => router.push('/admin/channels')}
                className="gap-2 rounded-xl bg-gradient-to-r from-[#6D28D9] to-[#5B21B6] hover:from-[#5B21B6] hover:to-[#4C1D95]"
              >
                <ArrowLeft className="h-4 w-4" />
                Back to Channels
              </Button>
            </div>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Back Button */}
      <Button
        variant="ghost"
        onClick={() => router.push('/admin/channels')}
        className="gap-2 rounded-xl hover:bg-purple-50 hover:text-[#6D28D9]"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Channels
      </Button>

      {/* Channel Header */}
      <FadeInWhenVisible>
        <Card className="bg-gradient-to-br from-white/90 to-purple-50/50 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50 p-8 hover-lift overflow-hidden relative group">
          <div className="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-[#6D28D9]/10 to-[#5B21B6]/10 rounded-full blur-3xl group-hover:scale-150 transition-transform duration-700"></div>
          <div className="relative z-10">
            <div className="flex items-start gap-6">
              <Avatar className="h-24 w-24 ring-4 ring-purple-200 shadow-lg">
                <AvatarImage src={channel.thumbnail_url} alt={channel.name} />
                <AvatarFallback className="bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] text-white text-3xl font-bold">
                  {channel.name.substring(0, 2).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <div className="flex-1">
                <div className="flex items-start justify-between mb-3">
                  <div>
                    <h1 className="text-4xl font-bold text-slate-900 mb-2">{channel.name}</h1>
                    <p className="text-slate-600 text-lg">{channel.description}</p>
                  </div>
                  <a
                    href={channel.youtube_url}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <Button
                      variant="outline"
                      className="gap-2 rounded-xl border-2 border-purple-200 hover:bg-purple-50"
                    >
                      <Video className="h-4 w-4 text-[#6D28D9]" />
                      View Channel
                    </Button>
                  </a>
                </div>
                <div className="flex gap-3">
                  <Badge className="bg-gradient-to-r from-[#6D28D9] to-[#5B21B6] gap-1">
                    <FileText className="h-3 w-3" />
                    {channel.transcription_count || 0} Transcriptions
                  </Badge>
                </div>
              </div>
            </div>
          </div>
        </Card>
      </FadeInWhenVisible>

      {/* File Upload Section */}
      <FadeInWhenVisible delay={200}>
        <div>
          <div className="flex items-center gap-3 mb-6">
            <div className="h-10 w-10 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-xl flex items-center justify-center shadow-lg">
              <UploadIcon className="h-5 w-5 text-white" />
            </div>
            <h2 className="text-2xl font-bold text-slate-900">Upload New Content</h2>
          </div>
          <FileUpload
            endpoint={`${process.env.NEXT_PUBLIC_API_URL}/admin/channels/${channelId}/transcriptions`}
            channelId={channelId}
            onUploadComplete={handleUploadComplete}
            onUploadError={handleUploadError}
            authTokenKey="admin_auth_token"
            allowedFileTypes={['.json', '.txt', '.csv', '.pdf', '.vtt', '.srt']}
          />
        </div>
      </FadeInWhenVisible>

      {/* Transcriptions List */}
      <FadeInWhenVisible delay={400}>
        <TranscriptionList
          channelId={channelId}
          transcriptions={transcriptions}
          loading={transcriptionsLoading}
          onDelete={handleDelete}
          onDownload={handleDownload}
        />
      </FadeInWhenVisible>
    </div>
  );
}