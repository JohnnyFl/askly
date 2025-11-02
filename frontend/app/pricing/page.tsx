'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { MessageSquare, Check, Zap, Crown, Sparkles } from 'lucide-react';
import Link from 'next/link';
import {
  SUBSCRIPTION_TIERS,
  openPaddleCheckout,
  getMonthlyPrice,
  getAnnualSavings,
  type SubscriptionTier
} from '@/lib/paddle';

const tiers = [
  {
    id: 'FREE' as const,
    name: SUBSCRIPTION_TIERS.FREE.name,
    price: SUBSCRIPTION_TIERS.FREE.price,
    annualPrice: undefined,
    icon: MessageSquare,
    gradient: 'from-slate-400 to-slate-500',
    popular: false,
    features: SUBSCRIPTION_TIERS.FREE.features,
  },
  {
    id: 'PRO' as const,
    name: SUBSCRIPTION_TIERS.PRO.name,
    price: SUBSCRIPTION_TIERS.PRO.price,
    annualPrice: SUBSCRIPTION_TIERS.PRO.annualPrice,
    icon: Zap,
    gradient: 'from-[#FB7185] to-[#EF4444]',
    popular: true,
    features: SUBSCRIPTION_TIERS.PRO.features,
  },
  {
    id: 'PREMIUM' as const,
    name: SUBSCRIPTION_TIERS.PREMIUM.name,
    price: SUBSCRIPTION_TIERS.PREMIUM.price,
    annualPrice: SUBSCRIPTION_TIERS.PREMIUM.annualPrice,
    icon: Crown,
    gradient: 'from-[#6D28D9] to-[#5B21B6]',
    popular: false,
    features: SUBSCRIPTION_TIERS.PREMIUM.features,
  },
];

