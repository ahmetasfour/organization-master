'use client';

import Link from 'next/link';
import { ShieldX } from 'lucide-react';

interface AccessDeniedProps {
  /** Title to show. Defaults to "Erişim Reddedildi" */
  title?: string;
  /** Description to show. */
  description?: string;
  /** Link to return to. Defaults to /applications */
  returnTo?: string;
  /** Label for the return link. Defaults to "Ana Sayfaya Dön" */
  returnLabel?: string;
}

/**
 * AccessDenied component is shown when a user doesn't have permission to view a page.
 * Used for 403 Forbidden scenarios in the frontend.
 */
export function AccessDenied({
  title = 'Erişim Reddedildi',
  description = 'Bu sayfayı görüntüleme yetkiniz bulunmamaktadır.',
  returnTo = '/applications',
  returnLabel = 'Ana Sayfaya Dön',
}: AccessDeniedProps) {
  return (
    <div className="flex min-h-[60vh] flex-col items-center justify-center p-6 text-center">
      <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-red-100">
        <ShieldX className="h-8 w-8 text-red-600" />
      </div>
      <h2 className="text-xl font-semibold text-gray-900">{title}</h2>
      <p className="mt-2 max-w-md text-sm text-gray-500">
        {description}
      </p>
      <Link
        href={returnTo}
        className="mt-6 rounded-lg bg-gray-900 px-5 py-2.5 text-sm font-medium text-white transition hover:bg-gray-800"
      >
        {returnLabel}
      </Link>
    </div>
  );
}
