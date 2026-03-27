'use client';

import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { Suspense, useCallback, useMemo } from 'react';
import { Download, FileText, Clock, CheckCircle2, XCircle } from 'lucide-react';
import { StatusBadge } from '../../../components/applications/StatusBadge';
import { useAuthStore } from '../../../lib/store/auth.store';
import { TableSkeleton, StatsSkeleton } from '../../../components/ui/Skeletons';
import { EmptyState } from '../../../components/ui/EmptyState';
import { showToast } from '../../../components/ui/Toaster';
import {
  ApplicationFilters,
  ApplicationSummary,
  ApplicationStatus,
  MembershipType,
  getApplications,
} from '../../../lib/api/applications';
import { useApplications } from '../../../lib/hooks/useApplications';

// ─── Constants ────────────────────────────────────────────────────────────────

const MEMBERSHIP_TYPES: { value: MembershipType | ''; label: string }[] = [
  { value: '', label: 'Tüm Tipler' },
  { value: 'asil', label: 'Asil' },
  { value: 'akademik', label: 'Akademik' },
  { value: 'profesyonel', label: 'Profesyonel' },
  { value: 'öğrenci', label: 'Öğrenci' },
  { value: 'onursal', label: 'Onursal' },
];

const STATUS_OPTIONS: { value: ApplicationStatus | ''; label: string }[] = [
  { value: '', label: 'Tüm Durumlar' },
  { value: 'başvuru_alındı', label: 'Başvuru Alındı' },
  { value: 'referans_bekleniyor', label: 'Referans Bekleniyor' },
  { value: 'referans_tamamlandı', label: 'Referans Tamamlandı' },
  { value: 'referans_red', label: 'Referans Red' },
  { value: 'yk_ön_incelemede', label: 'YK Ön İnceleme' },
  { value: 'ön_onaylandı', label: 'Ön Onaylandı' },
  { value: 'yk_red', label: 'YK Red' },
  { value: 'itibar_taramasında', label: 'İtibar Taramasında' },
  { value: 'itibar_temiz', label: 'İtibar Temiz' },
  { value: 'itibar_red', label: 'İtibar Red' },
  { value: 'danışma_sürecinde', label: 'Danışma Sürecinde' },
  { value: 'danışma_red', label: 'Danışma Red' },
  { value: 'öneri_alındı', label: 'Öneri Alındı' },
  { value: 'yik_değerlendirmede', label: 'YİK Değerlendirmede' },
  { value: 'yik_red', label: 'YİK Red' },
  { value: 'gündemde', label: 'Gündemde' },
  { value: 'kabul', label: 'Kabul' },
  { value: 'reddedildi', label: 'Reddedildi' },
];

const MEMBERSHIP_LABELS: Record<MembershipType, string> = {
  asil: 'Asil',
  akademik: 'Akademik',
  profesyonel: 'Profesyonel',
  'öğrenci': 'Öğrenci',
  onursal: 'Onursal',
};

// ─── Columns ──────────────────────────────────────────────────────────────────

const columns: ColumnDef<ApplicationSummary>[] = [
  {
    accessorKey: 'applicant_name',
    header: 'Ad Soyad',
    cell: ({ row }) => (
      <Link
        href={`/applications/${row.original.id}`}
        className="font-medium text-blue-700 hover:underline"
      >
        {row.getValue('applicant_name')}
      </Link>
    ),
  },
  {
    accessorKey: 'applicant_email',
    header: 'E-posta',
    cell: ({ getValue }) => (
      <span className="text-gray-600 text-sm">{getValue() as string}</span>
    ),
  },
  {
    accessorKey: 'membership_type',
    header: 'Üyelik Tipi',
    cell: ({ getValue }) => (
      <span className="text-sm text-gray-700">
        {MEMBERSHIP_LABELS[getValue() as MembershipType] ?? getValue() as string}
      </span>
    ),
  },
  {
    accessorKey: 'status',
    header: 'Durum',
    cell: ({ getValue }) => <StatusBadge status={getValue() as ApplicationStatus} />,
  },
  {
    accessorKey: 'created_at',
    header: 'Başvuru Tarihi',
    cell: ({ getValue }) => (
      <span className="text-sm text-gray-500">
        {new Date(getValue() as string).toLocaleDateString('tr-TR')}
      </span>
    ),
  },
  {
    id: 'actions',
    header: '',
    cell: ({ row }) => (
      <Link
        href={`/applications/${row.original.id}`}
        className="text-xs text-blue-600 hover:text-blue-800 font-medium"
      >
        Detay →
      </Link>
    ),
  },
];

