'use client';

import { useRouter } from 'next/navigation';
import { useEffect, ReactNode } from 'react';
import { useAuthStore } from '../../lib/store/auth.store';

interface AuthGuardProps {
  children: ReactNode;
  /** Required roles to access this page. If empty, any authenticated user can access. */
  allowedRoles?: string[];
  /** Redirect path when unauthorized. Defaults to /login for unauthenticated, /applications for wrong role. */
  redirectTo?: string;
}

/**
 * AuthGuard wraps pages that require authentication and optional role-based access control.
 * 
 * Usage:
 * ```tsx
 * <AuthGuard allowedRoles={['yk', 'admin']}>
 *   <ProtectedContent />
 * </AuthGuard>
 * ```
 */
export function AuthGuard({ 
  children, 
  allowedRoles = [], 
  redirectTo 
}: AuthGuardProps) {
  const { isAuthenticated, user } = useAuthStore();
  const router = useRouter();

  useEffect(() => {
    if (!isAuthenticated) {
      router.replace(redirectTo || '/login');
      return;
    }

    if (allowedRoles.length > 0 && user?.role) {
      if (!allowedRoles.includes(user.role)) {
        // User doesn't have required role — redirect to applications
        router.replace(redirectTo || '/applications');
      }
    }
  }, [isAuthenticated, user, allowedRoles, redirectTo, router]);

  // Don't render until we've checked auth
  if (!isAuthenticated) {
    return null;
  }

  // Check role if required
  if (allowedRoles.length > 0 && user?.role && !allowedRoles.includes(user.role)) {
    return null;
  }

  return <>{children}</>;
}

/**
 * Hook to check if current user has one of the specified roles.
 */
export function useHasRole(allowedRoles: string[]): boolean {
  const { user, isAuthenticated } = useAuthStore();
  
  if (!isAuthenticated || !user?.role) {
    return false;
  }
  
  return allowedRoles.includes(user.role);
}

/**
 * Component that only renders its children if the user has one of the specified roles.
 * Does not redirect — just hides content.
 */
export function RoleGate({ 
  children, 
  allowedRoles,
  fallback = null
}: { 
  children: ReactNode;
  allowedRoles: string[];
  fallback?: ReactNode;
}) {
  const hasRole = useHasRole(allowedRoles);
  
  if (!hasRole) {
    return <>{fallback}</>;
  }
  
  return <>{children}</>;
}

/**
 * HOC to wrap pages that require specific roles.
 */
export function withRoleGuard<P extends object>(
  Component: React.ComponentType<P>,
  allowedRoles: string[]
) {
  return function RoleGuardedComponent(props: P) {
    return (
      <AuthGuard allowedRoles={allowedRoles}>
        <Component {...props} />
      </AuthGuard>
    );
  };
}
