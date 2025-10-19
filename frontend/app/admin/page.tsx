'use client';

import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  Users,
  MessageSquare,
  Upload,
  Activity,
  ArrowRight,
  TrendingUp,
  FileText,
  Video,
  Loader2
} from 'lucide-react';
import Link from 'next/link';
import { FadeInWhenVisible } from '@/components/animations/FadeInWhenVisible';
import { AnimatedCounter } from '@/components/animations/AnimatedCounter';

type Stats = {
  totalChannels: number;
  totalUsers: number;
  totalQuestions: number;
  totalTranscriptions: number;
};

type RecentActivity = {
  id: string;
  type: 'channel' | 'user' | 'question' | 'transcription';
  description: string;
  timestamp: string;
};

export default function AdminDashboard() {
  const [stats, setStats] = useState<Stats>({
    totalChannels: 0,
    totalUsers: 0,
    totalQuestions: 0,
    totalTranscriptions: 0,
  });
  const [recentActivity, setRecentActivity] = useState<RecentActivity[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    try {
      const token = localStorage.getItem('admin_auth_token');

      // Fetch stats
      const statsResponse = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/admin/stats`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (statsResponse.ok) {
        const statsData = await statsResponse.json();
        setStats(statsData);
      }

      // Fetch recent activity
      const activityResponse = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/admin/activity`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (activityResponse.ok) {
        const activityData = await activityResponse.json();
        setRecentActivity(activityData.activities || []);
      }
    } catch (error) {
      console.error('Error fetching dashboard data:', error);
    } finally {
      setLoading(false);
    }
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
      {/* Welcome Section */}
      <FadeInWhenVisible>
        <Card className="bg-gradient-to-br from-white/90 to-purple-50/50 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50 p-8 hover-lift overflow-hidden relative group">
          <div className="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-[#6D28D9]/10 to-[#5B21B6]/10 rounded-full blur-3xl group-hover:scale-150 transition-transform duration-700"></div>
          <div className="relative z-10 flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div className="h-14 w-14 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-2xl flex items-center justify-center shadow-lg animate-float">
                <Activity className="h-7 w-7 text-white" />
              </div>
              <div>
                <h2 className="text-4xl font-bold text-slate-900 mb-2">Admin Dashboard</h2>
                <p className="text-slate-600 text-lg">Manage content and monitor platform activity</p>
              </div>
            </div>
          </div>
        </Card>
      </FadeInWhenVisible>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <FadeInWhenVisible delay={100}>
          <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50 hover:shadow-2xl hover:border-[#6D28D9]/50 transition-all hover-lift">
            <CardHeader className="p-6">
              <div className="flex items-center justify-between">
                <div className="h-12 w-12 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-2xl flex items-center justify-center shadow-lg">
                  <Video className="h-6 w-6 text-white" />
                </div>
                <TrendingUp className="h-5 w-5 text-green-500" />
              </div>
              <CardTitle className="text-sm font-medium text-slate-600 mt-4">Total Channels</CardTitle>
              <div className="mt-2">
                <AnimatedCounter
                  end={stats.totalChannels}
                  className="text-3xl font-bold text-slate-900"
                />
              </div>
            </CardHeader>
          </Card>
        </FadeInWhenVisible>

        <FadeInWhenVisible delay={200}>
          <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50 hover:shadow-2xl hover:border-[#6D28D9]/50 transition-all hover-lift">
            <CardHeader className="p-6">
              <div className="flex items-center justify-between">
                <div className="h-12 w-12 bg-gradient-to-br from-[#FB7185] to-[#EF4444] rounded-2xl flex items-center justify-center shadow-lg">
                  <Users className="h-6 w-6 text-white" />
                </div>
                <TrendingUp className="h-5 w-5 text-green-500" />
              </div>
              <CardTitle className="text-sm font-medium text-slate-600 mt-4">Total Users</CardTitle>
              <div className="mt-2">
                <AnimatedCounter
                  end={stats.totalUsers}
                  className="text-3xl font-bold text-slate-900"
                />
              </div>
            </CardHeader>
          </Card>
        </FadeInWhenVisible>

        <FadeInWhenVisible delay={300}>
          <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50 hover:shadow-2xl hover:border-[#6D28D9]/50 transition-all hover-lift">
            <CardHeader className="p-6">
              <div className="flex items-center justify-between">
                <div className="h-12 w-12 bg-gradient-to-br from-[#10B981] to-[#059669] rounded-2xl flex items-center justify-center shadow-lg">
                  <MessageSquare className="h-6 w-6 text-white" />
                </div>
                <TrendingUp className="h-5 w-5 text-green-500" />
              </div>
              <CardTitle className="text-sm font-medium text-slate-600 mt-4">Total Questions</CardTitle>
              <div className="mt-2">
                <AnimatedCounter
                  end={stats.totalQuestions}
                  className="text-3xl font-bold text-slate-900"
                />
              </div>
            </CardHeader>
          </Card>
        </FadeInWhenVisible>

        <FadeInWhenVisible delay={400}>
          <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50 hover:shadow-2xl hover:border-[#6D28D9]/50 transition-all hover-lift">
            <CardHeader className="p-6">
              <div className="flex items-center justify-between">
                <div className="h-12 w-12 bg-gradient-to-br from-[#F59E0B] to-[#D97706] rounded-2xl flex items-center justify-center shadow-lg">
                  <FileText className="h-6 w-6 text-white" />
                </div>
                <TrendingUp className="h-5 w-5 text-green-500" />
              </div>
              <CardTitle className="text-sm font-medium text-slate-600 mt-4">Transcriptions</CardTitle>
              <div className="mt-2">
                <AnimatedCounter
                  end={stats.totalTranscriptions}
                  className="text-3xl font-bold text-slate-900"
                />
              </div>
            </CardHeader>
          </Card>
        </FadeInWhenVisible>
      </div>

      {/* Quick Actions & Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Quick Actions */}
        <FadeInWhenVisible delay={500}>
          <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50 lg:col-span-1">
            <CardHeader className="p-6">
              <CardTitle className="text-xl font-bold text-slate-900">Quick Actions</CardTitle>
              <CardDescription>Common administrative tasks</CardDescription>
            </CardHeader>
            <CardContent className="p-6 pt-0 space-y-3">
              <Link href="/admin/channels">
                <Button className="w-full justify-between rounded-xl bg-gradient-to-r from-[#6D28D9] to-[#5B21B6] hover:from-[#5B21B6] hover:to-[#4C1D95] shadow-lg group">
                  <span className="flex items-center gap-2">
                    <Upload className="h-4 w-4" />
                    Manage Channels
                  </span>
                  <ArrowRight className="h-4 w-4 group-hover:translate-x-1 transition-transform" />
                </Button>
              </Link>
              <Link href="/admin/channels">
                <Button variant="outline" className="w-full justify-between rounded-xl border-2 border-purple-200 hover:bg-purple-50 group">
                  <span className="flex items-center gap-2 text-[#6D28D9]">
                    <FileText className="h-4 w-4" />
                    Upload Content
                  </span>
                  <ArrowRight className="h-4 w-4 group-hover:translate-x-1 transition-transform" />
                </Button>
              </Link>
              <Link href="/">
                <Button variant="outline" className="w-full justify-between rounded-xl border-2 border-purple-200 hover:bg-purple-50 group">
                  <span className="flex items-center gap-2 text-[#6D28D9]">
                    <Users className="h-4 w-4" />
                    View User Site
                  </span>
                  <ArrowRight className="h-4 w-4 group-hover:translate-x-1 transition-transform" />
                </Button>
              </Link>
            </CardContent>
          </Card>
        </FadeInWhenVisible>

        {/* Recent Activity */}
        <FadeInWhenVisible delay={600}>
          <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50 lg:col-span-2">
            <CardHeader className="p-6">
              <CardTitle className="text-xl font-bold text-slate-900">Recent Activity</CardTitle>
              <CardDescription>Latest platform events and updates</CardDescription>
            </CardHeader>
            <CardContent className="p-6 pt-0">
              {recentActivity.length === 0 ? (
                <div className="text-center py-12 text-slate-500">
                  <Activity className="h-16 w-16 mx-auto mb-4 opacity-30 text-[#6D28D9]" />
                  <p className="text-lg font-medium">No Recent Activity</p>
                  <p className="text-sm mt-1">Activity will appear here as actions are performed</p>
                </div>
              ) : (
                <div className="space-y-4 max-h-[300px] overflow-y-auto">
                  {recentActivity.map((activity) => (
                    <div
                      key={activity.id}
                      className="flex items-start gap-4 p-4 rounded-2xl bg-gradient-to-br from-purple-50 to-pink-50 border border-purple-100 hover:border-[#6D28D9]/50 transition-colors"
                    >
                      <div className="h-10 w-10 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-xl flex items-center justify-center shrink-0">
                        {activity.type === 'channel' && <Video className="h-5 w-5 text-white" />}
                        {activity.type === 'user' && <Users className="h-5 w-5 text-white" />}
                        {activity.type === 'question' && <MessageSquare className="h-5 w-5 text-white" />}
                        {activity.type === 'transcription' && <FileText className="h-5 w-5 text-white" />}
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm text-slate-900 font-medium">{activity.description}</p>
                        <p className="text-xs text-slate-500 mt-1">
                          {new Date(activity.timestamp).toLocaleString()}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </FadeInWhenVisible>
      </div>
    </div>
  );
}