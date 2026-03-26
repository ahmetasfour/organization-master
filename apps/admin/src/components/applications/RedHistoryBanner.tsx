'use client';

import Link from 'next/link';

interface RedHistoryBannerProps {
  applicationId: string;
  previousAppId?: string;
  repeatApplicant: boolean;
  userRole: string;
}

export function RedHistoryBanner({
  applicationId,
  previousAppId,
  repeatApplicant,
  userRole,
}: RedHistoryBannerProps) {
  const canView = userRole === 'yk' || userRole === 'admin';

  if (!repeatApplicant || !canView) return null;

  return (
    <div className="flex items-start gap-3 rounded-lg border border-red-300 bg-red-50 px-4 py-3">
      {/* Icon */}
      <div className="mt-0.5 flex-shrink-0">
        <svg
          className="h-5 w-5 text-red-500"
          fill="none"
          viewBox="0 0 24 24"
          strokeWidth={2}
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z"
          />
        </svg>
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0">
        <p className="text-sm font-semibold text-red-800">
          Tekrar Başvuran
        </p>
        <p className="mt-0.5 text-sm text-red-700">
          Bu başvuran daha önce reddedilmiştir.
        </p>

        <div className="mt-2 flex flex-wrap gap-2">
          {previousAppId && (
            <Link
              href={`/applications/${previousAppId}`}
              className="text-xs font-medium text-red-700 underline underline-offset-2 hover:text-red-900"
            >
              Önceki Başvuruyu Görüntüle
            </Link>
          )}
          <Link
            href={`/applications/${applicationId}/red-history`}
            className="text-xs font-medium text-red-700 underline underline-offset-2 hover:text-red-900"
          >
            Red Geçmişini Görüntüle
          </Link>
        </div>
      </div>
    </div>
  );
}
