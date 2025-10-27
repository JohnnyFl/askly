'use client';

import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import {
  Search,
  Video,
  Upload,
  Loader2,
  FileText,
  ArrowRight,
  Calendar,
  Hash
} from 'lucide-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { FadeInWhenVisible } from '@/components/animations/FadeInWhenVisible';
import { AddChannelDialog } from '@/components/admin/AddChannelDialog';
import { useToast } from '@/hooks/use-toast';

type Channel = {
  id: string;
  name: string;
  description: string;
  thumbnail_url: string;
  youtube_url: string;
  youtube_channel_id: string;
  created_at: string;
  transcription_count?: number;
  status?: 'active' | 'inactive';
};

export default function AdminChannelsPage() {
  const router = useRouter();
  const { toast } = useToast();
  const [channels, setChannels] = useState<Channel[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    fetchChannels();
  }, []);

  const fetchChannels = async () => {
    try {
      const token = localStorage.getItem('admin_auth_token');
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/admin/channels`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) throw new Error('Failed to fetch channels');

      const data = await response.json();
      setChannels(data.channels || []);
    } catch (error) {
      console.error('Error fetching channels:', error);
    } finally {
      setLoading(false);
    }
  };

  const filteredChannels = channels.filter(channel =>
    channel.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    channel.description?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleAddSuccess = () => {
    toast({
      title: 'Success',
      description: 'Channel created successfully',
    });
    fetchChannels();
  };

  const handleAddError = (error: Error) => {
    toast({
      title: 'Error',
      description: error.message || 'Failed to create channel',
      variant: 'destructive',
    });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-[#6D28D9]" />
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <FadeInWhenVisible>
        <Card className="bg-gradient-to-br from-white/90 to-purple-50/50 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50 p-8 hover-lift overflow-hidden relative group">
          <div className="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-[#6D28D9]/10 to-[#5B21B6]/10 rounded-full blur-3xl group-hover:scale-150 transition-transform duration-700"></div>
          <div className="relative z-10 flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div className="h-14 w-14 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-2xl flex items-center justify-center shadow-lg animate-float">
                <Video className="h-7 w-7 text-white" />
              </div>
              <div>
                <h2 className="text-4xl font-bold text-slate-900 mb-2">Channel Management</h2>
                <p className="text-slate-600 text-lg">Manage YouTube channels and upload content</p>
              </div>
            </div>
          </div>
        </Card>
      </FadeInWhenVisible>

      {/* Search Bar and Add Button */}
      <div className="flex items-center gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-slate-400" />
          <Input
            type="text"
            placeholder="Search channels..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-12 h-14 rounded-2xl border-2 border-purple-100 focus:border-[#6D28D9] text-base bg-white/90 backdrop-blur-md shadow-lg"
          />
        </div>
        <AddChannelDialog onSuccess={handleAddSuccess} onError={handleAddError} />
      </div>

      {/* Channels Grid */}
      {filteredChannels.length === 0 ? (
        <FadeInWhenVisible>
          <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50 max-w-md mx-auto hover-lift">
            <CardHeader className="p-8 text-center">
              <div className="h-20 w-20 bg-gradient-to-br from-purple-100 to-indigo-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
                <Search className="h-10 w-10 text-[#6D28D9]" />
              </div>
              <CardTitle className="text-2xl mb-2">No Channels Found</CardTitle>
              <CardDescription className="text-base">
                {searchQuery ? 'Try adjusting your search terms' : 'No channels available yet'}
              </CardDescription>
            </CardHeader>
          </Card>
        </FadeInWhenVisible>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredChannels.map((channel, index) => (
            <FadeInWhenVisible key={channel.id} delay={index * 100}>
              <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50 hover:shadow-2xl hover:border-[#6D28D9]/50 transition-all hover-lift overflow-hidden relative group">
                <div className="absolute top-0 right-0 w-32 h-32 bg-gradient-to-br from-[#6D28D9]/10 to-transparent rounded-full blur-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>

                <CardHeader className="p-6 relative z-10">
                  <div className="flex items-start gap-4 mb-4">
                    <Avatar className="h-16 w-16 ring-4 ring-purple-100 group-hover:ring-[#6D28D9]/30 group-hover:scale-110 transition-all duration-300">
                      <AvatarImage src={channel.thumbnail_url} alt={channel.name} />
                      <AvatarFallback className="bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] text-white text-xl font-bold">
                        {channel.name.substring(0, 2).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div className="flex-1 min-w-0">
                      <CardTitle className="text-lg truncate group-hover:text-[#6D28D9] transition-colors mb-2">
                        {channel.name}
                      </CardTitle>
                      <div className="flex flex-wrap gap-2">
                        <Badge
                          variant={channel.status === 'active' ? 'default' : 'secondary'}
                          className={`text-xs ${
                            channel.status === 'active'
                              ? 'bg-green-500 hover:bg-green-600'
                              : 'bg-slate-400'
                          }`}
                        >
                          {channel.status === 'active' ? 'Active' : 'Inactive'}
                        </Badge>
                        <Badge variant="outline" className="text-xs gap-1">
                          <FileText className="h-3 w-3" />
                          {channel.transcription_count || 0} files
                        </Badge>
                      </div>
                    </div>
                  </div>
                  <CardDescription className="line-clamp-2 min-h-[3rem] text-slate-600 leading-relaxed">
                    {channel.description || 'No description available'}
                  </CardDescription>
                </CardHeader>

                <CardContent className="px-6 pb-6 pt-0 relative z-10">
                  <div className="space-y-3">
                    <div className="flex items-center gap-2 text-sm text-slate-500">
                      <Calendar className="h-4 w-4" />
                      <span>Added {new Date(channel.created_at).toLocaleDateString()}</span>
                    </div>

                    <div className="grid grid-cols-2 gap-2">
                      <Link href={`/admin/channels/${channel.id}`} className="block">
                        <Button
                          className="w-full rounded-xl bg-gradient-to-r from-[#6D28D9] to-[#5B21B6] hover:from-[#5B21B6] hover:to-[#4C1D95] font-medium shadow-lg group/btn"
                        >
                          <span className="flex items-center gap-2">
                            <Upload className="h-4 w-4" />
                            Manage
                          </span>
                        </Button>
                      </Link>
                      <Link href={channel.youtube_url} target="_blank" rel="noopener noreferrer">
                        <Button
                          variant="outline"
                          className="w-full rounded-xl border-2 border-purple-200 hover:bg-purple-50 font-medium group/btn"
                        >
                          <span className="flex items-center gap-2 text-[#6D28D9]">
                            <Video className="h-4 w-4" />
                            View
                          </span>
                        </Button>
                      </Link>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </FadeInWhenVisible>
          ))}
        </div>
      )}
    </div>
  );
}