'use client';

import { CheckCircle2, MinusCircle, XCircle, Users } from 'lucide-react';
import { VoteSummary, VoteResponse } from '@/lib/api/voting';
import { format } from 'date-fns';
import { tr } from 'date-fns/locale';

interface VoteSummaryProps {
  summary: VoteSummary;
  /** Whether the viewer can see individual vote details */
  canSeeDetails: boolean;
}

const stageLabel: Record<string, string> = {
  yk_prelim: 'YK Ön İnceleme',
  yik: 'YİK Değerlendirme',
  yk_final: 'YK Genel Kurul',
};

const voteTypeConfig = {
  approve: {
    label: 'Onay',
    icon: CheckCircle2,
    className: 'text-green-600 bg-green-50 border-green-200',
    badgeClass: 'bg-green-100 text-green-800',
  },
  abstain: {
    label: 'Çekimser',
    icon: MinusCircle,
    className: 'text-yellow-600 bg-yellow-50 border-yellow-200',
    badgeClass: 'bg-yellow-100 text-yellow-800',
  },
  reject: {
    label: 'Red (Veto)',
    icon: XCircle,
    className: 'text-red-600 bg-red-50 border-red-200',
    badgeClass: 'bg-red-100 text-red-800',
  },
} as const;

/**
 * VoteSummary renders:
 *  - A counts row (approve / abstain / reject / remaining)
 *  - Optionally a voter details table for yk / admin viewers
 */
export function VoteSummaryPanel({ summary, canSeeDetails }: VoteSummaryProps) {
  const remaining =
    summary.total_voters - summary.approved - summary.abstained - summary.rejected;

  return (
    <div className="space-y-4">
      {/* Stage header */}
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-slate-700">
          {stageLabel[summary.stage] ?? summary.stage} — Oy Özeti
        </h3>
        <span className="flex items-center gap-1 text-xs text-slate-500">
          <Users className="h-3.5 w-3.5" />
          {summary.total_voters} toplam seçmen
        </span>
      </div>

      {/* Counts row */}
      <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
        <StatCard
          icon={CheckCircle2}
          label="Onay"
          value={summary.approved}
          colorClass="text-green-600"
          bgClass="bg-green-50"
        />
        <StatCard
          icon={MinusCircle}
          label="Çekimser"
          value={summary.abstained}
          colorClass="text-yellow-600"
          bgClass="bg-yellow-50"
        />
        <StatCard
          icon={XCircle}
          label="Red"
          value={summary.rejected}
          colorClass="text-red-600"
          bgClass="bg-red-50"
        />
        <StatCard
          icon={Users}
          label="Bekleyen"
          value={remaining}
          colorClass="text-slate-500"
          bgClass="bg-slate-50"
        />
      </div>

      {/* Voter details table (yk / admin only) */}
      {canSeeDetails && summary.votes.length > 0 && (
        <div className="overflow-hidden rounded-lg border border-slate-200">
          <table className="w-full text-sm">
            <thead className="bg-slate-50">
              <tr>
                <th className="px-4 py-2 text-left font-medium text-slate-600">
                  Seçmen
                </th>
                <th className="px-4 py-2 text-left font-medium text-slate-600">
                  Karar
                </th>
                <th className="px-4 py-2 text-left font-medium text-slate-600">
                  Tarih
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {summary.votes.map((v) => (
                <VoteRow key={v.id} vote={v} />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

// ─── sub-components ───────────────────────────────────────────────────────────

function StatCard({
  icon: Icon,
  label,
  value,
  colorClass,
  bgClass,
}: {
  icon: React.ElementType;
  label: string;
  value: number;
  colorClass: string;
  bgClass: string;
}) {
  return (
    <div className={`flex items-center gap-2 rounded-lg border border-slate-200 ${bgClass} px-3 py-2`}>
      <Icon className={`h-4 w-4 ${colorClass}`} />
      <div>
        <p className={`text-lg font-bold ${colorClass}`}>{value}</p>
        <p className="text-xs text-slate-500">{label}</p>
      </div>
    </div>
  );
}

function VoteRow({ vote }: { vote: VoteResponse }) {
  const cfg = voteTypeConfig[vote.vote_type] ?? voteTypeConfig.abstain;
  const Icon = cfg.icon;

  return (
    <tr className="hover:bg-slate-50">
      <td className="px-4 py-2 font-medium text-slate-800">{vote.voter_name}</td>
      <td className="px-4 py-2">
        <span
          className={`inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs font-medium ${cfg.className}`}
        >
          <Icon className="h-3 w-3" />
          {cfg.label}
        </span>
        {vote.reason && (
          <p className="mt-0.5 text-xs text-slate-500 italic">
            &ldquo;{vote.reason}&rdquo;
          </p>
        )}
      </td>
      <td className="px-4 py-2 text-xs text-slate-400">
        {format(new Date(vote.created_at), 'dd MMM yyyy HH:mm', { locale: tr })}
      </td>
    </tr>
  );
}
