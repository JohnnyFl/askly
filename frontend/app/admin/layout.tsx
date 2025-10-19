'use client';

import { useEffect, useState } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Shield, LogOut, LayoutDashboard, Upload, Users } from 'lucide-react';
import Link from 'next/link';
import { Toaster } from '@/components/ui/toaster';

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const pathname = usePathname();
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Skip auth check for login page
    if (pathname === '/admin/login') {
      setLoading(false);
      return;
    }

    const token = localStorage.getItem('admin_auth_token');
    if (!token) {
      router.push('/admin/login');
      return;
    }

    setIsAuthenticated(true);
    setLoading(false);
  }, [pathname, router]);

  const handleLogout = () => {
    localStorage.removeItem('admin_auth_token');
    router.push('/admin/login');
  };

  // Show nothing while checking auth (except on login page)
  if (loading && pathname !== '/admin/login') {
    return null;
  }

  // Don't show header on login page
  if (pathname === '/admin/login') {
    return <>{children}</>;
  }

  // Show authenticated layout
  return (
    <div className="min-h-screen bg-[#F5F3FF]">
      <div className="min-h-screen bg-gradient-to-br from-purple-50/50 to-indigo-50/30">
        <div className="max-w-[1400px] mx-auto p-8">
          {/* Admin Header */}
          <header className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50 mb-8 px-8 py-5 animate-fade-in-down">
            <div className="flex justify-between items-center">
              <div className="flex items-center gap-6">
                <div className="flex items-center gap-3">
                  <div className="h-10 w-10 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-xl flex items-center justify-center shadow-lg animate-float">
                    <Shield className="h-6 w-6 text-white" />
                  </div>
                  <div>
                    <h1 className="text-2xl font-bold bg-gradient-to-r from-[#6D28D9] to-[#5B21B6] bg-clip-text text-transparent">
                      Admin Portal
                    </h1>
                  </div>
                </div>

                {/* Navigation */}
                <nav className="hidden md:flex gap-2 ml-8">
                  <Link href="/admin">
                    <Button
                      variant={pathname === '/admin' ? 'default' : 'ghost'}
                      className={`gap-2 rounded-xl font-medium transition-all ${
                        pathname === '/admin'
                          ? 'bg-gradient-to-r from-[#6D28D9] to-[#5B21B6] text-white'
                          : 'hover:bg-purple-50 hover:text-[#6D28D9]'
                      }`}
                    >
                      <LayoutDashboard className="h-4 w-4" />
                      Dashboard
                    </Button>
                  </Link>
                  <Link href="/admin/channels">
                    <Button
                      variant={pathname?.startsWith('/admin/channels') ? 'default' : 'ghost'}
                      className={`gap-2 rounded-xl font-medium transition-all ${
                        pathname?.startsWith('/admin/channels')
                          ? 'bg-gradient-to-r from-[#6D28D9] to-[#5B21B6] text-white'
                          : 'hover:bg-purple-50 hover:text-[#6D28D9]'
                      }`}
                    >
                      <Upload className="h-4 w-4" />
                      Channels
                    </Button>
                  </Link>
                </nav>
              </div>

              <div className="flex items-center gap-3">
                <Link href="/">
                  <Button
                    variant="ghost"
                    className="gap-2 rounded-xl font-medium hover:bg-purple-50 hover:text-[#6D28D9] transition-all"
                  >
                    <Users className="h-4 w-4" />
                    View Site
                  </Button>
                </Link>
                <Button
                  variant="ghost"
                  onClick={handleLogout}
                  className="gap-2 rounded-xl font-medium hover:bg-purple-50 hover:text-[#EF4444] transition-all"
                >
                  <LogOut className="h-4 w-4" />
                  Logout
                </Button>
              </div>
            </div>
          </header>

          {/* Page Content */}
          <main>{children}</main>
        </div>
      </div>
      <Toaster />
    </div>
  );
}