'use client';

import { useState, useCallback, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table';
import { FileText, ChevronLeft, ChevronRight, X, Eye } from 'lucide-react';
import { useLogs } from '@/lib/hooks/useLogs';
import { LogEntry, LogFilters } from '@/lib/api/logs';
import { useAuthStore } from '@/lib/store/auth.store';
import { PageHeader } from '@/components/ui/PageHeader';
import { EmptyState } from '@/components/ui/EmptyState';
import { TableSkeleton } from '@/components/ui/Skeletons';

// ─── Constants ────────────────────────────────────────────────────────────────

const ENTITY_TYPES = [
  { value: '', label: 'Tüm Varlıklar' },
  { value: 'application', label: 'Başvuru' },
  { value: 'user', label: 'Kullanıcı' },
  { value: 'reference', label: 'Referans' },
  { value: 'vote', label: 'Oy' },
  { value: 'consultation', label: 'Danışma' },
  { value: 'reputation_contact', label: 'İtibar Kişisi' },
  { value: 'web_publish_consent', label: 'Web Yayın Onayı' },
];

const PAGE_SIZE = 20;

// ─── Log Detail Drawer ────────────────────────────────────────────────────────

interface LogDetailDrawerProps {
  log: LogEntry;
  onClose: () => void;
  showActorName: boolean;
}

function LogDetailDrawer({ log, onClose, showActorName }: LogDetailDrawerProps) {
  return (
    <div className="fixed inset-0 z-50 flex">
      {/* Overlay */}
      <div className="flex-1 bg-black/40" onClick={onClose} />
      
      {/* Drawer */}
      <div className="w-full max-w-lg bg-white shadow-xl flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
          <h2 className="text-lg font-semibold text-gray-900">Log Detayı</h2>
          <button
            onClick={onClose}
            className="rounded-lg p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {/* Basic info */}
          <div className="space-y-3">
            <InfoRow label="Tarih" value={formatDateTime(log.created_at)} />
            <InfoRow label="İşlem" value={log.action} />
            <InfoRow label="Varlık Tipi" value={log.entity_type} />
            <InfoRow label="Varlık ID" value={log.entity_id} mono />
            <InfoRow label="Aktör Rol" value={log.actor_role} />
            {showActorName && log.actor_name && (
              <InfoRow label="Aktör" value={log.actor_name} />
            )}
            {log.ip_address && (
              <InfoRow label="IP Adresi" value={log.ip_address} mono />
            )}
          </div>

          {/* Metadata */}
          {log.metadata && Object.keys(log.metadata).length > 0 && (
            <div>
              <h3 className="text-sm font-semibold text-gray-700 mb-2">
                Metadata
              </h3>
              <pre className="rounded-lg bg-gray-50 p-4 text-xs text-gray-700 overflow-x-auto">
                {JSON.stringify(log.metadata, null, 2)}
              </pre>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function InfoRow({
  label,
  value,
  mono,
}: {
  label: string;
  value: string;
  mono?: boolean;
}) {
  return (
    <div className="flex justify-between gap-4">
      <span className="text-sm text-gray-500">{label}</span>
      <span
        className={`text-sm text-gray-900 text-right ${mono ? 'font-mono' : ''}`}
      >
        {value}
      </span>
    </div>
  );
}

function formatDateTime(dateStr: string) {
  return new Date(dateStr).toLocaleString('tr-TR', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

// ─── Table Content ────────────────────────────────────────────────────────────

function LogsContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const { user } = useAuthStore();
  const isAdmin = user?.role === 'admin';

  const [selectedLog, setSelectedLog] = useState<LogEntry | null>(null);

  const filters: LogFilters = {
    action: searchParams.get('action') || undefined,
    entity_type: searchParams.get('entity_type') || undefined,
    start_date: searchParams.get('start_date') || undefined,
    end_date: searchParams.get('end_date') || undefined,
    page: Number(searchParams.get('page') ?? '1'),
    page_size: PAGE_SIZE,
  };

  const { data, isLoading, isError } = useLogs(filters);

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
      router.push(`/logs?${params.toString()}`);
    },
    [searchParams, router]
  );

  // ─── Columns ──────────────────────────────────────────────────────────────

  const columns: ColumnDef<LogEntry>[] = [
    {
      accessorKey: 'created_at',
      header: 'Tarih',
      cell: ({ getValue }) => (
        <span className="text-sm text-gray-500 whitespace-nowrap">
          {formatDateTime(getValue() as string)}
        </span>
      ),
    },
    {
      accessorKey: 'actor_role',
      header: 'Aktör Rol',
      cell: ({ getValue }) => (
        <span className="inline-flex rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-700">
          {getValue() as string}
        </span>
      ),
    },
    ...(isAdmin
      ? [
          {
            accessorKey: 'actor_name',
            header: 'Aktör',
            cell: ({ getValue }: { getValue: () => unknown }) => (
              <span className="text-sm text-gray-700">
                {(getValue() as string) || '—'}
              </span>
            ),
          } as ColumnDef<LogEntry>,
        ]
      : []),
    {
      accessorKey: 'action',
      header: 'İşlem',
      cell: ({ getValue }) => (
        <span className="text-sm font-medium text-gray-900">
          {getValue() as string}
        </span>
      ),
    },
    {
      accessorKey: 'entity_type',
      header: 'Varlık',
      cell: ({ getValue }) => (
        <span className="text-sm text-gray-600">{getValue() as string}</span>
      ),
    },
    {
      accessorKey: 'entity_id',
      header: 'Varlık ID',
      cell: ({ getValue }) => (
        <span className="font-mono text-xs text-gray-500 truncate max-w-[120px] block">
          {getValue() as string}
        </span>
      ),
    },
    {
      accessorKey: 'ip_address',
      header: 'IP',
      cell: ({ getValue }) => (
        <span className="font-mono text-xs text-gray-400">
          {(getValue() as string) || '—'}
        </span>
      ),
    },
    {
      id: 'actions',
      header: '',
      cell: ({ row }) => (
        <button
          onClick={() => setSelectedLog(row.original)}
          className="p-1 text-gray-400 hover:text-gray-700 transition-colors"
        >
          <Eye className="h-4 w-4" />
        </button>
      ),
    },
  ];

  const table = useReactTable({
    data: data?.data ?? [],
    columns,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    pageCount: data?.total_pages ?? 1,
  });

  // ─── Render ───────────────────────────────────────────────────────────────

  return (
    <div className="p-6 space-y-4">
      <PageHeader
        title="Sistem Logları"
        description="Sistemde gerçekleştirilen tüm işlemlerin kaydı"
      />

      {/* Filters */}
      <div className="flex flex-wrap gap-3 items-center bg-white border border-gray-200 rounded-lg px-4 py-3">
        {/* Action search */}
        <input
          type="text"
          placeholder="İşlem ara..."
          defaultValue={filters.action ?? ''}
          onChange={(e) => setParam('action', e.target.value)}
          className="flex-1 min-w-[180px] text-sm border border-gray-300 rounded-md px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-blue-500"
        />

        {/* Entity type */}
        <select
          value={filters.entity_type ?? ''}
          onChange={(e) => setParam('entity_type', e.target.value)}
          className="text-sm border border-gray-300 rounded-md px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          {ENTITY_TYPES.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>

        {/* Start date */}
        <input
          type="date"
          value={filters.start_date ?? ''}
          onChange={(e) => setParam('start_date', e.target.value)}
          className="text-sm border border-gray-300 rounded-md px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-blue-500"
        />

        {/* End date */}
        <input
          type="date"
          value={filters.end_date ?? ''}
          onChange={(e) => setParam('end_date', e.target.value)}
          className="text-sm border border-gray-300 rounded-md px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      {/* Table */}
      <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
        {isError ? (
          <div className="p-8 text-center text-red-600 text-sm">
            Loglar yüklenirken hata oluştu.
          </div>
        ) : isLoading ? (
          <TableSkeleton rows={10} />
        ) : table.getRowModel().rows.length === 0 ? (
          <EmptyState
            icon={<FileText className="h-6 w-6" />}
            title="Log bulunamadı"
            description="Seçili filtrelere uygun kayıt bulunmamaktadır."
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
                      {flexRender(
                        header.column.columnDef.header,
                        header.getContext()
                      )}
                    </th>
                  ))}
                </tr>
              ))}
            </thead>
            <tbody className="divide-y divide-gray-100">
              {table.getRowModel().rows.map((row) => (
                <tr
                  key={row.id}
                  className="hover:bg-gray-50 transition-colors cursor-pointer"
                  onClick={() => setSelectedLog(row.original)}
                >
                  {row.getVisibleCells().map((cell) => (
                    <td key={cell.id} className="px-4 py-3">
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
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
            Toplam {data.total} kayıt — Sayfa {data.page} / {data.total_pages}
          </p>
          <div className="flex gap-2">
            <button
              disabled={data.page <= 1}
              onClick={() => setParam('page', String(data.page - 1))}
              className="inline-flex items-center gap-1 px-3 py-1.5 text-sm border border-gray-300 rounded-md disabled:opacity-40 hover:bg-gray-50"
            >
              <ChevronLeft className="h-4 w-4" />
              Önceki
            </button>
            <button
              disabled={data.page >= data.total_pages}
              onClick={() => setParam('page', String(data.page + 1))}
              className="inline-flex items-center gap-1 px-3 py-1.5 text-sm border border-gray-300 rounded-md disabled:opacity-40 hover:bg-gray-50"
            >
              Sonraki
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>
        </div>
      )}

      {/* Detail drawer */}
      {selectedLog && (
        <LogDetailDrawer
          log={selectedLog}
          onClose={() => setSelectedLog(null)}
          showActorName={isAdmin}
        />
      )}
    </div>
  );
}

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function LogsPage() {
  return (
    <Suspense fallback={<div className="p-6"><TableSkeleton rows={10} /></div>}>
      <LogsContent />
    </Suspense>
  );
}