export default function PricingPage() {
  const router = useRouter();
  const [loading, setLoading] = useState<string | null>(null);
  const [billingCycle, setBillingCycle] = useState<'monthly' | 'annual'>('monthly');

  const handleSubscribe = async (tierId: string) => {
    if (tierId === 'FREE') {
      router.push('/signup');
      return;
    }

    setLoading(tierId);

    try {
      const tier = SUBSCRIPTION_TIERS[tierId as SubscriptionTier];
      const priceId = billingCycle === 'monthly' ? tier.priceIdMonthly : tier.priceIdAnnual;

      if (!priceId) {
        console.error('Price ID not configured for tier:', tierId, billingCycle);
        alert('This plan is not yet configured. Please contact support.');
        setLoading(null);
        return;
      }

      const userEmail = localStorage.getItem('user_email');
      const userId = localStorage.getItem('user_id');

      await openPaddleCheckout(priceId, userEmail || undefined, userId || undefined);
    } catch (error) {
      console.error('Error opening checkout:', error);
      alert('Failed to open checkout. Please try again.');
    } finally {
      setLoading(null);
    }
  };

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
                <Link href="/">
                  <span className="text-2xl font-bold bg-gradient-to-r from-[#FB7185] to-[#EF4444] bg-clip-text text-transparent cursor-pointer">
                    Askly
                  </span>
                </Link>
              </div>
              <nav className="hidden md:flex gap-8 text-slate-700 font-medium">
                <Link href="/" className="hover:text-[#EF4444] transition-colors">Home</Link>
                <a href="/#features" className="hover:text-[#EF4444] transition-colors">Features</a>
                <Link href="/pricing" className="text-[#EF4444]">Pricing</Link>
                <a href="/#community" className="hover:text-[#EF4444] transition-colors">Community</a>
              </nav>
              <div className="flex gap-3">
                <Link href="/login">
                  <Button variant="ghost" className="rounded-xl font-medium hover:text-[#EF4444]">Log In</Button>
                </Link>
                <Link href="/signup">
                  <Button className="rounded-xl bg-[#EF4444] hover:bg-[#DC2626] font-medium shadow-lg">
                    Get Started
                  </Button>
                </Link>
              </div>
            </div>
          </header>

          <main>
            <div className="text-center mb-12">
              <Badge className="bg-gradient-to-r from-[#FB7185] to-[#EF4444] text-white px-5 py-2 text-sm font-semibold mb-4">
                Simple & Transparent Pricing
              </Badge>
              <h1 className="text-5xl font-bold text-slate-900 mb-4">
                Choose Your Plan
              </h1>
              <p className="text-lg text-slate-600 max-w-2xl mx-auto">
                Start free and upgrade as you grow. All plans include access to our AI-powered creator insights.
              </p>

              {/* Billing cycle toggle */}
              <div className="flex items-center justify-center gap-3 mt-8">
                <span className={`text-sm font-medium ${billingCycle === 'monthly' ? 'text-slate-900' : 'text-slate-500'}`}>
                  Monthly
                </span>
                <button
                  onClick={() => setBillingCycle(billingCycle === 'monthly' ? 'annual' : 'monthly')}
                  className={`relative w-14 h-7 rounded-full transition-colors ${
                    billingCycle === 'annual' ? 'bg-gradient-to-r from-[#FB7185] to-[#EF4444]' : 'bg-slate-300'
                  }`}
                >
                  <span
                    className={`absolute top-0.5 left-0.5 w-6 h-6 bg-white rounded-full transition-transform ${
                      billingCycle === 'annual' ? 'translate-x-7' : ''
                    }`}
                  />
                </button>
                <span className={`text-sm font-medium ${billingCycle === 'annual' ? 'text-slate-900' : 'text-slate-500'}`}>
                  Annual
                </span>
                {billingCycle === 'annual' && (
                  <Badge className="bg-green-100 text-green-700 border-green-200">
                    Save up to 20%
                  </Badge>
                )}
              </div>
            </div>

            <div className="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto mb-12">
              {tiers.map((tier) => {
                const Icon = tier.icon;
                const displayPrice = tier.id === 'FREE'
                  ? 0
                  : getMonthlyPrice(tier.id as SubscriptionTier, billingCycle);
                const savings = tier.id !== 'FREE' ? getAnnualSavings(tier.id as SubscriptionTier) : 0;

                return (
                  <Card
                    key={tier.id}
                    className={`bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border transition-all ${
                      tier.popular
                        ? 'border-[#FB7185] ring-4 ring-[#FB7185]/20 scale-105'
                        : 'border-pink-100/50 hover:border-[#FB7185]/50'
                    }`}
                  >
                    <CardHeader className="p-8">
                      {tier.popular && (
                        <Badge className="bg-gradient-to-r from-[#FB7185] to-[#EF4444] text-white mb-4 w-fit">
                          Most Popular
                        </Badge>
                      )}
                      <div className={`h-16 w-16 bg-gradient-to-br ${tier.gradient} rounded-2xl flex items-center justify-center mb-6 shadow-lg`}>
                        <Icon className="h-8 w-8 text-white" />
                      </div>
                      <CardTitle className="text-3xl mb-2">{tier.name}</CardTitle>
                      <div className="mb-2">
                        <span className="text-5xl font-bold text-slate-900">
                          ${displayPrice.toFixed(displayPrice % 1 === 0 ? 0 : 2)}
                        </span>
                        <span className="text-slate-600 ml-2">/month</span>
                      </div>
                      {billingCycle === 'annual' && tier.id !== 'FREE' && savings > 0 && (
                        <Badge className="bg-green-100 text-green-700 border-green-200 mb-2 w-fit">
                          Save ${savings}/year
                        </Badge>
                      )}
                      {billingCycle === 'annual' && tier.annualPrice && (
                        <p className="text-sm text-slate-500 mb-4">
                          ${tier.annualPrice} billed annually
                        </p>
                      )}
                      <CardDescription className="text-base">
                        {tier.id === 'FREE' && 'Perfect for trying out Askly'}
                        {tier.id === 'PRO' && 'For regular creator enthusiasts'}
                        {tier.id === 'PREMIUM' && 'For power users and professionals'}
                      </CardDescription>
                    </CardHeader>
                    <CardContent className="px-8 pb-8">
                      <Button
                        onClick={() => handleSubscribe(tier.id)}
                        disabled={loading === tier.id}
                        className={`w-full rounded-xl font-medium shadow-lg mb-6 ${
                          tier.popular
                            ? 'bg-gradient-to-r from-[#FB7185] to-[#EF4444] hover:from-[#EF4444] hover:to-[#DC2626]'
                            : tier.id === 'PREMIUM'
                            ? 'bg-gradient-to-r from-[#6D28D9] to-[#5B21B6] hover:from-[#5B21B6] hover:to-[#4C1D95]'
                            : 'bg-slate-600 hover:bg-slate-700'
                        }`}
                      >
                        {loading === tier.id ? 'Loading...' : tier.id === 'FREE' ? 'Get Started' : 'Subscribe Now'}
                      </Button>
                      <ul className="space-y-3">
                        {tier.features.map((feature, index) => (
                          <li key={index} className="flex items-start gap-3">
                            <Check className="h-5 w-5 text-[#EF4444] shrink-0 mt-0.5" />
                            <span className="text-slate-700">{feature}</span>
                          </li>
                        ))}
                      </ul>
                    </CardContent>
                  </Card>
                );
              })}
            </div>

            <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 max-w-4xl mx-auto">
              <CardContent className="p-12">
                <div className="text-center">
                  <h2 className="text-3xl font-bold text-slate-900 mb-4">
                    Frequently Asked Questions
                  </h2>
                  <div className="space-y-6 text-left mt-8">
                    <div>
                      <h3 className="font-semibold text-lg text-slate-900 mb-2">
                        Can I change my plan anytime?
                      </h3>
                      <p className="text-slate-600">
                        Yes! You can upgrade or downgrade your plan at any time. Changes take effect immediately.
                      </p>
                    </div>
                    <div>
                      <h3 className="font-semibold text-lg text-slate-900 mb-2">
                        What happens when I reach my question limit?
                      </h3>
                      <p className="text-slate-600">
                        You&#39;ll be prompted to upgrade to continue asking questions. Your usage resets at the start of each billing period.
                      </p>
                    </div>
                    <div>
                      <h3 className="font-semibold text-lg text-slate-900 mb-2">
                        Do you offer refunds?
                      </h3>
                      <p className="text-slate-600">
                        Yes, we offer a 30-day money-back guarantee on all paid plans. Contact support for assistance.
                      </p>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </main>
        </div>
      </div>
    </div>
  );
}
