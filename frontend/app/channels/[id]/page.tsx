'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Textarea } from '@/components/ui/textarea';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { ScrollArea } from '@/components/ui/scroll-area';
import { ArrowLeft, Send, Loader2, MessageSquare, HelpCircle, Clock, Sparkles, User, AlertCircle } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog';
import Link from 'next/link';

type Channel = {
  id: string;
  name: string;
  description: string;
  thumbnail_url: string;
  youtube_url: string;
  youtube_channel_id: string;
  created_at: string;
};

type Question = {
  id: string;
  channel_id: string;
  user_id: string;
  question: string;  // Backend returns 'question' not 'question_text'
  answer: string;     // Backend returns 'answer' not 'answer_text'
  sources?: any[];
  created_at: string;
};

export default function ChannelChatPage() {
  const router = useRouter();
  const params = useParams();
  const channelId = params.id as string;

  const [channel, setChannel] = useState<Channel | null>(null);
  const [faqs, setFaqs] = useState<Question[]>([]);
  const [recentQuestions, setRecentQuestions] = useState<Question[]>([]);
  const [question, setQuestion] = useState('');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [showUpgradeModal, setShowUpgradeModal] = useState(false);
  const [showChannelAccessModal, setShowChannelAccessModal] = useState(false);
  const [usageData, setUsageData] = useState<{ current: number; limit: number; tier: string } | null>(null);

  useEffect(() => {
    const token = localStorage.getItem('auth_token');
    if (!token) {
      router.push('/login');
      return;
    }
    if (channelId) {
      fetchChannelData();
      fetchFAQs();
      fetchRecentQuestions();
      fetchUsageData();
    }
  }, [channelId, router]);

  const fetchUsageData = async () => {
    try {
      const token = localStorage.getItem('auth_token');
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/user/subscription`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) return;

      const data = await response.json();
      setUsageData({
        current: data.questions_used_this_month || 0,
        limit: data.questions_limit || 5,
        tier: data.subscription_tier?.toUpperCase() || 'FREE',
      });
    } catch (error) {
      console.error('Error fetching usage data:', error);
    }
  };

  const fetchChannelData = async () => {
    try {
      const token = localStorage.getItem('auth_token');
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/channels/${channelId}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) throw new Error('Failed to fetch channel');

      const data = await response.json();
      // Backend returns channel directly, not wrapped
      if (!data.id) {
        router.push('/channels');
        return;
      }
      setChannel(data);
    } catch (error) {
      console.error('Error fetching channel:', error);
      router.push('/channels');
    } finally {
      setLoading(false);
    }
  };

  const fetchFAQs = async () => {
    try {
      const token = localStorage.getItem('auth_token');
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/channels/${channelId}/messages`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) throw new Error('Failed to fetch FAQs');

      const data = await response.json();
      // Backend returns messages array
      setFaqs(data.messages || []);
    } catch (error) {
      console.error('Error fetching FAQs:', error);
    }
  };

  const fetchRecentQuestions = async () => {
    try {
      const token = localStorage.getItem('auth_token');
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/channels/${channelId}/messages`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) throw new Error('Failed to fetch messages');

      const data = await response.json();
      setRecentQuestions(data.messages || []);
    } catch (error) {
      console.error('Error fetching recent questions:', error);
    }
  };

  const handleSubmitQuestion = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!question.trim()) return;

    // Check usage limit before submitting
    if (usageData && usageData.current >= usageData.limit) {
      setShowUpgradeModal(true);
      return;
    }

    // Check channel access for free users
    if (usageData && usageData.tier.toLowerCase() === 'free') {
      try {
        const token = localStorage.getItem('auth_token');
        const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/user/subscription`, {
          headers: { 'Authorization': `Bearer ${token}` },
        });

        const userData = await response.json();
        if (userData.selected_channel_id && userData.selected_channel_id !== channelId) {
          setShowChannelAccessModal(true);
          return;
        }
      } catch (error) {
        console.error('Error checking channel access:', error);
      }
    }

    setSubmitting(true);

    try {
      const token = localStorage.getItem('auth_token');
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/channels/${channelId}/ask`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ question: question.trim() }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        // Check if it's a usage limit error
        if (response.status === 429 || errorData.error?.includes('limit')) {
          setShowUpgradeModal(true);
          return;
        }
        throw new Error('Failed to submit question');
      }

      setQuestion('');
      fetchRecentQuestions();
      fetchUsageData(); // Update usage count
    } catch (error) {
      console.error('Error submitting question:', error);
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[#FDF2F8]">
        <Loader2 className="h-8 w-8 animate-spin text-[#EF4444]" />
      </div>
    );
  }

  if (!channel) return null;

  // @ts-ignore
  return (
    <div className="min-h-screen bg-[#FDF2F8]">
      <div className="min-h-screen bg-gradient-to-br from-pink-50/50 to-purple-50/30">
        <div className="max-w-[1400px] mx-auto p-8">
          <header className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 mb-8 px-8 py-5">
            <div className="flex items-center gap-4">
              <Button
                variant="ghost"
                size="icon"
                onClick={() => router.push('/channels')}
                className="shrink-0 rounded-xl hover:bg-pink-50"
              >
                <ArrowLeft className="h-5 w-5" />
              </Button>
              <Avatar className="h-14 w-14 shrink-0 ring-4 ring-pink-100">
                <AvatarImage src={channel.thumbnail_url} alt={channel.name} />
                <AvatarFallback className="bg-gradient-to-br from-[#FB7185] to-[#EF4444] text-white font-bold text-lg">
                  {channel.name.substring(0, 2).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <div className="flex-1 min-w-0">
                <h1 className="text-2xl font-bold text-slate-900 truncate">{channel.name}</h1>
                <p className="text-sm text-slate-600 truncate">{channel.description}</p>
              </div>
              {usageData && (
                <div className="hidden sm:block">
                  <Badge
                    className={`px-3 py-1 ${
                      usageData.current >= usageData.limit
                        ? 'bg-red-100 text-red-700 border-red-200'
                        : usageData.current / usageData.limit >= 0.8
                        ? 'bg-orange-100 text-orange-700 border-orange-200'
                        : 'bg-green-100 text-green-700 border-green-200'
                    }`}
                    variant="outline"
                  >
                    {usageData.current}/{usageData.limit === -1 ? '∞' : usageData.limit} questions
                  </Badge>
                </div>
              )}
              <nav className="hidden md:flex gap-4">
                <Link href="/pricing">
                  <Button variant="ghost" className="rounded-xl font-medium hover:bg-pink-50 hover:text-[#EF4444]">
                    Pricing
                  </Button>
                </Link>
                <Link href="/account">
                  <Button variant="ghost" className="gap-2 rounded-xl font-medium hover:bg-pink-50 hover:text-[#EF4444]">
                    <User className="h-4 w-4" />
                    Account
                  </Button>
                </Link>
              </nav>
            </div>
          </header>

          <main className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2 space-y-6">
              <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50">
                <CardHeader className="p-8 pb-6">
                  <div className="flex items-center gap-3 mb-2">
                    <div className="h-12 w-12 bg-gradient-to-br from-[#EF4444] to-[#DC2626] rounded-2xl flex items-center justify-center shadow-lg">
                      <MessageSquare className="h-6 w-6 text-white" />
                    </div>
                    <div>
                      <CardTitle className="text-2xl">Ask Your Question</CardTitle>
                      <CardDescription className="text-base mt-1">
                        Get instant insights and connect with the community
                      </CardDescription>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="px-8 pb-8">
                  <form onSubmit={handleSubmitQuestion} className="space-y-4">
                    <Textarea
                      placeholder="e.g., What marketing tools does this creator recommend for SaaS startups?"
                      value={question}
                      onChange={(e) => setQuestion(e.target.value)}
                      className="min-h-[140px] resize-none rounded-2xl border-2 border-pink-100 focus:border-[#FB7185] text-base"
                      disabled={submitting}
                    />
                    <Button
                      type="submit"
                      disabled={submitting || !question.trim()}
                      className="w-full sm:w-auto gap-2 rounded-xl bg-[#EF4444] hover:bg-[#DC2626] shadow-lg text-base py-6 px-8"
                    >
                      {submitting ? (
                        <>
                          <Loader2 className="h-5 w-5 animate-spin" />
                          Submitting...
                        </>
                      ) : (
                        <>
                          <Send className="h-5 w-5" />
                          Send Question
                        </>
                      )}
                    </Button>
                  </form>
                </CardContent>
              </Card>

              <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50">
                <CardHeader className="p-8 pb-6">
                  <div className="flex items-center gap-3">
                    <div className="h-12 w-12 bg-gradient-to-br from-[#FB7185] to-[#EF4444] rounded-2xl flex items-center justify-center shadow-lg">
                      <Clock className="h-6 w-6 text-white" />
                    </div>
                    <div>
                      <CardTitle className="text-2xl">Your Discoveries</CardTitle>
                      <CardDescription className="text-base mt-1">
                        Your journey with this creator
                      </CardDescription>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="px-8 pb-8">
                  <ScrollArea className="h-[500px] pr-4">
                    {recentQuestions.length === 0 ? (
                      <div className="text-center py-12 text-slate-500">
                        <Sparkles className="h-16 w-16 mx-auto mb-4 opacity-30 text-[#FB7185]" />
                        <p className="text-lg font-medium">Start Your Journey</p>
                        <p className="text-sm mt-1">Ask your first question above!</p>
                      </div>
                    ) : (
                      <div className="space-y-6">
                        {recentQuestions.map((qa) => (
                          <div key={qa.id} className="space-y-3">
                            <div className="bg-gradient-to-br from-pink-50 to-red-50 rounded-2xl p-5 border-2 border-pink-200/50">
                              <div className="flex items-start gap-2 mb-2">
                                <Badge className="bg-[#EF4444] text-white">You</Badge>
                                <p className="text-xs text-slate-500 mt-0.5">
                                  {new Date(qa.created_at).toLocaleDateString('en-US', {
                                    month: 'short',
                                    day: 'numeric',
                                    hour: '2-digit',
                                    minute: '2-digit',
                                  })}
                                </p>
                              </div>
                              <p className="text-slate-900 leading-relaxed">{qa.question}</p>
                            </div>
                            <div className="bg-gradient-to-br from-purple-50 to-pink-50 rounded-2xl p-5 border-2 border-purple-200/50">
                              <Badge className="bg-[#6D28D9] text-white mb-2">AI Answer</Badge>
                              {/*<div className="text-slate-700 leading-relaxed">{qa.answer}</div>*/}
                              {/* @ts-ignore */}
                              <div className="text-slate-700 leading-relaxed" dangerouslySetInnerHTML={{__html: qa.answer}}></div>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </ScrollArea>
                </CardContent>
              </Card>
            </div>

            {/*<div className="lg:col-span-1">*/}
            {/*  <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50 sticky top-8">*/}
            {/*    <CardHeader className="p-8 pb-6">*/}
            {/*      <div className="flex items-center gap-3 mb-2">*/}
            {/*        <div className="h-12 w-12 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-2xl flex items-center justify-center shadow-lg">*/}
            {/*          <HelpCircle className="h-6 w-6 text-white" />*/}
            {/*        </div>*/}
            {/*        <div>*/}
            {/*          <CardTitle className="text-xl">Community FAQs</CardTitle>*/}
            {/*          <CardDescription className="text-sm mt-1">*/}
            {/*            Popular questions from fans*/}
            {/*          </CardDescription>*/}
            {/*        </div>*/}
            {/*      </div>*/}
            {/*    </CardHeader>*/}
            {/*    <CardContent className="px-8 pb-8">*/}
            {/*      <ScrollArea className="h-[600px] pr-4">*/}
            {/*        {faqs.length === 0 ? (*/}
            {/*          <div className="text-center py-12 text-slate-500">*/}
            {/*            <HelpCircle className="h-16 w-16 mx-auto mb-4 opacity-30 text-[#6D28D9]" />*/}
            {/*            <p className="text-base font-medium">No FAQs Yet</p>*/}
            {/*            <p className="text-sm mt-1">Be the first to ask!</p>*/}
            {/*          </div>*/}
            {/*        ) : (*/}
            {/*          <div className="space-y-5">*/}
            {/*            {faqs.map((faq, index) => (*/}
            {/*              <div key={faq.id}>*/}
            {/*                <div className="bg-gradient-to-br from-purple-50 to-pink-50 rounded-2xl p-5 border-2 border-purple-200/50 hover:border-[#6D28D9]/50 transition-colors">*/}
            {/*                  <div className="flex items-start gap-3 mb-3">*/}
            {/*                    <Badge className="bg-[#6D28D9] text-white shrink-0">*/}
            {/*                      #{index + 1}*/}
            {/*                    </Badge>*/}
            {/*                  </div>*/}
            {/*                  <p className="font-semibold text-slate-900 leading-snug mb-2">*/}
            {/*                    {faq.question}*/}
            {/*                  </p>*/}
            {/*                  <p className="text-sm text-slate-600 leading-relaxed">*/}
            {/*                    {faq.answer}*/}
            {/*                  </p>*/}
            {/*                </div>*/}
            {/*                {index < faqs.length - 1 && <div className="h-5" />}*/}
            {/*              </div>*/}
            {/*            ))}*/}
            {/*          </div>*/}
            {/*        )}*/}
            {/*      </ScrollArea>*/}
            {/*    </CardContent>*/}
            {/*  </Card>*/}
            {/*</div>*/}
          </main>
        </div>
      </div>

      {/* Upgrade Modal */}
      <Dialog open={showUpgradeModal} onOpenChange={setShowUpgradeModal}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <div className="flex items-center gap-3 mb-2">
              <div className="h-12 w-12 bg-gradient-to-br from-[#FB7185] to-[#EF4444] rounded-2xl flex items-center justify-center">
                <AlertCircle className="h-6 w-6 text-white" />
              </div>
              <DialogTitle className="text-2xl">Question Limit Reached</DialogTitle>
            </div>
            <DialogDescription className="text-base pt-2">
              You&#39;ve used {usageData?.current || 0} of {usageData?.limit || 5} questions this month.
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <p className="text-slate-700 mb-6">
              Upgrade to <strong>Pro</strong> for 50 questions/month ($7/mo) or <strong>Premium</strong> for 200 questions/month ($15/mo) to continue exploring creator insights.
            </p>
            <div className="flex gap-3">
              <Button
                onClick={() => router.push('/pricing')}
                className="flex-1 bg-gradient-to-r from-[#FB7185] to-[#EF4444] hover:from-[#EF4444] hover:to-[#DC2626] shadow-lg rounded-xl"
              >
                View Plans
              </Button>
              <Button onClick={() => setShowUpgradeModal(false)} variant="outline" className="flex-1 rounded-xl">
                Close
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* Channel Access Modal (Free Users) */}
      <Dialog open={showChannelAccessModal} onOpenChange={setShowChannelAccessModal}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <div className="flex items-center gap-3 mb-2">
              <div className="h-12 w-12 bg-gradient-to-br from-purple-500 to-purple-600 rounded-2xl flex items-center justify-center">
                <AlertCircle className="h-6 w-6 text-white" />
              </div>
              <DialogTitle className="text-2xl">Channel Access Limited</DialogTitle>
            </div>
            <DialogDescription className="text-base pt-2">
              Free users can only access one channel at a time.
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <p className="text-slate-700 mb-6">
              You&#39;ve already selected a different channel. Upgrade to <strong>Pro</strong> to access 2 channels or <strong>Premium</strong> for unlimited channel access.
            </p>
            <div className="flex gap-3">
              <Button
                onClick={() => router.push('/pricing')}
                className="flex-1 bg-gradient-to-r from-[#6D28D9] to-[#5B21B6] hover:from-[#5B21B6] hover:to-[#4C1D95] shadow-lg rounded-xl"
              >
                Upgrade Plan
              </Button>
              <Button onClick={() => setShowChannelAccessModal(false)} variant="outline" className="flex-1 rounded-xl">
                Go Back
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
