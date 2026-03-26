'use client';

import { notFound, useParams } from 'next/navigation';
import { useState } from 'react';
import { RedHistoryBanner } from '../../../../components/applications/RedHistoryBanner';
import { StatusBadge } from '../../../../components/applications/StatusBadge';
import { StatusTimeline } from '../../../../components/applications/StatusTimeline';
import { ReputationPanel } from '../../../../components/reputation/ReputationPanel';
import { useApplication, useRedHistory, useTimeline } from '../../../../lib/hooks/useApplications';
import { useAuthStore } from '../../../../lib/store/auth.store';

type Tab = 'overview' | 'timeline' | 'red-history' | 'reputation';

export default function ApplicationDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params?.id ?? '';
  const { user } = useAuthStore();
  const role = user?.role ?? '';

  const [activeTab, setActiveTab] = useState<Tab>('overview');

  const { data: app, isLoading, isError } = useApplication(id);
  const { data: timeline } = useTimeline(id);
  const canViewRedHistory = role === 'yk' || role === 'admin';
  const { data: redHistory } = useRedHistory(
    id,
    canViewRedHistory && activeTab === 'red-history'
  );

  if (isLoading) {
    return (
      <div className="p-6 text-gray-400 text-sm">Yükleniyor...</div>
    );
  }

  if (isError || !app) {
    return notFound();
  }

  const tabs: { key: Tab; label: string }[] = [
    { key: 'overview', label: 'Genel Bilgi' },
    { key: 'timeline', label: 'Zaman Çizelgesi' },
    ...(canViewRedHistory ? [{ key: 'red-history' as Tab, label: 'Red Geçmişi' }] : []),
    ...(['asil', 'akademik'].includes(app.membership_type) &&
    ['yk', 'koordinator', 'admin'].includes(role)
      ? [{ key: 'reputation' as Tab, label: 'İtibar Tarama' }]
      : []),
  ];

  return (
    <div className="p-6 max-w-5xl space-y-6">
      {/* Back */}
      <a href="/applications" className="text-sm text-blue-600 hover:underline">
        ← Başvuru Listesine Dön
      </a>

      {/* Red history warning */}
      {app.repeat_applicant && (
        <RedHistoryBanner
          applicationId={id}
          previousAppId={app.previous_app_id}
          repeatApplicant={app.repeat_applicant}
          userRole={role}
        />
      )}

      {/* Card header */}
      <div className="bg-white border border-gray-200 rounded-xl p-6 space-y-4">
        <div className="flex items-start justify-between gap-4 flex-wrap">
          <div>
            <h1 className="text-xl font-bold text-gray-900">{app.applicant_name}</h1>
            <p className="text-sm text-gray-500">{app.applicant_email}</p>
            {app.applicant_phone && (
              <p className="text-sm text-gray-500">{app.applicant_phone}</p>
            )}
            {app.linkedin_url && (
              <a
                href={app.linkedin_url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm text-blue-600 hover:underline"
              >
                LinkedIn Profili
              </a>
            )}
          </div>
          <div className="flex flex-col items-end gap-2">
            <StatusBadge status={app.status} />
            <span className="text-xs text-gray-400 capitalize">
              {app.membership_type} üyeliği
            </span>
          </div>
        </div>

        {/* Timeline strip */}
        <div className="border-t border-gray-100 pt-4">
          <StatusTimeline membershipType={app.membership_type} currentStatus={app.status} />
        </div>
      </div>

      {/* Tabs */}
      <div>
        <div className="flex gap-0 border-b border-gray-200">
          {tabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab.key
                  ? 'border-blue-600 text-blue-700'
                  : 'border-transparent text-gray-500 hover:text-gray-900'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>

        {/* Tab: Genel Bilgi */}
        {activeTab === 'overview' && (
          <div className="bg-white border border-gray-200 border-t-0 rounded-b-xl p-6 space-y-4">
            {app.proposal_reason && (
              <InfoRow label="Teklif/Başvuru Gerekçesi" value={app.proposal_reason} />
            )}
            {app.rejection_reason && (
              <InfoRow
                label="Red Gerekçesi"
                value={app.rejection_reason}
                valueClassName="text-red-700"
              />
            )}
            {app.rejected_by_role && (
              <InfoRow label="Reddeden Rol" value={app.rejected_by_role} />
            )}
            <InfoRow
              label="Başvuru Tarihi"
              value={new Date(app.created_at).toLocaleString('tr-TR')}
            />
            <InfoRow
              label="Son Güncelleme"
              value={new Date(app.updated_at).toLocaleString('tr-TR')}
            />
            {app.allowed_next_statuses.length > 0 && (
              <div>
                <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">
                  İzin Verilen Sonraki Durumlar
                </p>
                <div className="flex flex-wrap gap-2">
                  {app.allowed_next_statuses.map((s) => (
                    <StatusBadge key={s} status={s} />
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

        {/* Tab: Timeline */}
        {activeTab === 'timeline' && (
          <div className="bg-white border border-gray-200 border-t-0 rounded-b-xl p-6">
            {!timeline || timeline.length === 0 ? (
              <p className="text-sm text-gray-400">Zaman çizelgesi verisi bulunamadı.</p>
            ) : (
              <ol className="relative border-l border-gray-200 ml-2 space-y-6">
                {timeline.map((entry, i) => (
                  <li key={i} className="ml-4">
                    <div className="absolute w-2.5 h-2.5 bg-blue-500 rounded-full -left-1.5 mt-1.5 border-2 border-white" />
                    <p className="text-sm font-semibold text-gray-900">{entry.status}</p>
                    {entry.changed_by && (
                      <p className="text-xs text-gray-500">{entry.changed_by}</p>
                    )}
                    {entry.changed_at && (
                      <time className="text-xs text-gray-400">
                        {new Date(entry.changed_at).toLocaleString('tr-TR')}
                      </time>
                    )}
                    {entry.notes && (
                      <p className="text-sm text-gray-600 mt-1">{entry.notes}</p>
                    )}
                  </li>
                ))}
              </ol>
            )}
          </div>
        )}

        {/* Tab: Red Geçmişi */}
        {activeTab === 'red-history' && canViewRedHistory && (
          <div className="bg-white border border-gray-200 border-t-0 rounded-b-xl p-6">
            {!redHistory || redHistory.length === 0 ? (
              <p className="text-sm text-gray-400">Daha önce red geçmişi bulunmamaktadır.</p>
            ) : (
              <div className="space-y-3">
                {redHistory.map((entry, i) => (
                  <div
                    key={i}
                    className="flex items-start justify-between rounded-lg border border-red-100 bg-red-50 px-4 py-3 gap-4"
                  >
                    <div className="min-w-0">
                      <p className="text-sm font-medium text-red-800">{entry.applicant_name}</p>
                      <StatusBadge status={entry.status} />
                      {entry.rejection_reason && (
                        <p className="text-xs text-red-600 mt-1">{entry.rejection_reason}</p>
                      )}
                    </div>
                    <time className="text-xs text-red-400 flex-shrink-0">
                      {new Date(entry.created_at).toLocaleDateString('tr-TR')}
                    </time>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Tab: İtibar Tarama */}
        {activeTab === 'reputation' && (
          <div className="bg-white border border-gray-200 border-t-0 rounded-b-xl p-6">
            <ReputationPanel
              applicationId={id}
              membershipType={app.membership_type}
            />
          </div>
        )}
      </div>
    </div>
  );
}

function InfoRow({
  label,
  value,
  valueClassName,
}: {
  label: string;
  value: string;
  valueClassName?: string;
}) {
  return (
    <div className="grid grid-cols-3 gap-2">
      <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide col-span-1 pt-0.5">
        {label}
      </p>
      <p className={`text-sm text-gray-900 col-span-2 ${valueClassName ?? ''}`}>{value}</p>
    </div>
  );
}
