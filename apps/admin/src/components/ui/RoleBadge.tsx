'use client';

import { cn } from '@/lib/utils';

export type Role = 'admin' | 'yk' | 'yik' | 'koordinator' | 'asil_uye' | 'yik_uye';

interface RoleBadgeProps {
  role: string;
  className?: string;
}

const roleConfig: Record<
  string,
  { label: string; className: string }
> = {
  admin: {
    label: 'Sistem Yöneticisi',
    className: 'bg-purple-100 text-purple-800 border-purple-200',
  },
  yk: {
    label: 'YK Üyesi',
    className: 'bg-blue-100 text-blue-800 border-blue-200',
  },
  yik: {
    label: 'YİK Üyesi',
    className: 'bg-indigo-100 text-indigo-800 border-indigo-200',
  },
  koordinator: {
    label: 'Üyelik Koordinatörü',
    className: 'bg-teal-100 text-teal-800 border-teal-200',
  },
  asil_uye: {
    label: 'Asil Üye',
    className: 'bg-green-100 text-green-800 border-green-200',
  },
  yik_uye: {
    label: 'YİK Üyesi',
    className: 'bg-cyan-100 text-cyan-800 border-cyan-200',
  },
};

export function RoleBadge({ role, className }: RoleBadgeProps) {
  const config = roleConfig[role] ?? {
    label: role,
    className: 'bg-gray-100 text-gray-700 border-gray-200',
  };

  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium',
        config.className,
        className
      )}
    >
      {config.label}
    </span>
  );
}
