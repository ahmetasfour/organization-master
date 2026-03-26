import { ApplicationStatus } from '../../lib/api/applications';
import { cn } from '../../lib/utils';

interface StatusBadgeProps {
  status: ApplicationStatus;
  className?: string;
}

// Color mapping per status
const statusConfig: Record<
  ApplicationStatus,
  { label: string; className: string }
> = {
  'başvuru_alındı':       { label: 'Başvuru Alındı',       className: 'bg-blue-100 text-blue-800' },
  'referans_bekleniyor':  { label: 'Referans Bekleniyor',  className: 'bg-yellow-100 text-yellow-800' },
  'referans_tamamlandı':  { label: 'Referans Tamamlandı',  className: 'bg-green-100 text-green-700' },
  'referans_red':         { label: 'Referans Red',         className: 'bg-red-100 text-red-800' },
  'yk_ön_incelemede':     { label: 'YK Ön İnceleme',       className: 'bg-yellow-100 text-yellow-800' },
  'ön_onaylandı':         { label: 'Ön Onaylandı',         className: 'bg-green-100 text-green-700' },
  'yk_red':               { label: 'YK Red',               className: 'bg-red-100 text-red-800' },
  'itibar_taramasında':   { label: 'İtibar Taraması',       className: 'bg-yellow-100 text-yellow-800' },
  'itibar_temiz':         { label: 'İtibar Temiz',          className: 'bg-green-100 text-green-700' },
  'itibar_red':           { label: 'İtibar Red',            className: 'bg-red-100 text-red-800' },
  'danışma_sürecinde':    { label: 'Danışma Sürecinde',    className: 'bg-yellow-100 text-yellow-800' },
  'danışma_red':          { label: 'Danışma Red',          className: 'bg-red-100 text-red-800' },
  'öneri_alındı':         { label: 'Öneri Alındı',         className: 'bg-blue-100 text-blue-800' },
  'yik_değerlendirmede':  { label: 'YİK Değerlendirme',   className: 'bg-yellow-100 text-yellow-800' },
  'yik_red':              { label: 'YİK Red',              className: 'bg-red-100 text-red-800' },
  'gündemde':             { label: 'Gündemde',             className: 'bg-purple-100 text-purple-800' },
  'kabul':                { label: 'Kabul',                className: 'bg-green-600 text-white' },
  'reddedildi':           { label: 'Reddedildi',           className: 'bg-red-600 text-white' },
};

export function StatusBadge({ status, className }: StatusBadgeProps) {
  const config = statusConfig[status] ?? {
    label: status,
    className: 'bg-gray-100 text-gray-700',
  };

  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
        config.className,
        className
      )}
    >
      {config.label}
    </span>
  );
}
