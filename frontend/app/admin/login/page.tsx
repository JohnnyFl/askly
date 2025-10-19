'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { AlertCircle, Shield, ArrowLeft } from 'lucide-react';
import { Alert, AlertDescription } from '@/components/ui/alert';

export default function AdminLoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/admin/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Login failed');
      }

      const data = await response.json();
      localStorage.setItem('admin_auth_token', data.token);
      router.push('/admin');
    } catch (err: any) {
      setError(err.message || 'An error occurred during login');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#F5F3FF]">
      <div className="min-h-screen bg-gradient-to-br from-purple-50/50 to-indigo-50/30 flex items-center justify-center p-4">
        <div className="w-full max-w-md">
          <Link href="/">
            <Button variant="ghost" className="mb-6 text-slate-700 hover:bg-purple-50 hover:text-[#6D28D9] rounded-xl gap-2">
              <ArrowLeft className="h-4 w-4" />
              Back to Home
            </Button>
          </Link>

          <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50">
            <CardHeader className="space-y-4 p-8 pb-6">
              <div className="flex justify-center">
                <div className="h-16 w-16 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-2xl flex items-center justify-center shadow-lg">
                  <Shield className="h-8 w-8 text-white" />
                </div>
              </div>
              <CardTitle className="text-3xl font-bold text-center text-slate-900">Admin Portal</CardTitle>
              <CardDescription className="text-center text-base text-slate-600">
                Secure access to content management
              </CardDescription>
            </CardHeader>
            <CardContent className="p-8 pt-0">
              <form onSubmit={handleLogin} className="space-y-5">
                {error && (
                  <Alert variant="destructive" className="rounded-xl border-red-200">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription>{error}</AlertDescription>
                  </Alert>
                )}

                <div className="space-y-2">
                  <Label htmlFor="email" className="text-slate-700 font-medium">Admin Email</Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="[email protected]"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                    className="rounded-xl border-2 border-purple-100 focus:border-[#6D28D9] h-12 text-base"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="password" className="text-slate-700 font-medium">Password</Label>
                  <Input
                    id="password"
                    type="password"
                    placeholder="••••••••"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                    className="rounded-xl border-2 border-purple-100 focus:border-[#6D28D9] h-12 text-base"
                  />
                </div>

                <Button
                  type="submit"
                  className="w-full rounded-xl h-12 text-base font-medium bg-gradient-to-r from-[#6D28D9] to-[#5B21B6] hover:from-[#5B21B6] hover:to-[#4C1D95] shadow-lg"
                  disabled={loading}
                >
                  {loading ? 'Signing in...' : 'Sign In to Admin'}
                </Button>
              </form>

              <div className="mt-6 text-center text-sm text-slate-600">
                Not an admin?{' '}
                <Link href="/login" className="text-[#6D28D9] hover:text-[#5B21B6] hover:underline font-semibold">
                  User login
                </Link>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
