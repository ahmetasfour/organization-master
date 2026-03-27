'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import {
  ClipboardList,
  Star,
  Users,
  FileText,
  Menu,
  X,
  LogOut,
} from 'lucide-react';
import { useAuthStore, useAuthHydrated } from '../../lib/store/auth.store';
import { RoleBadge } from '../../components/ui/RoleBadge';
import { cn } from '../../lib/utils';

// ─── Navigation Configuration ─────────────────────────────────────────────────

interface NavItem {
  href: string;
  label: string;
  icon: React.ReactNode;
  roles: string[];
}

const navItems: NavItem[] = [
  {
    href: '/applications',
    label: 'Başvurular',
    roles: ['koordinator', 'asil_uye', 'yik_uye', 'yk', 'admin'],
    icon: <ClipboardList className="h-5 w-5" />,
  },
  {
    href: '/members',
    label: 'Üyeler',
    roles: ['koordinator', 'asil_uye', 'yik_uye', 'yk', 'admin'],
    icon: <Users className="h-5 w-5" />,
  },
  {
    href: '/honorary',
    label: 'Onursal Öneriler',
    roles: ['asil_uye', 'yik_uye', 'yk', 'admin'],
    icon: <Star className="h-5 w-5" />,
  },
  {
    href: '/logs',
    label: 'Sistem Logları',
    roles: ['yk', 'admin', 'koordinator'],
    icon: <FileText className="h-5 w-5" />,
  },
];

// ─── Helper Functions ─────────────────────────────────────────────────────────

function hasRole(userRole: string | undefined, allowedRoles: string[]): boolean {
  if (!userRole) return false;
  return allowedRoles.includes(userRole);
}

// ─── Layout Component ─────────────────────────────────────────────────────────

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, user, clearAuth } = useAuthStore();
  const hydrated = useAuthHydrated();
  const router = useRouter();
  const pathname = usePathname();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  useEffect(() => {
    if (hydrated && !isAuthenticated) {
      router.replace('/login');
    }
  }, [hydrated, isAuthenticated, router]);

  // Close sidebar on route change (mobile)
  useEffect(() => {
    setSidebarOpen(false);
  }, [pathname]);

  // Wait for auth state to rehydrate from localStorage
  if (!hydrated) {
    return (
      <div className="flex h-screen items-center justify-center bg-gray-50">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-gray-300 border-t-gray-800" />
      </div>
    );
  }

  if (!isAuthenticated) return null;

  const filteredNavItems = navItems.filter((item) =>
    hasRole(user?.role, item.roles)
  );

  const handleLogout = () => {
    clearAuth();
    router.push('/login');
  };

  return (
    <div className="flex h-screen bg-gray-50">
      {/* Mobile overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside
        className={cn(
          'fixed inset-y-0 left-0 z-50 w-64 flex-shrink-0 transform bg-white border-r border-gray-200 flex flex-col transition-transform duration-300 ease-in-out lg:static lg:translate-x-0',
          sidebarOpen ? 'translate-x-0' : '-translate-x-full'
        )}
      >
        {/* Logo */}
        <div className="h-16 flex items-center justify-between px-6 border-b border-gray-200">
          <div className="flex items-center">
            <img
              src="/teknokratlar-logo.svg"
              alt="Teknokratlar"
              className="h-8 w-auto mr-3"
            />
            <span className="font-bold text-gray-900 text-sm leading-tight">
              Teknokratlar
              <br />
              Derneği
            </span>
          </div>
          <button
            onClick={() => setSidebarOpen(false)}
            className="lg:hidden text-gray-400 hover:text-gray-600"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Navigation */}
        <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
          {filteredNavItems.map((item) => {
            const isActive =
              pathname === item.href || pathname.startsWith(`${item.href}/`);
            return (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  'flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-blue-50 text-blue-700'
                    : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                )}
              >
                {item.icon}
                {item.label}
              </Link>
            );
          })}

          {/* Yeni Onursal Öneri sub-link */}
          {hasRole(user?.role, ['asil_uye', 'yik_uye']) && (
            <Link
              href="/honorary/new"
              className={cn(
                'flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors ml-6',
                pathname === '/honorary/new'
                  ? 'bg-blue-50 text-blue-700'
                  : 'text-gray-500 hover:bg-gray-100 hover:text-gray-900'
              )}
            >
              <span className="text-xs">＋</span>
              Yeni Öneri
            </Link>
          )}
        </nav>

        {/* User footer */}
        <div className="border-t border-gray-200 p-4">
          <div className="flex items-center gap-3 mb-3">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-white font-semibold text-sm shadow-sm">
              {user?.fullName?.charAt(0)?.toUpperCase() ?? 'U'}
            </div>
            <div className="min-w-0 flex-1">
              <p className="text-sm font-medium text-gray-900 truncate">
                {user?.fullName}
              </p>
              <RoleBadge role={user?.role ?? ''} className="mt-0.5" />
            </div>
          </div>
          <button
            onClick={handleLogout}
            className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-sm text-gray-600 transition-colors hover:bg-gray-100 hover:text-red-600"
          >
            <LogOut className="h-4 w-4" />
            Çıkış Yap
          </button>
        </div>
      </aside>

      {/* Main content */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* Mobile header */}
        <header className="flex h-16 items-center justify-between border-b border-gray-200 bg-white px-4 lg:hidden">
          <button
            onClick={() => setSidebarOpen(true)}
            className="rounded-lg p-2 text-gray-600 hover:bg-gray-100"
          >
            <Menu className="h-6 w-6" />
          </button>
          <div className="flex items-center">
            <img
              src="/teknokratlar-logo.svg"
              alt="Teknokratlar"
              className="h-8 w-auto"
            />
          </div>
          <div className="w-10" /> {/* Spacer for alignment */}
        </header>

        {/* Page content */}
        <main className="flex-1 overflow-y-auto">{children}</main>
      </div>
    </div>
  );
}
