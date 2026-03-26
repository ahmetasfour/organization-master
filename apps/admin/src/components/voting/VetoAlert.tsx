'use client';

import { AlertTriangle } from 'lucide-react';

interface VetoAlertProps {
  /** The applicant's name shown in the banner */
  applicantName: string;
  /** Rejection reason (always visible for yk/admin) */
  reason?: string;
  /** The role of the logged-in viewer — affects message copy */
  viewerRole: string;
}

/**
 * VetoAlert renders a full-width red termination banner.
 * Shown when an application has been permanently rejected via veto.
 */
export function VetoAlert({ applicantName, reason, viewerRole }: VetoAlertProps) {
  const isPrivileged = viewerRole === 'yk' || viewerRole === 'admin';

  return (
    <div className="rounded-xl border border-red-200 bg-red-50 p-5">
      <div className="flex items-start gap-3">
        <AlertTriangle className="mt-0.5 h-5 w-5 flex-shrink-0 text-red-600" />
        <div className="space-y-1">
          <p className="text-sm font-semibold text-red-800">
            Başvuru Kalıcı Olarak Reddedildi
          </p>
          <p className="text-sm text-red-700">
            {applicantName} adlı adayın başvurusu oylama sürecinde veto oyu
            aldığı için kalıcı olarak sonlandırılmıştır.{' '}
            {isPrivileged
              ? 'Bu karar geri alınamaz.'
              : 'Daha fazla bilgi için YK ile iletişime geçiniz.'}
          </p>
          {isPrivileged && reason && (
            <p className="mt-2 rounded-md bg-red-100 px-3 py-2 text-xs text-red-800">
              <span className="font-medium">Gerekçe:</span> {reason}
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
