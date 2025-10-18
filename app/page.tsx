'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { MessageSquare, Lightbulb, Zap, Sparkles, ArrowRight, Play, Users, Heart } from 'lucide-react';
import Link from 'next/link';

export default function Home() {
  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem('auth_token');
    if (token) {
      router.push('/channels');
    }
  }, [router]);

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
                <span className="text-2xl font-bold bg-gradient-to-r from-[#FB7185] to-[#EF4444] bg-clip-text text-transparent">Askly</span>
              </div>
              <nav className="hidden md:flex gap-8 text-slate-700 font-medium">
                <a href="#" className="hover:text-[#EF4444] transition-colors">Home</a>
                <a href="#features" className="hover:text-[#EF4444] transition-colors">Features</a>
                <Link href="/pricing" className="hover:text-[#EF4444] transition-colors">Pricing</Link>
                <a href="#community" className="hover:text-[#EF4444] transition-colors">Community</a>
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
            <section className="grid lg:grid-cols-2 gap-8 mb-8">
              <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 p-10">
                <div className="space-y-6">
                  <div className="inline-flex items-center gap-2 bg-gradient-to-r from-[#FB7185]/20 to-[#EF4444]/20 text-[#EF4444] px-5 py-2.5 rounded-full text-sm font-semibold shadow-sm border border-[#FB7185]/30">
                    <Sparkles className="h-4 w-4" />
                    Connect with Your Favorite Creators
                  </div>

                  <h1 className="text-5xl font-bold text-slate-900 leading-tight">
                    Learn from Creators You Love
                  </h1>

                  <p className="text-lg text-slate-600 leading-relaxed">
                    Join a vibrant community exploring startup wisdom, marketing tips, and creative tools. Chat with AI trained on entire channels and get the insights you need instantly.
                  </p>

                  <div className="flex gap-4 pt-4">
                    <Link href="/signup">
                      <Button size="lg" className="gap-2 text-lg px-8 rounded-xl bg-[#EF4444] hover:bg-[#DC2626] shadow-lg">
                        Start Exploring <ArrowRight className="h-5 w-5" />
                      </Button>
                    </Link>
                    <Link href="#community">
                      <Button size="lg" variant="outline" className="gap-2 text-lg px-8 rounded-xl border-2 border-[#6D28D9] text-[#6D28D9] hover:bg-[#6D28D9]/10">
                        <Users className="h-5 w-5" />
                        Join Community
                      </Button>
                    </Link>
                  </div>

                  <div className="flex items-center gap-8 pt-6 text-sm text-slate-600">
                    <div>
                      <div className="text-3xl font-bold bg-gradient-to-r from-[#EF4444] to-[#FB7185] bg-clip-text text-transparent">10K+</div>
                      <div>Active Users</div>
                    </div>
                    <div className="h-12 w-px bg-pink-200"></div>
                    <div>
                      <div className="text-3xl font-bold bg-gradient-to-r from-[#6D28D9] to-[#FB7185] bg-clip-text text-transparent">50+</div>
                      <div>Creator Channels</div>
                    </div>
                  </div>
                </div>
              </Card>

              <div className="grid grid-rows-2 gap-6 h-full">
                <Card className="bg-gradient-to-br from-[#EF4444] to-[#DC2626] rounded-3xl shadow-xl border border-white/20 p-8 text-white flex flex-col">
                  <div className="flex items-start justify-between mb-4">
                    <div className="h-14 w-14 bg-white/20 backdrop-blur-sm rounded-2xl flex items-center justify-center">
                      <Lightbulb className="h-8 w-8" />
                    </div>
                    <div className="text-right">
                      <div className="text-3xl font-bold">Instant</div>
                      <div className="text-white/80">Insights</div>
                    </div>
                  </div>
                  <h3 className="text-2xl font-bold mb-3">Ask Any Question</h3>
                  <p className="text-white/90 leading-relaxed">
                    "What marketing tools does this creator recommend for SaaS startups?" Get comprehensive answers instantly from your favorite channels.
                  </p>
                </Card>

                <Card className="bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-3xl shadow-xl border border-white/20 p-8 text-white flex flex-col">
                  <div className="flex items-start justify-between mb-4">
                    <div className="h-14 w-14 bg-white/20 backdrop-blur-sm rounded-2xl flex items-center justify-center">
                      <Heart className="h-8 w-8" />
                    </div>
                    <div className="text-right">
                      <div className="text-3xl font-bold">100%</div>
                      <div className="text-white/80">Creator Love</div>
                    </div>
                  </div>
                  <h3 className="text-2xl font-bold mb-3">Support Your Favorites</h3>
                  <p className="text-white/90 leading-relaxed">
                    Every question helps creators understand what resonates with their community. Be part of the conversation.
                  </p>
                </Card>
              </div>
            </section>

            <section id="features" className="mb-8">
              <div className="text-center mb-8">
                <h2 className="text-4xl font-bold text-slate-900 mb-3">Why Creators & Fans Love Askly</h2>
                <p className="text-slate-600 text-lg">Join the community discovering insights together</p>
              </div>

              <div className="grid md:grid-cols-3 gap-6">
                <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 hover:shadow-2xl hover:border-[#FB7185]/50 transition-all">
                  <CardHeader className="p-8">
                    <div className="h-16 w-16 bg-gradient-to-br from-[#FB7185] to-[#EF4444] rounded-2xl flex items-center justify-center mb-6 shadow-lg">
                      <MessageSquare className="h-8 w-8 text-white" />
                    </div>
                    <CardTitle className="text-2xl mb-3">Skip the Scrolling</CardTitle>
                    <CardDescription className="text-base leading-relaxed text-slate-600">
                      Want to know what tools your favorite creator recommends? Just ask. Get instant answers without watching every video.
                    </CardDescription>
                  </CardHeader>
                </Card>

                <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50 hover:shadow-2xl hover:border-[#6D28D9]/50 transition-all">
                  <CardHeader className="p-8">
                    <div className="h-16 w-16 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-2xl flex items-center justify-center mb-6 shadow-lg">
                      <Users className="h-8 w-8 text-white" />
                    </div>
                    <CardTitle className="text-2xl mb-3">Community Powered</CardTitle>
                    <CardDescription className="text-base leading-relaxed text-slate-600">
                      See what questions others are asking. Share insights. Build connections with fans who love the same creators you do.
                    </CardDescription>
                  </CardHeader>
                </Card>

                <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 hover:shadow-2xl hover:border-[#FB7185]/50 transition-all">
                  <CardHeader className="p-8">
                    <div className="h-16 w-16 bg-gradient-to-br from-[#FB7185] via-[#EF4444] to-[#6D28D9] rounded-2xl flex items-center justify-center mb-6 shadow-lg">
                      <Zap className="h-8 w-8 text-white" />
                    </div>
                    <CardTitle className="text-2xl mb-3">Your Knowledge Hub</CardTitle>
                    <CardDescription className="text-base leading-relaxed text-slate-600">
                      Building something new? All your questions and discoveries are saved. Reference them anytime, share with friends.
                    </CardDescription>
                  </CardHeader>
                </Card>
              </div>
            </section>

            <section id="community" className="mb-8">
              <Card className="bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-3xl shadow-xl border border-white/10 p-12 text-white">
                <div className="text-center max-w-3xl mx-auto space-y-6">
                  <div className="inline-flex items-center gap-2 bg-white/10 backdrop-blur-sm px-5 py-2.5 rounded-full text-sm font-semibold">
                    <Heart className="h-4 w-4" />
                    Join the Creator Community
                  </div>
                  <h2 className="text-4xl font-bold">Ready to Connect?</h2>
                  <p className="text-purple-100 text-lg leading-relaxed">
                    Thousands of fans are already discovering insights from their favorite creators. Start exploring today and be part of something special.
                  </p>
                  <Link href="/signup">
                    <Button size="lg" className="gap-2 text-lg px-8 rounded-xl bg-[#EF4444] hover:bg-[#DC2626] text-white shadow-lg mt-6">
                      Start Your Journey <ArrowRight className="h-5 w-5" />
                    </Button>
                  </Link>
                </div>
              </Card>
            </section>
          </main>
        </div>
      </div>
    </div>
  );
}
