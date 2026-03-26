'use client';

import { notFound, useParams } from 'next/navigation';
import { ConsultationPanel } from '../../../../../components/consultation/ConsultationPanel';
import { useApplication } from '../../../../../lib/hooks/useApplications';
import { useAuthStore } from '../../../../../lib/store/auth.store';

export default function ConsultationManagementPage() {
  const params = useParams<{ id: string }>();
  const id = params?.id ?? '';
  const { user } = useAuthStore();
  const role = user?.role ?? '';

  const { data: app, isLoading, isError } = useApplication(id);

  // Only koordinator and admin may access this page
  if (role !== 'koordinator' && role !== 'admin') {
    return (
      <div className="p-6">
        <p className="text-red-600 text-sm font-medium">Bu sayfaya erişim yetkiniz bulunmamaktadır.</p>
      </div>
    );
  }

  if (isLoading) {
    return <div className="p-6 text-gray-400 text-sm animate-pulse">Yükleniyor...</div>;
  }

  if (isError || !app) {
    return notFound();
  }

  return (
    <div className="p-6 max-w-4xl space-y-6">
      {/* Back */}
      <a
        href={`/applications/${id}`}
        className="text-sm text-blue-600 hover:underline inline-flex items-center gap-1"
      >
        ← Başvuruya Dön
      </a>

      {/* Page header */}
      <div className="border-b border-gray-200 pb-4">
        <h1 className="text-xl font-bold text-gray-900">Danışma Yönetimi</h1>
        <p className="mt-1 text-sm text-gray-500">
          <span className="font-medium">{app.applicant_name}</span> —{' '}
          <span className="capitalize">{app.membership_type}</span>
        </p>
      </div>

      {/* Info banner for non-consultation types */}
      {app.membership_type !== 'profesyonel' && app.membership_type !== 'öğrenci' && (
        <div className="rounded-lg bg-amber-50 border border-amber-200 px-4 py-3 text-sm text-amber-800">
          ⚠️ Danışma süreci yalnızca <strong>Profesyonel</strong> ve{' '}
          <strong>Öğrenci</strong> başvuruları için geçerlidir.
        </div>
      )}

      {/* Consultation panel */}
      <ConsultationPanel applicationId={id} membershipType={app.membership_type} />
    </div>
  );
}
