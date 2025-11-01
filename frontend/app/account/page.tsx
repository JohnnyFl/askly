'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { MessageSquare, Settings, CreditCard, User, LogOut, Crown, Zap } from 'lucide-react';
import Link from 'next/link';
import {
  SUBSCRIPTION_TIERS,
  getPaddleBillingPortalUrl,
  type SubscriptionTier
} from '@/lib/paddle';

export default function AccountPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [user, setUser] = useState<any>(null);
  const [subscription, setSubscription] = useState<SubscriptionTier>('FREE');
  const [usage, setUsage] = useState({ current: 0, limit: 10 });

  useEffect(() => {
    const token = localStorage.getItem('auth_token');
    if (!token) {
      router.push('/login');
      return;
    }

    fetchUserData(token);
  }, [router]);

  const fetchUserData = async (token: string) => {
    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/user/subscription`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch user data');
      }

      const data = await response.json();

      setUser({ email: data.email });
      setSubscription(data.subscription_tier as SubscriptionTier);
      setUsage({
        current: data.questions_used_this_month || 0,
        limit: data.questions_limit || SUBSCRIPTION_TIERS[data.subscription_tier as SubscriptionTier].questionsPerMonth,
      });

      setLoading(false);
    } catch (error) {
      console.error('Error fetching user data:', error);
      // Fallback to localStorage for development/testing
      const email = localStorage.getItem('user_email') || 'user@example.com';
      const tier = (localStorage.getItem('subscription_tier') || 'FREE') as SubscriptionTier;
      setUser({ email });
      setSubscription(tier);
      const tierData = SUBSCRIPTION_TIERS[tier];
      setUsage({
        current: parseInt(localStorage.getItem('question_count') || '0'),
        limit: tierData.questionsPerMonth,
      });
      setLoading(false);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('user_email');
    localStorage.removeItem('subscription_tier');
    localStorage.removeItem('question_count');
    router.push('/');
  };

  const handleManageSubscription = () => {
    if (subscription === 'FREE') {
      router.push('/pricing');
    } else {
      // Open Paddle billing portal
      const portalUrl = getPaddleBillingPortalUrl();
      window.open(portalUrl, '_blank');
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-[#FDF2F8] flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-[#EF4444] mx-auto mb-4"></div>
          <p className="text-slate-600">Loading account...</p>
        </div>
      </div>
    );
  }

  const currentTier = SUBSCRIPTION_TIERS[subscription];
  const usagePercentage = usage.limit === 0
    ? 0
    : (usage.current / usage.limit) * 100;

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
                <Link href="/channels">
                  <span className="text-2xl font-bold bg-gradient-to-r from-[#FB7185] to-[#EF4444] bg-clip-text text-transparent cursor-pointer">
                    Askly
                  </span>
                </Link>
              </div>
              <nav className="hidden md:flex gap-8 text-slate-700 font-medium">
                <Link href="/channels" className="hover:text-[#EF4444] transition-colors">Channels</Link>
                <Link href="/pricing" className="hover:text-[#EF4444] transition-colors">Pricing</Link>
                <Link href="/account" className="text-[#EF4444]">Account</Link>
              </nav>
              <Button
                onClick={handleLogout}
                variant="ghost"
                className="rounded-xl font-medium hover:text-[#EF4444]"
              >
                <LogOut className="h-4 w-4 mr-2" />
                Log Out
              </Button>
            </div>
          </header>

          <main className="max-w-4xl mx-auto">
            <div className="mb-8">
              <h1 className="text-4xl font-bold text-slate-900 mb-2">Account Settings</h1>
              <p className="text-slate-600">Manage your subscription and account preferences</p>
            </div>

            <div className="space-y-6">
              <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50">
                <CardHeader className="p-8">
                  <div className="flex items-center gap-3 mb-2">
                    <User className="h-6 w-6 text-[#EF4444]" />
                    <CardTitle className="text-2xl">Profile Information</CardTitle>
                  </div>
                </CardHeader>
                <CardContent className="px-8 pb-8">
                  <div className="space-y-4">
                    <div>
                      <label className="text-sm font-medium text-slate-600">Email Address</label>
                      <p className="text-lg text-slate-900 mt-1">{user?.email}</p>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50">
                <CardHeader className="p-8">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <CreditCard className="h-6 w-6 text-[#EF4444]" />
                      <div>
                        <CardTitle className="text-2xl">Current Plan</CardTitle>
                        <CardDescription className="mt-1">
                          Your subscription and usage details
                        </CardDescription>
                      </div>
                    </div>
                    <Badge className={`px-4 py-2 text-sm font-semibold ${
                      subscription === 'PREMIUM'
                        ? 'bg-gradient-to-r from-[#6D28D9] to-[#5B21B6]'
                        : subscription === 'PRO'
                        ? 'bg-gradient-to-r from-[#FB7185] to-[#EF4444]'
                        : 'bg-slate-600'
                    }`}>
                      {subscription === 'PREMIUM' && <Crown className="h-4 w-4 mr-2" />}
                      {subscription === 'PRO' && <Zap className="h-4 w-4 mr-2" />}
                      {currentTier.name}
                    </Badge>
                  </div>
                </CardHeader>
                <CardContent className="px-8 pb-8">
                  <div className="space-y-6">
                    <div>
                      <div className="flex justify-between mb-3">
                        <span className="text-sm font-medium text-slate-600">Monthly Questions</span>
                        <span className="text-sm font-semibold text-slate-900">
                          {`${usage.current} / ${usage.limit}`}
                        </span>
                      </div>
                      <div className="w-full bg-slate-200 rounded-full h-3 overflow-hidden">
                        <div
                          className="h-full bg-gradient-to-r from-[#FB7185] to-[#EF4444] transition-all duration-300"
                          style={{ width: `${usagePercentage}%` }}
                        />
                      </div>
                    </div>

                    <div className="bg-gradient-to-r from-pink-50 to-purple-50 rounded-2xl p-6">
                      <h3 className="font-semibold text-slate-900 mb-3">Plan Features</h3>
                      <ul className="space-y-2">
                        {currentTier.features.map((feature, index) => (
                          <li key={index} className="flex items-center gap-2 text-slate-700">
                            <div className="h-1.5 w-1.5 rounded-full bg-[#EF4444]"></div>
                            {feature}
                          </li>
                        ))}
                      </ul>
                    </div>

                    <div className="pt-4">
                      <Button
                        onClick={handleManageSubscription}
                        className="w-full rounded-xl bg-gradient-to-r from-[#FB7185] to-[#EF4444] hover:from-[#EF4444] hover:to-[#DC2626] font-medium shadow-lg"
                      >
                        {subscription === 'FREE' ? 'Upgrade Plan' : 'Manage Subscription'}
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>

              <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50">
                <CardHeader className="p-8">
                  <div className="flex items-center gap-3 mb-2">
                    <Settings className="h-6 w-6 text-[#EF4444]" />
                    <CardTitle className="text-2xl">Account Actions</CardTitle>
                  </div>
                </CardHeader>
                <CardContent className="px-8 pb-8">
                  <div className="space-y-4">
                    <Button
                      onClick={handleLogout}
                      variant="outline"
                      className="w-full rounded-xl border-slate-300 hover:border-[#EF4444] hover:text-[#EF4444]"
                    >
                      <LogOut className="h-4 w-4 mr-2" />
                      Log Out
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </div>
          </main>
        </div>
      </div>
    </div>
  );
}
