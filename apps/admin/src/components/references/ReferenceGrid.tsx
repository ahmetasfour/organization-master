'use client';

import { useState } from 'react';
import { Mail, Check, X, HelpCircle, Loader2, RefreshCw } from 'lucide-react';
import { useReferences, useResendReferenceEmail } from '@/lib/hooks/useReferences';
import { useAuthStore } from '@/lib/store/auth.store';
import { showToast } from '@/components/ui/Toaster';
import type { ReferenceStatus } from '@/lib/api/references';

interface ReferenceGridProps {
  applicationId: string;
}

const statusConfig: Record<
  ReferenceStatus,
  { label: string; icon: React.ElementType; className: string }
> = {
  pending: {
    label: 'Bekliyor',
    icon: Mail,
    className: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  },
  positive: {
    label: 'Olumlu',
    icon: Check,
    className: 'bg-green-100 text-green-800 border-green-200',
  },
  negative: {
    label: 'Olumsuz',
    icon: X,
    className: 'bg-red-100 text-red-800 border-red-200',
  },
  unknown: {
    label: 'Tanımıyor',
    icon: HelpCircle,
    className: 'bg-gray-100 text-gray-700 border-gray-200',
  },
};

function StatusBadge({ status }: { status: ReferenceStatus }) {
  const config = statusConfig[status];
  const Icon = config.icon;

  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full border px-2.5 py-0.5 text-xs font-medium ${config.className}`}
    >
      <Icon className="h-3 w-3" />
      {config.label}
    </span>
  );
}

export function ReferenceGrid({ applicationId }: ReferenceGridProps) {
  const { user } = useAuthStore();
  const role = user?.role ?? '';
  const canResend = role === 'koordinator' || role === 'admin';

  const { data, isLoading, isError } = useReferences(applicationId);
  const resendMutation = useResendReferenceEmail(applicationId);

  const [resendingId, setResendingId] = useState<string | null>(null);

  const handleResend = async (referenceId: string) => {
    setResendingId(referenceId);
    try {
      await resendMutation.mutateAsync(referenceId);
      showToast('Referans e-postası yeniden gönderildi.', 'success');
    } catch {
      showToast('E-posta gönderilemedi. Lütfen tekrar deneyin.', 'error');
    } finally {
      setResendingId(null);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 py-8 text-sm text-gray-400">
        <Loader2 className="h-4 w-4 animate-spin" />
        Referanslar yükleniyor...
      </div>
    );
  }

  if (isError || !data) {
    return (
      <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700">
        Referans bilgileri yüklenirken bir hata oluştu.
      </div>
    );
  }

  const { references, total, responded, positive, negative, unknown } = data;

  if (references.length === 0) {
    return (
      <div className="rounded-lg border-2 border-dashed border-gray-200 p-8 text-center">
        <Mail className="mx-auto h-10 w-10 text-gray-300" />
        <p className="mt-2 text-sm font-medium text-gray-600">
          Henüz referans eklenmemiş
        </p>
        <p className="mt-1 text-xs text-gray-400">
          Bu başvuru için referans bilgisi bulunmamaktadır.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Stats row */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-5">
        <StatCard label="Toplam" value={total} color="text-gray-700" bg="bg-gray-50" />
        <StatCard label="Yanıtlanan" value={responded} color="text-blue-700" bg="bg-blue-50" />
        <StatCard label="Olumlu" value={positive} color="text-green-700" bg="bg-green-50" />
        <StatCard label="Olumsuz" value={negative} color="text-red-700" bg="bg-red-50" />
        <StatCard label="Tanımıyor" value={unknown} color="text-gray-600" bg="bg-gray-50" />
      </div>

      {/* References table */}
      <div className="overflow-hidden rounded-lg border border-gray-200">
        <table className="min-w-full divide-y divide-gray-200 text-sm">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                Referans
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                E-posta
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                Durum
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                Yanıt Tarihi
              </th>
              {canResend && (
                <th className="px-4 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500">
                  İşlem
                </th>
              )}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100 bg-white">
            {references.map((ref) => (
              <tr key={ref.id} className="hover:bg-gray-50 transition-colors">
                <td className="whitespace-nowrap px-4 py-3 font-medium text-gray-900">
                  {ref.referee_name}
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-gray-500">
                  {ref.referee_email}
                </td>
                <td className="whitespace-nowrap px-4 py-3">
                  <StatusBadge status={ref.status} />
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-gray-500">
                  {ref.responded_at
                    ? new Date(ref.responded_at).toLocaleDateString('tr-TR', {
                        day: '2-digit',
                        month: '2-digit',
                        year: 'numeric',
                        hour: '2-digit',
                        minute: '2-digit',
                      })
                    : '—'}
                </td>
                {canResend && (
                  <td className="whitespace-nowrap px-4 py-3 text-right">
                    {ref.status === 'pending' && (
                      <button
                        onClick={() => handleResend(ref.id)}
                        disabled={resendingId === ref.id}
                        className="inline-flex items-center gap-1.5 rounded-md bg-gray-100 px-2.5 py-1 text-xs font-medium text-gray-700 transition hover:bg-gray-200 disabled:opacity-50"
                      >
                        {resendingId === ref.id ? (
                          <Loader2 className="h-3 w-3 animate-spin" />
                        ) : (
                          <RefreshCw className="h-3 w-3" />
                        )}
                        Tekrar Gönder
                      </button>
                    )}
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function StatCard({
  label,
  value,
  color,
  bg,
}: {
  label: string;
  value: number;
  color: string;
  bg: string;
}) {
  return (
    <div className={`rounded-lg ${bg} px-3 py-2`}>
      <p className={`text-lg font-bold ${color}`}>{value}</p>
      <p className="text-xs text-gray-500">{label}</p>
    </div>
  );
}
