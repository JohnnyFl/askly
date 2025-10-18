'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { MessageSquare, LogOut, Loader2, Search, Sparkles, User } from 'lucide-react';
import { Input } from '@/components/ui/input';
import Link from 'next/link';

type Channel = {
  id: string;
  name: string;
  description: string;
  avatar_url: string;
  youtube_url: string;
  created_at: string;
};

export default function ChannelsPage() {
  const router = useRouter();
  const [channels, setChannels] = useState<Channel[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    const token = localStorage.getItem('auth_token');
    if (!token) {
      router.push('/login');
      return;
    }
    fetchChannels();
  }, [router]);

  const fetchChannels = async () => {
    try {
      const token = localStorage.getItem('auth_token');
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/channels`, {
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

  const handleSignOut = () => {
    localStorage.removeItem('auth_token');
    router.push('/login');
  };

  const filteredChannels = channels.filter(channel =>
    channel.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    channel.description?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[#FDF2F8]">
        <Loader2 className="h-8 w-8 animate-spin text-[#EF4444]" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#FDF2F8]">
      <div className="min-h-screen bg-gradient-to-br from-pink-50/50 to-purple-50/30">
        <div className="max-w-[1400px] mx-auto p-8">
          <header className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 mb-8 px-8 py-5">
            <div className="flex justify-between items-center">
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 bg-gradient-to-br from-[#FB7185] to-[#EF4444] rounded-xl flex items-center justify-center shadow-lg">
                  <MessageSquare className="h-6 w-6 text-white" />
                </div>
                <h1 className="text-2xl font-bold bg-gradient-to-r from-[#FB7185] to-[#EF4444] bg-clip-text text-transparent">Askly</h1>
              </div>
              <div className="flex items-center gap-6">
                <nav className="hidden md:flex gap-6 text-slate-700 font-medium">
                  <Link href="/pricing" className="hover:text-[#EF4444] transition-colors">Pricing</Link>
                </nav>
                <div className="relative hidden md:block">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-slate-400" />
                  <Input
                    type="text"
                    placeholder="Search creators..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-10 w-64 rounded-xl border-2 border-pink-100 focus:border-[#FB7185]"
                  />
                </div>
                <Link href="/account">
                  <Button
                    variant="ghost"
                    className="gap-2 rounded-xl font-medium hover:bg-pink-50 hover:text-[#EF4444]"
                  >
                    <User className="h-4 w-4" />
                    Account
                  </Button>
                </Link>
                <Button
                  variant="ghost"
                  onClick={handleSignOut}
                  className="gap-2 rounded-xl font-medium hover:bg-pink-50 hover:text-[#EF4444]"
                >
                  <LogOut className="h-4 w-4" />
                  Sign Out
                </Button>
              </div>
            </div>
          </header>

          <main>
            <div className="mb-8">
              <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 p-8">
                <div className="flex items-center gap-4">
                  <div className="h-14 w-14 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-2xl flex items-center justify-center shadow-lg">
                    <Sparkles className="h-7 w-7 text-white" />
                  </div>
                  <div>
                    <h2 className="text-4xl font-bold text-slate-900 mb-2">Discover Creators</h2>
                    <p className="text-slate-600 text-lg">Choose a channel and start exploring with the community</p>
                  </div>
                </div>
              </Card>
            </div>

            {filteredChannels.length === 0 ? (
              <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 max-w-md mx-auto">
                <CardHeader className="p-8 text-center">
                  <CardTitle className="text-2xl mb-2">No Creators Found</CardTitle>
                  <CardDescription className="text-base">
                    {searchQuery ? 'Try adjusting your search terms' : 'New creators coming soon! Check back later.'}
                  </CardDescription>
                </CardHeader>
              </Card>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {filteredChannels.map((channel) => (
                  <Card
                    key={channel.id}
                    className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 hover:shadow-2xl hover:scale-[1.02] hover:border-[#FB7185]/50 transition-all cursor-pointer group overflow-hidden"
                    onClick={() => router.push(`/channels/${channel.id}`)}
                  >
                    <CardHeader className="p-6">
                      <div className="flex items-start gap-4 mb-4">
                        <Avatar className="h-20 w-20 ring-4 ring-pink-100 group-hover:ring-[#FB7185]/30 transition-all">
                          <AvatarImage src={channel.avatar_url} alt={channel.name} />
                          <AvatarFallback className="bg-gradient-to-br from-[#FB7185] to-[#EF4444] text-white text-2xl font-bold">
                            {channel.name.substring(0, 2).toUpperCase()}
                          </AvatarFallback>
                        </Avatar>
                        <div className="flex-1 min-w-0">
                          <CardTitle className="text-xl truncate group-hover:text-[#EF4444] transition-colors mb-1">
                            {channel.name}
                          </CardTitle>
                          <div className="flex items-center gap-2 text-xs text-slate-500">
                            <MessageSquare className="h-3 w-3 text-[#FB7185]" />
                            <span>AI-Powered Insights</span>
                          </div>
                        </div>
                      </div>
                      <CardDescription className="line-clamp-3 min-h-[4.5rem] text-slate-600 leading-relaxed">
                        {channel.description || 'Get instant answers about tools, strategies, and insights from this creator'}
                      </CardDescription>
                    </CardHeader>
                    <CardContent className="px-6 pb-6">
                      <Button
                        className="w-full rounded-xl bg-[#EF4444] hover:bg-[#DC2626] font-medium shadow-lg group-hover:bg-gradient-to-r group-hover:from-[#FB7185] group-hover:to-[#EF4444] transition-all"
                      >
                        Start Exploring
                      </Button>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </main>
        </div>
      </div>
    </div>
  );
}
