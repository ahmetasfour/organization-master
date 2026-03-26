'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../lib/api/client';
import { useAuthStore } from '../../lib/store/auth.store';

// ─── Types ────────────────────────────────────────────────────────────────────

interface ConsultationSummary {
  id: string;
  application_id: string;
  member_user_id: string;
  member_name: string;
  member_email: string;
  response_type?: 'positive' | 'negative';
  status: 'pending' | 'positive' | 'negative' | 'expired';
  created_at: string;
}

interface UserSearchResult {
  id: string;
  full_name: string;
  email: string;
  role: string;
}

// ─── Status badge ──────────────────────────────────────────────────────────────

function ConsultStatusBadge({ status }: { status: ConsultationSummary['status'] }) {
  const styles: Record<string, string> = {
    pending: 'bg-yellow-100 text-yellow-800',
    positive: 'bg-green-100 text-green-800',
    negative: 'bg-red-100 text-red-800',
    expired: 'bg-gray-100 text-gray-500',
  };
  const labels: Record<string, string> = {
    pending: 'Bekliyor',
    positive: 'Olumlu',
    negative: 'Olumsuz',
    expired: 'Süresi Dolmuş',
  };
  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${styles[status] ?? 'bg-gray-100 text-gray-700'}`}
    >
      {labels[status] ?? status}
    </span>
  );
}

// ─── Add consultee modal ───────────────────────────────────────────────────────

function AddConsulteeModal({
  applicationId,
  onClose,
}: {
  applicationId: string;
  onClose: () => void;
}) {
  const queryClient = useQueryClient();
  const [userIds, setUserIds] = useState<string[]>(['', '']);
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: async (consultees: { user_id: string }[]) => {
      const res = await apiClient.post(`/applications/${applicationId}/consultations`, {
        consultees,
      });
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['consultations', applicationId] });
      queryClient.invalidateQueries({ queryKey: ['application', applicationId] });
      onClose();
    },
    onError: (err: unknown) => {
      if (err && typeof err === 'object' && 'response' in err) {
        const axiosErr = err as { response?: { data?: { error?: { message?: string } } } };
        setError(axiosErr.response?.data?.error?.message ?? 'Bir hata oluştu.');
      } else {
        setError('Bir hata oluştu.');
      }
    },
  });

  const handleAddField = () => setUserIds((prev) => [...prev, '']);
  const handleRemoveField = (i: number) =>
    setUserIds((prev) => prev.filter((_, idx) => idx !== i));
  const handleChange = (i: number, val: string) =>
    setUserIds((prev) => prev.map((v, idx) => (idx === i ? val : v)));

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    const filled = userIds.map((id) => id.trim()).filter(Boolean);
    if (filled.length < 2) {
      setError('En az 2 üye ID giriniz.');
      return;
    }
    mutation.mutate(filled.map((id) => ({ user_id: id })));
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-md p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Danışman Ekle</h2>

        <form onSubmit={handleSubmit} className="space-y-3">
          <p className="text-sm text-gray-500">
            Danışılacak üyelerin sistem kullanıcı ID&apos;lerini giriniz (en az 2).
          </p>

          {userIds.map((val, i) => (
            <div key={i} className="flex gap-2">
              <input
                type="text"
                value={val}
                onChange={(e) => handleChange(i, e.target.value)}
                placeholder={`Üye ID #${i + 1}`}
                className="flex-1 rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"
              />
              {userIds.length > 2 && (
                <button
                  type="button"
                  onClick={() => handleRemoveField(i)}
                  className="text-red-500 hover:text-red-700 text-sm px-1"
                >
                  ✕
                </button>
              )}
            </div>
          ))}

          <button
            type="button"
            onClick={handleAddField}
            className="text-sm text-blue-600 hover:underline"
          >
            + Daha fazla üye ekle
          </button>

          {error && (
            <div className="rounded-md bg-red-50 border border-red-200 px-3 py-2 text-sm text-red-700">
              {error}
            </div>
          )}

          <div className="flex gap-2 pt-2">
            <button
              type="submit"
              disabled={mutation.isPending}
              className="flex-1 bg-slate-800 text-white py-2 rounded-md text-sm font-medium hover:bg-slate-700 disabled:opacity-50 transition-colors"
            >
              {mutation.isPending ? 'Gönderiliyor…' : 'Danışma Taleplerini Gönder'}
            </button>
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded-md text-sm font-medium text-gray-600 hover:bg-gray-100 transition-colors"
            >
              İptal
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// ─── Panel ────────────────────────────────────────────────────────────────────

interface ConsultationPanelProps {
  applicationId: string;
  membershipType: string;
}

export function ConsultationPanel({ applicationId, membershipType }: ConsultationPanelProps) {
  const { user } = useAuthStore();
  const role = user?.role ?? '';
  const canManage = role === 'koordinator' || role === 'admin';
  const isConsultType = membershipType === 'profesyonel' || membershipType === 'öğrenci';

  const [showModal, setShowModal] = useState(false);

  const { data: consultations, isLoading } = useQuery<ConsultationSummary[]>({
    queryKey: ['consultations', applicationId],
    queryFn: async () => {
      const res = await apiClient.get(`/applications/${applicationId}/consultations`);
      return res.data.data ?? [];
    },
    enabled: isConsultType,
  });

  // Only shown for profesyonel / öğrenci
  if (!isConsultType) {
    return (
      <div className="rounded-lg border border-gray-200 p-5 text-sm text-gray-500">
        Danışma süreci yalnızca Profesyonel ve Öğrenci başvuruları için geçerlidir.
      </div>
    );
  }

  if (isLoading) {
    return <div className="text-sm text-gray-400 animate-pulse">Danışmalar yükleniyor…</div>;
  }

  const list = consultations ?? [];

  return (
    <div className="space-y-4">
      {/* Header row */}
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-gray-700">
          Danışılan Üyeler ({list.length})
        </h3>
        {canManage && list.length === 0 && (
          <button
            onClick={() => setShowModal(true)}
            className="text-sm font-medium text-blue-600 hover:underline"
          >
            + Danışman Ekle
          </button>
        )}
      </div>

      {/* Empty state */}
      {list.length === 0 && (
        <div className="rounded-lg border border-dashed border-gray-300 p-6 text-center text-sm text-gray-400">
          Henüz danışılan üye bulunmamaktadır.
          {canManage && (
            <p className="mt-1 text-xs">
              Danışma sürecini başlatmak için üye ekleyiniz.
            </p>
          )}
        </div>
      )}

      {/* Table */}
      {list.length > 0 && (
        <div className="overflow-hidden rounded-lg border border-gray-200">
          <table className="min-w-full divide-y divide-gray-200 text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Üye
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  E-posta
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Durum
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Tarih
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {list.map((c) => (
                <tr key={c.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-4 py-3 font-medium text-gray-900">{c.member_name}</td>
                  <td className="px-4 py-3 text-gray-500">{c.member_email}</td>
                  <td className="px-4 py-3">
                    <ConsultStatusBadge status={c.status} />
                  </td>
                  <td className="px-4 py-3 text-gray-400 text-xs">
                    {new Date(c.created_at).toLocaleDateString('tr-TR')}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {showModal && (
        <AddConsulteeModal
          applicationId={applicationId}
          onClose={() => setShowModal(false)}
        />
      )}
    </div>
  );
}
