'use client';

import { notFound, useParams, useRouter } from 'next/navigation';
import { ArrowLeft } from 'lucide-react';
import { VotePanel } from '@/components/voting/VotePanel';
import { useApplication } from '@/lib/hooks/useApplications';
import { useAuthStore } from '@/lib/store/auth.store';
import { StatusBadge } from '@/components/applications/StatusBadge';

/**
 * YİK Değerlendirme voting page.
 * Only for Onursal applications. Accessible by users with role "yik".
 * Application must be in status "yik_değerlendirmede".
 */
export default function YIKVotingPage() {
  const params = useParams<{ id: string }>();
  const id = params?.id ?? '';
  const router = useRouter();
  const { user } = useAuthStore();
  const role = user?.role ?? '';
  const userId = user?.id ?? '';

  const { data: app, isLoading, isError } = useApplication(id);

  if (isLoading) {
    return <div className="p-6 text-sm text-slate-400 animate-pulse">Yükleniyor…</div>;
  }

  if (isError || !app) {
    return notFound();
  }

  // Gate: only yik and admin can access this page
  if (role !== 'yik' && role !== 'admin') {
    return (
      <div className="p-6 text-sm text-red-600">
        Bu sayfaya erişim yetkiniz bulunmamaktadır.
      </div>
    );
  }

  // This stage only applies to Onursal applications
  if (app.membership_type !== 'onursal') {
    return (
      <div className="p-6 text-sm text-slate-500">
        YİK oylaması yalnızca Onursal üyelik başvuruları için geçerlidir.
      </div>
    );
  }

  return (
    <div className="max-w-3xl space-y-6 p-6">
      {/* Back button */}
      <button
        onClick={() => router.push(`/applications/${id}`)}
        className="flex items-center gap-1.5 text-sm text-blue-600 hover:underline"
      >
        <ArrowLeft className="h-4 w-4" />
        Başvuru Detayına Dön
      </button>

      {/* Application header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-800">{app.applicant_name}</h1>
          <p className="text-sm text-slate-500 capitalize">{app.membership_type} üyeliği</p>
        </div>
        <StatusBadge status={app.status} />
      </div>

      {/* Vote panel */}
      <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
        <VotePanel
          applicationId={id}
          applicantName={app.applicant_name}
          applicationStatus={app.status}
          rejectionReason={app.rejection_reason}
          stage="yik"
          viewerRole={role}
          viewerId={userId}
        />
      </div>
    </div>
  );
}
