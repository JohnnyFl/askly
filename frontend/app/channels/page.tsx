'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { MessageSquare, LogOut, Loader2, Search, Sparkles, User, TrendingUp, Zap } from 'lucide-react';
import { Input } from '@/components/ui/input';
import Link from 'next/link';
import { FadeInWhenVisible } from '@/components/animations/FadeInWhenVisible';
import { SkeletonCard } from '@/components/animations/SkeletonCard';

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
      <div className="min-h-screen bg-[#FDF2F8]">
        <div className="min-h-screen bg-gradient-to-br from-pink-50/50 to-purple-50/30">
          <div className="max-w-[1400px] mx-auto p-8">
            <header className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 mb-8 px-8 py-5 animate-fade-in-down">
              <div className="flex justify-between items-center">
                <div className="flex items-center gap-3">
                  <div className="h-10 w-10 bg-gradient-to-br from-[#FB7185] to-[#EF4444] rounded-xl flex items-center justify-center shadow-lg animate-float">
                    <MessageSquare className="h-6 w-6 text-white" />
                  </div>
                  <h1 className="text-2xl font-bold bg-gradient-to-r from-[#FB7185] to-[#EF4444] bg-clip-text text-transparent">Askly</h1>
                </div>
                <div className="flex items-center gap-6">
                  <Loader2 className="h-5 w-5 animate-spin text-[#EF4444]" />
                </div>
              </div>
            </header>

            <main>
              <div className="mb-8">
                <div className="h-32 bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 skeleton"></div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {[...Array(6)].map((_, i) => (
                  <SkeletonCard key={i} />
                ))}
              </div>
            </main>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#FDF2F8]">
      <div className="min-h-screen bg-gradient-to-br from-pink-50/50 to-purple-50/30">
        <div className="max-w-[1400px] mx-auto p-8">
          <header className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 mb-8 px-8 py-5 animate-fade-in-down">
            <div className="flex justify-between items-center">
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 bg-gradient-to-br from-[#FB7185] to-[#EF4444] rounded-xl flex items-center justify-center shadow-lg animate-float">
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
                    className="pl-10 w-64 rounded-xl border-2 border-pink-100 focus:border-[#FB7185] transition-all"
                  />
                </div>
                <Link href="/account">
                  <Button
                    variant="ghost"
                    className="gap-2 rounded-xl font-medium hover:bg-pink-50 hover:text-[#EF4444] transition-all"
                  >
                    <User className="h-4 w-4" />
                    Account
                  </Button>
                </Link>
                <Button
                  variant="ghost"
                  onClick={handleSignOut}
                  className="gap-2 rounded-xl font-medium hover:bg-pink-50 hover:text-[#EF4444] transition-all"
                >
                  <LogOut className="h-4 w-4" />
                  Sign Out
                </Button>
              </div>
            </div>
          </header>

          <main>
            <FadeInWhenVisible>
              <div className="mb-8">
                <Card className="bg-gradient-to-br from-white/90 to-pink-50/50 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 p-8 hover-lift overflow-hidden relative group">
                  <div className="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-[#FB7185]/10 to-[#6D28D9]/10 rounded-full blur-3xl group-hover:scale-150 transition-transform duration-700"></div>
                  <div className="relative z-10 flex items-center gap-4">
                    <div className="h-14 w-14 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-2xl flex items-center justify-center shadow-lg animate-float">
                      <Sparkles className="h-7 w-7 text-white" />
                    </div>
                    <div>
                      <h2 className="text-4xl font-bold text-slate-900 mb-2">Discover Creators</h2>
                      <p className="text-slate-600 text-lg">Choose a channel and start exploring with the community</p>
                    </div>
                  </div>
                </Card>
              </div>
            </FadeInWhenVisible>

            {filteredChannels.length === 0 ? (
              <FadeInWhenVisible>
                <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 max-w-md mx-auto hover-lift">
                  <CardHeader className="p-8 text-center">
                    <div className="h-20 w-20 bg-gradient-to-br from-pink-100 to-purple-100 rounded-2xl flex items-center justify-center mx-auto mb-4">
                      <Search className="h-10 w-10 text-[#EF4444]" />
                    </div>
                    <CardTitle className="text-2xl mb-2">No Creators Found</CardTitle>
                    <CardDescription className="text-base">
                      {searchQuery ? 'Try adjusting your search terms' : 'New creators coming soon! Check back later.'}
                    </CardDescription>
                  </CardHeader>
                </Card>
              </FadeInWhenVisible>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {filteredChannels.map((channel, index) => (
                  <FadeInWhenVisible key={channel.id} delay={index * 100}>
                    <Card
                      className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 hover:shadow-2xl hover:scale-[1.02] hover:border-[#FB7185]/50 transition-all duration-300 cursor-pointer group overflow-hidden relative"
                      onClick={() => router.push(`/channels/${channel.id}`)}
                    >
                      <div className="absolute top-0 right-0 w-32 h-32 bg-gradient-to-br from-[#FB7185]/10 to-transparent rounded-full blur-2xl opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>

                      {index < 3 && (
                        <div className="absolute top-4 right-4 z-20">
                          <div className="bg-gradient-to-r from-[#EF4444] to-[#DC2626] text-white px-3 py-1 rounded-full text-xs font-bold flex items-center gap-1 animate-pulse-slow">
                            <TrendingUp className="h-3 w-3" />
                            Trending
                          </div>
                        </div>
                      )}

                      <CardHeader className="p-6 relative z-10">
                        <div className="flex items-start gap-4 mb-4">
                          <Avatar className="h-20 w-20 ring-4 ring-pink-100 group-hover:ring-[#FB7185]/30 group-hover:scale-110 transition-all duration-300">
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
                      <CardContent className="px-6 pb-6 relative z-10">
                        <Button
                          className="w-full rounded-xl bg-gradient-to-r from-[#EF4444] to-[#DC2626] hover:from-[#DC2626] hover:to-[#B91C1C] font-medium shadow-lg group-hover:shadow-xl transition-all btn-ripple group/btn"
                        >
                          <span className="flex items-center gap-2">
                            Start Exploring
                            <Zap className="h-4 w-4 group-hover/btn:rotate-12 transition-transform" />
                          </span>
                        </Button>
                      </CardContent>
                    </Card>
                  </FadeInWhenVisible>
                ))}
              </div>
            )}
          </main>
        </div>
      </div>
    </div>
  );
}
