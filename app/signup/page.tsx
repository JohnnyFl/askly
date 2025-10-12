'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { AlertCircle, MessageSquare, ArrowLeft, Sparkles } from 'lucide-react';
import { Alert, AlertDescription } from '@/components/ui/alert';

export default function SignUpPage() {
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [fullName, setFullName] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSignUp = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/auth/signup`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email,
          password,
          full_name: fullName
        }),
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Sign up failed');
      }

      const data = await response.json();
      localStorage.setItem('auth_token', data.token);
      router.push('/channels');
    } catch (err: any) {
      setError(err.message || 'An error occurred during sign up');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#FDF2F8]">
      <div className="min-h-screen bg-gradient-to-br from-pink-50/50 to-purple-50/30 flex items-center justify-center p-4">
        <div className="w-full max-w-md">
          <Link href="/">
            <Button variant="ghost" className="mb-6 text-slate-700 hover:bg-pink-50 hover:text-[#EF4444] rounded-xl gap-2">
              <ArrowLeft className="h-4 w-4" />
              Back to Home
            </Button>
          </Link>

          <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50">
            <CardHeader className="space-y-4 p-8 pb-6">
              <div className="flex justify-center">
                <div className="h-16 w-16 bg-gradient-to-br from-[#FB7185] to-[#EF4444] rounded-2xl flex items-center justify-center shadow-lg">
                  <Sparkles className="h-8 w-8 text-white" />
                </div>
              </div>
              <CardTitle className="text-3xl font-bold text-center text-slate-900">Join Askly</CardTitle>
              <CardDescription className="text-center text-base text-slate-600">
                Connect with creators and start your journey
              </CardDescription>
            </CardHeader>
            <CardContent className="p-8 pt-0">
              <form onSubmit={handleSignUp} className="space-y-5">
                {error && (
                  <Alert variant="destructive" className="rounded-xl border-red-200">
                    <AlertCircle className="h-4 w-4" />
                    <AlertDescription>{error}</AlertDescription>
                  </Alert>
                )}

                <div className="space-y-2">
                  <Label htmlFor="fullName" className="text-slate-700 font-medium">Full Name</Label>
                  <Input
                    id="fullName"
                    type="text"
                    placeholder="John Doe"
                    value={fullName}
                    onChange={(e) => setFullName(e.target.value)}
                    required
                    className="rounded-xl border-2 border-pink-100 focus:border-[#FB7185] h-12 text-base"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="email" className="text-slate-700 font-medium">Email</Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="[email protected]"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                    className="rounded-xl border-2 border-pink-100 focus:border-[#FB7185] h-12 text-base"
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
                    minLength={6}
                    className="rounded-xl border-2 border-pink-100 focus:border-[#FB7185] h-12 text-base"
                  />
                  <p className="text-xs text-slate-500">Must be at least 6 characters</p>
                </div>

                <Button
                  type="submit"
                  className="w-full rounded-xl h-12 text-base font-medium bg-[#EF4444] hover:bg-[#DC2626] shadow-lg"
                  disabled={loading}
                >
                  {loading ? 'Creating account...' : 'Create Account'}
                </Button>
              </form>

              <div className="mt-6 text-center text-sm text-slate-600">
                Already have an account?{' '}
                <Link href="/login" className="text-[#EF4444] hover:text-[#DC2626] hover:underline font-semibold">
                  Log in
                </Link>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