// ─── Inner Component (needs useSearchParams) ─────────────────────────────────

function ApplicationsContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const { user } = useAuthStore();
  const isAdmin = user?.role === 'admin';

  const filters: ApplicationFilters = {
    membership_type: searchParams.get('membership_type') || undefined,
    status: searchParams.get('status') || undefined,
    search: searchParams.get('search') || undefined,
    page: Number(searchParams.get('page') ?? '1'),
    page_size: 20,
  };

  const { data, isLoading, isError } = useApplications(filters);

  // Calculate stats from data
  const stats = useMemo(() => {
    if (!data) return { total: 0, pending: 0, accepted: 0, rejected: 0 };
    
    const applications = data.data;
    const pendingStatuses: ApplicationStatus[] = [
      'başvuru_alındı', 'referans_bekleniyor', 'referans_tamamlandı',
      'yk_ön_incelemede', 'ön_onaylandı', 'itibar_taramasında',
      'itibar_temiz', 'danışma_sürecinde', 'öneri_alındı',
      'yik_değerlendirmede', 'gündemde'
    ];
    const rejectedStatuses: ApplicationStatus[] = [
      'referans_red', 'yk_red', 'itibar_red', 'danışma_red', 'yik_red', 'reddedildi'
    ];

    return {
      total: data.total,
      pending: applications.filter(a => pendingStatuses.includes(a.status)).length,
      accepted: applications.filter(a => a.status === 'kabul').length,
      rejected: applications.filter(a => rejectedStatuses.includes(a.status)).length,
    };
  }, [data]);

  const handleExportCSV = async () => {
    try {
      // Fetch all applications (no pagination)
      const allData = await getApplications({ ...filters, page: 1, page_size: 10000 });
      
      // Create CSV content
      const headers = ['Ad Soyad', 'E-posta', 'Üyelik Tipi', 'Durum', 'Başvuru Tarihi'];
      const rows = allData.data.map(app => [
        app.applicant_name,
        app.applicant_email,
        MEMBERSHIP_LABELS[app.membership_type] || app.membership_type,
        app.status,
        new Date(app.created_at).toLocaleDateString('tr-TR'),
      ]);

      const csvContent = [
        headers.join(','),
        ...rows.map(row => row.map(cell => `"${cell}"`).join(','))
      ].join('\n');

      // Download file
      const blob = new Blob(['\uFEFF' + csvContent], { type: 'text/csv;charset=utf-8;' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `basvurular_${new Date().toISOString().split('T')[0]}.csv`;
      link.click();
      URL.revokeObjectURL(url);
      
      showToast('CSV dosyası indirildi.', 'success');
    } catch {
      showToast('CSV dışa aktarılamadı.', 'error');
    }
  };

  const setParam = useCallback(
    (key: string, value: string) => {
      const params = new URLSearchParams(searchParams.toString());
      if (value) {
        params.set(key, value);
      } else {
        params.delete(key);
      }
      // Reset to page 1 when filter changes
      if (key !== 'page') params.set('page', '1');
      router.push(`/applications?${params.toString()}`);
    },
    [searchParams, router]
  );

  const table = useReactTable({
    data: data?.data ?? [],
    columns,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    pageCount: data?.total_pages ?? 1,
  });

  return (
    <div className="p-6 space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between flex-wrap gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Başvurular</h1>
          {data && (
            <p className="text-sm text-gray-500 mt-0.5">
              Toplam {data.total} başvuru
            </p>
          )}
        </div>
        {isAdmin && (
          <button
            onClick={handleExportCSV}
            className="inline-flex items-center gap-2 rounded-lg bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-800 transition-colors"
          >
            <Download className="h-4 w-4" />
            CSV İndir
          </button>
        )}
      </div>

      {/* Stats Row */}
      {isLoading ? (
        <StatsSkeleton />
      ) : (
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
          <StatCard
            icon={<FileText className="h-5 w-5" />}
            label="Toplam"
            value={stats.total}
            color="text-gray-700"
            bg="bg-gray-50"
          />
          <StatCard
            icon={<Clock className="h-5 w-5" />}
            label="Bekleyen"
            value={stats.pending}
            color="text-yellow-700"
            bg="bg-yellow-50"
          />
          <StatCard
            icon={<CheckCircle2 className="h-5 w-5" />}
            label="Kabul"
            value={stats.accepted}
            color="text-green-700"
            bg="bg-green-50"
          />
          <StatCard
            icon={<XCircle className="h-5 w-5" />}
            label="Reddedildi"
            value={stats.rejected}
            color="text-red-700"
            bg="bg-red-50"
          />
        </div>
      )}

      {/* Filters */}
      <div className="flex flex-wrap gap-3 items-center bg-white border border-gray-200 rounded-lg px-4 py-3">
        {/* Search */}
        <input
          type="text"
          placeholder="İsim veya e-posta ara..."
          defaultValue={filters.search ?? ''}
          onChange={(e) => setParam('search', e.target.value)}
          className="flex-1 min-w-[200px] text-sm border border-gray-300 rounded-md px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        {/* Membership type */}
        <select
          value={filters.membership_type ?? ''}
          onChange={(e) => setParam('membership_type', e.target.value)}
          className="text-sm border border-gray-300 rounded-md px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          {MEMBERSHIP_TYPES.map((o) => (
            <option key={o.value} value={o.value}>{o.label}</option>
          ))}
        </select>
        {/* Status */}
        <select
          value={filters.status ?? ''}
          onChange={(e) => setParam('status', e.target.value)}
          className="text-sm border border-gray-300 rounded-md px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          {STATUS_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>{o.label}</option>
          ))}
        </select>
      </div>

      {/* Table */}
      <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
        {isError ? (
          <div className="p-8 text-center text-red-600 text-sm">
            Veriler yüklenirken hata oluştu.
          </div>
        ) : isLoading ? (
          <TableSkeleton rows={10} />
        ) : table.getRowModel().rows.length === 0 ? (
          <EmptyState
            icon={<FileText className="h-6 w-6" />}
            title="Başvuru bulunamadı"
            description="Seçili filtrelere uygun başvuru bulunmamaktadır."
          />
        ) : (
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-200">
              {table.getHeaderGroups().map((hg) => (
                <tr key={hg.id}>
                  {hg.headers.map((header) => (
                    <th
                      key={header.id}
                      className="text-left text-xs font-semibold text-gray-500 uppercase tracking-wide px-4 py-3"
                    >
                      {flexRender(header.column.columnDef.header, header.getContext())}
                    </th>
                  ))}
                </tr>
              ))}
            </thead>
            <tbody className="divide-y divide-gray-100">
              {table.getRowModel().rows.map((row) => (
                <tr key={row.id} className="hover:bg-gray-50 transition-colors">
                  {row.getVisibleCells().map((cell) => (
                    <td key={cell.id} className="px-4 py-3">
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Pagination */}
      {data && data.total_pages > 1 && (
        <div className="flex items-center justify-between pt-2">
          <p className="text-sm text-gray-500">
            Sayfa {data.page} / {data.total_pages}
          </p>
          <div className="flex gap-2">
            <button
              disabled={data.page <= 1}
              onClick={() => setParam('page', String(data.page - 1))}
              className="px-3 py-1.5 text-sm border border-gray-300 rounded-md disabled:opacity-40 hover:bg-gray-50"
            >
              ← Önceki
            </button>
            <button
              disabled={data.page >= data.total_pages}
              onClick={() => setParam('page', String(data.page + 1))}
              className="px-3 py-1.5 text-sm border border-gray-300 rounded-md disabled:opacity-40 hover:bg-gray-50"
            >
              Sonraki →
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

// ─── Page ─────────────────────────────────────────────────────────────────────

function StatCard({
  icon,
  label,
  value,
  color,
  bg,
}: {
  icon: React.ReactNode;
  label: string;
  value: number;
  color: string;
  bg: string;
}) {
  return (
    <div className={`flex items-center gap-3 rounded-lg border border-gray-200 ${bg} px-4 py-3`}>
      <div className={color}>{icon}</div>
      <div>
        <p className={`text-xl font-bold ${color}`}>{value}</p>
        <p className="text-xs text-gray-500">{label}</p>
      </div>
    </div>
  );
}

export default function ApplicationsPage() {
  return (
    <Suspense fallback={<div className="p-6"><TableSkeleton rows={10} /></div>}>
      <ApplicationsContent />
    </Suspense>
  );
}
