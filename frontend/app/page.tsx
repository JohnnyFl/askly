'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { MessageSquare, Lightbulb, Zap, Sparkles, ArrowRight, Users, Heart, TrendingUp, Star } from 'lucide-react';
import Link from 'next/link';
import { AnimatedCounter } from '@/components/animations/AnimatedCounter';
import { TypingAnimation } from '@/components/animations/TypingAnimation';
import { FadeInWhenVisible } from '@/components/animations/FadeInWhenVisible';

export default function Home() {
  const router = useRouter();

  useEffect(() => {
    const token = localStorage.getItem('auth_token');
    if (token) {
      router.push('/channels');
    }
  }, [router]);

  return (
    <div className="min-h-screen bg-[#FDF2F8] relative overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-br from-pink-50/50 to-purple-50/30 animate-gradient-y"></div>

      <div className="relative z-10">
        <div className="max-w-[1400px] mx-auto p-8">
          <header className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 mb-8 px-8 py-5 animate-fade-in-down">
            <div className="flex justify-between items-center">
              <div className="flex items-center gap-3">
                <div className="h-10 w-10 bg-gradient-to-br from-[#FB7185] to-[#EF4444] rounded-xl flex items-center justify-center shadow-lg animate-float">
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
                  <Button variant="ghost" className="rounded-xl font-medium hover:text-[#EF4444] transition-all">Log In</Button>
                </Link>
                <Link href="/signup">
                  <Button className="rounded-xl bg-[#EF4444] hover:bg-[#DC2626] font-medium shadow-lg hover-lift btn-ripple transition-all">
                    Get Started
                  </Button>
                </Link>
              </div>
            </div>
          </header>

          <main>
            <section className="grid lg:grid-cols-2 gap-8 mb-8">
              <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 p-10 animate-fade-in-up hover-lift">
                <div className="space-y-6">
                  <div className="inline-flex items-center gap-2 bg-gradient-to-r from-[#FB7185]/20 to-[#EF4444]/20 text-[#EF4444] px-5 py-2.5 rounded-full text-sm font-semibold shadow-sm border border-[#FB7185]/30 animate-bounce-in">
                    <Sparkles className="h-4 w-4 animate-pulse-slow" />
                    Connect with Your Favorite Creators
                  </div>

                  <h1 className="text-5xl font-bold text-slate-900 leading-tight min-h-[140px] flex items-center">
                    <TypingAnimation
                      texts={[
                        'Learn from Creators You Love',
                        'Discover Marketing Secrets',
                        'Find the Best Tools Fast',
                        'Get Startup Wisdom Instantly'
                      ]}
                      className="text-gradient"
                    />
                  </h1>

                  <p className="text-lg text-slate-600 leading-relaxed animate-fade-in animation-delay-400">
                    Join a vibrant community exploring startup wisdom, marketing tips, and creative tools. Chat with AI trained on entire channels and get the insights you need instantly.
                  </p>

                  <div className="flex gap-4 pt-4 animate-fade-in animation-delay-600">
                    <Link href="/signup">
                      <Button size="lg" className="gap-2 text-lg px-8 rounded-xl bg-gradient-to-r from-[#EF4444] to-[#DC2626] hover:from-[#DC2626] hover:to-[#B91C1C] shadow-lg hover-lift btn-ripple transition-all group">
                        Start Exploring <ArrowRight className="h-5 w-5 group-hover:translate-x-1 transition-transform" />
                      </Button>
                    </Link>
                    <Link href="#community">
                      <Button size="lg" variant="outline" className="gap-2 text-lg px-8 rounded-xl border-2 border-[#6D28D9] text-[#6D28D9] hover:bg-[#6D28D9]/10 hover-lift transition-all">
                        <Users className="h-5 w-5" />
                        Join Community
                      </Button>
                    </Link>
                  </div>

                  <div className="flex items-center gap-8 pt-6 text-sm text-slate-600 animate-fade-in animation-delay-800">
                    <div>
                      <AnimatedCounter
                        end={10000}
                        suffix="+"
                        className="text-3xl font-bold bg-gradient-to-r from-[#EF4444] to-[#FB7185] bg-clip-text text-transparent"
                      />
                      <div>Active Users</div>
                    </div>
                    <div className="h-12 w-px bg-pink-200"></div>
                    <div>
                      <AnimatedCounter
                        end={50}
                        suffix="+"
                        className="text-3xl font-bold bg-gradient-to-r from-[#6D28D9] to-[#FB7185] bg-clip-text text-transparent"
                      />
                      <div>Creator Channels</div>
                    </div>
                    <div className="h-12 w-px bg-pink-200"></div>
                    <div>
                      <AnimatedCounter
                        end={98}
                        suffix="%"
                        className="text-3xl font-bold bg-gradient-to-r from-[#10B981] to-[#059669] bg-clip-text text-transparent"
                      />
                      <div>Satisfaction</div>
                    </div>
                  </div>
                </div>
              </Card>

              <div className="grid grid-rows-2 gap-6 h-full">
                <Card className="bg-gradient-to-br from-[#EF4444] to-[#DC2626] rounded-3xl shadow-xl border border-white/20 p-8 text-white flex flex-col animate-fade-in animation-delay-200 hover-lift overflow-hidden relative group">
                  <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
                  <div className="relative z-10">
                    <div className="flex items-start justify-between mb-4">
                      <div className="h-14 w-14 bg-white/20 backdrop-blur-sm rounded-2xl flex items-center justify-center animate-float">
                        <Lightbulb className="h-8 w-8" />
                      </div>
                      <div className="text-right">
                        <div className="text-3xl font-bold animate-bounce-in animation-delay-400">Instant</div>
                        <div className="text-white/80">Insights</div>
                      </div>
                    </div>
                    <h3 className="text-2xl font-bold mb-3">Ask Any Question</h3>
                    <p className="text-white/90 leading-relaxed">
                      "What marketing tools does this creator recommend for SaaS startups?" Get comprehensive answers instantly from your favorite channels.
                    </p>
                  </div>
                </Card>

                <Card className="bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-3xl shadow-xl border border-white/20 p-8 text-white flex flex-col animate-fade-in animation-delay-400 hover-lift overflow-hidden relative group">
                  <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
                  <div className="relative z-10">
                    <div className="flex items-start justify-between mb-4">
                      <div className="h-14 w-14 bg-white/20 backdrop-blur-sm rounded-2xl flex items-center justify-center animate-float animation-delay-200">
                        <Heart className="h-8 w-8" />
                      </div>
                      <div className="text-right">
                        <div className="text-3xl font-bold animate-bounce-in animation-delay-600">100%</div>
                        <div className="text-white/80">Creator Love</div>
                      </div>
                    </div>
                    <h3 className="text-2xl font-bold mb-3">Support Your Favorites</h3>
                    <p className="text-white/90 leading-relaxed">
                      Every question helps creators understand what resonates with their community. Be part of the conversation.
                    </p>
                  </div>
                </Card>
              </div>
            </section>

            <FadeInWhenVisible>
              <section id="features" className="mb-8">
                <div className="text-center mb-8">
                  <h2 className="text-4xl font-bold text-slate-900 mb-3">Why Creators & Fans Love Askly</h2>
                  <p className="text-slate-600 text-lg">Join the community discovering insights together</p>
                </div>

                <div className="grid md:grid-cols-3 gap-6">
                  <FadeInWhenVisible delay={100}>
                    <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 hover:shadow-2xl hover:border-[#FB7185]/50 transition-all hover-lift group">
                      <CardHeader className="p-8">
                        <div className="h-16 w-16 bg-gradient-to-br from-[#FB7185] to-[#EF4444] rounded-2xl flex items-center justify-center mb-6 shadow-lg group-hover:scale-110 transition-transform">
                          <MessageSquare className="h-8 w-8 text-white" />
                        </div>
                        <CardTitle className="text-2xl mb-3 group-hover:text-[#EF4444] transition-colors">Skip the Scrolling</CardTitle>
                        <CardDescription className="text-base leading-relaxed text-slate-600">
                          Want to know what tools your favorite creator recommends? Just ask. Get instant answers without watching every video.
                        </CardDescription>
                      </CardHeader>
                    </Card>
                  </FadeInWhenVisible>

                  <FadeInWhenVisible delay={200}>
                    <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50 hover:shadow-2xl hover:border-[#6D28D9]/50 transition-all hover-lift group">
                      <CardHeader className="p-8">
                        <div className="h-16 w-16 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-2xl flex items-center justify-center mb-6 shadow-lg group-hover:scale-110 transition-transform">
                          <Users className="h-8 w-8 text-white" />
                        </div>
                        <CardTitle className="text-2xl mb-3 group-hover:text-[#6D28D9] transition-colors">Community Powered</CardTitle>
                        <CardDescription className="text-base leading-relaxed text-slate-600">
                          See what questions others are asking. Share insights. Build connections with fans who love the same creators you do.
                        </CardDescription>
                      </CardHeader>
                    </Card>
                  </FadeInWhenVisible>

                  <FadeInWhenVisible delay={300}>
                    <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 hover:shadow-2xl hover:border-[#FB7185]/50 transition-all hover-lift group">
                      <CardHeader className="p-8">
                        <div className="h-16 w-16 bg-gradient-to-br from-[#FB7185] via-[#EF4444] to-[#6D28D9] rounded-2xl flex items-center justify-center mb-6 shadow-lg group-hover:scale-110 transition-transform">
                          <Zap className="h-8 w-8 text-white" />
                        </div>
                        <CardTitle className="text-2xl mb-3 group-hover:text-[#EF4444] transition-colors">Your Knowledge Hub</CardTitle>
                        <CardDescription className="text-base leading-relaxed text-slate-600">
                          Building something new? All your questions and discoveries are saved. Reference them anytime, share with friends.
                        </CardDescription>
                      </CardHeader>
                    </Card>
                  </FadeInWhenVisible>
                </div>
              </section>
            </FadeInWhenVisible>

            <FadeInWhenVisible>
              <section className="mb-8">
                <Card className="bg-gradient-to-br from-white/90 to-pink-50/50 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 p-12 hover-lift">
                  <div className="grid md:grid-cols-3 gap-8 text-center">
                    <div className="space-y-2">
                      <div className="inline-flex items-center justify-center h-16 w-16 bg-gradient-to-br from-[#FB7185] to-[#EF4444] rounded-2xl mx-auto mb-4 animate-float">
                        <TrendingUp className="h-8 w-8 text-white" />
                      </div>
                      <AnimatedCounter
                        end={1000000}
                        suffix="+"
                        className="text-4xl font-bold text-slate-900"
                      />
                      <p className="text-slate-600 font-medium">Questions Answered</p>
                    </div>
                    <div className="space-y-2">
                      <div className="inline-flex items-center justify-center h-16 w-16 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-2xl mx-auto mb-4 animate-float animation-delay-200">
                        <Star className="h-8 w-8 text-white" />
                      </div>
                      <div className="text-4xl font-bold text-slate-900">4.9/5</div>
                      <p className="text-slate-600 font-medium">Average Rating</p>
                    </div>
                    <div className="space-y-2">
                      <div className="inline-flex items-center justify-center h-16 w-16 bg-gradient-to-br from-[#EF4444] to-[#DC2626] rounded-2xl mx-auto mb-4 animate-float animation-delay-400">
                        <Sparkles className="h-8 w-8 text-white" />
                      </div>
                      <AnimatedCounter
                        end={150}
                        suffix="+"
                        className="text-4xl font-bold text-slate-900"
                      />
                      <p className="text-slate-600 font-medium">Creator Partnerships</p>
                    </div>
                  </div>
                </Card>
              </section>
            </FadeInWhenVisible>

            <FadeInWhenVisible>
              <section id="community" className="mb-8">
                <Card className="bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-3xl shadow-xl border border-white/10 p-12 text-white hover-lift overflow-hidden relative group">
                  <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjxwYXRoIGQ9Ik0zNiAxOGMzLjMxNCAwIDYgMi42ODYgNiA2cy0yLjY4NiA2LTYgNi02LTIuNjg2LTYtNiAyLjY4Ni02IDYtNnoiIHN0cm9rZT0iI2ZmZiIgc3Ryb2tlLW9wYWNpdHk9Ii4wNSIgc3Ryb2tlLXdpZHRoPSIyIi8+PC9nPjwvc3ZnPg==')] opacity-20"></div>
                  <div className="relative z-10 text-center max-w-3xl mx-auto space-y-6">
                    <div className="inline-flex items-center gap-2 bg-white/10 backdrop-blur-sm px-5 py-2.5 rounded-full text-sm font-semibold animate-bounce-in">
                      <Heart className="h-4 w-4 animate-pulse-slow" />
                      Join the Creator Community
                    </div>
                    <h2 className="text-4xl font-bold animate-fade-in-up animation-delay-200">Ready to Connect?</h2>
                    <p className="text-purple-100 text-lg leading-relaxed animate-fade-in-up animation-delay-400">
                      Thousands of fans are already discovering insights from their favorite creators. Start exploring today and be part of something special.
                    </p>
                    <Link href="/signup">
                      <Button size="lg" className="gap-2 text-lg px-8 rounded-xl bg-gradient-to-r from-[#EF4444] to-[#DC2626] hover:from-[#DC2626] hover:to-[#B91C1C] text-white shadow-lg mt-6 hover-lift btn-ripple animate-bounce-in animation-delay-600 group">
                        Start Your Journey <ArrowRight className="h-5 w-5 group-hover:translate-x-1 transition-transform" />
                      </Button>
                    </Link>
                  </div>
                </Card>
              </section>
            </FadeInWhenVisible>
          </main>
        </div>
      </div>
    </div>
  );
}
