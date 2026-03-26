'use client';

import { useState } from 'react';
import { CheckCircle2, MinusCircle, XCircle, Loader2 } from 'lucide-react';
import { useCastVote, useVotes } from '@/lib/hooks/useVoting';
import { VoteStage, VoteType } from '@/lib/api/voting';
import { VoteSummaryPanel } from './VoteSummary';
import { VetoAlert } from './VetoAlert';

interface VotePanelProps {
  applicationId: string;
  applicantName: string;
  /** Application status at the time the panel is rendered */
  applicationStatus: string;
  /** Rejection reason, if terminated */
  rejectionReason?: string;
  stage: VoteStage;
  /** The viewer's role — determines whether vote details are visible */
  viewerRole: string;
  viewerId: string;
}

const stageLabel: Record<VoteStage, string> = {
  yk_prelim: 'YK Ön İnceleme Oylaması',
  yik: 'YİK Değerlendirme Oylaması',
  yk_final: 'YK Genel Kurul Oylaması',
};

/**
 * VotePanel renders:
 *  1. Vote summary (counts + optionally voter details)
 *  2. Veto alert if application is terminated
 *  3. Three vote buttons (Onay / Çekimser / Red) with a confirmation modal
 *     Reject requires a 20+ character reason. Disappears once the viewer voted.
 */
export function VotePanel({
  applicationId,
  applicantName,
  applicationStatus,
  rejectionReason,
  stage,
  viewerRole,
  viewerId,
}: VotePanelProps) {
  const [pending, setPending] = useState<VoteType | null>(null);
  const [reason, setReason] = useState('');
  const [reasonError, setReasonError] = useState('');
  const [apiError, setApiError] = useState('');
  const [showConfirm, setShowConfirm] = useState(false);
  const [selectedVote, setSelectedVote] = useState<VoteType | null>(null);

  const { data: summary, isLoading } = useVotes(applicationId, stage);
  const castVoteMutation = useCastVote(applicationId, stage);

  const isTerminated = summary?.is_terminated ?? false;
  const canSeeDetails = viewerRole === 'yk' || viewerRole === 'admin';

  // Check if this viewer already voted
  const alreadyVoted =
    summary?.votes.some((v) => v.voter_id === viewerId) ?? false;

  // Determine whether voting is still open for this viewer
  const votingOpen = !isTerminated && !alreadyVoted;

  const openConfirm = (vt: VoteType) => {
    setSelectedVote(vt);
    setReason('');
    setReasonError('');
    setApiError('');
    setShowConfirm(true);
  };

  const closeConfirm = () => {
    setShowConfirm(false);
    setSelectedVote(null);
    setReason('');
    setReasonError('');
    setApiError('');
  };

  const handleConfirm = async () => {
    if (!selectedVote) return;

    if (selectedVote === 'reject') {
      if (reason.trim().length < 20) {
        setReasonError('Red gerekçesi en az 20 karakter olmalıdır.');
        return;
      }
    }

    setPending(selectedVote);
    setApiError('');
    try {
      await castVoteMutation.mutateAsync({
        vote_type: selectedVote,
        reason: selectedVote === 'reject' ? reason.trim() : undefined,
      });
      closeConfirm();
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message ?? 'Bir hata oluştu.';
      setApiError(msg);
    } finally {
      setPending(null);
    }
  };

  // ─── Loading state ────────────────────────────────────────────────────────

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 py-6 text-sm text-slate-400">
        <Loader2 className="h-4 w-4 animate-spin" />
        Oylama durumu yükleniyor…
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Section header */}
      <h2 className="text-base font-semibold text-slate-800">
        {stageLabel[stage]}
      </h2>

      {/* Veto alert (when terminated) */}
      {isTerminated && (
        <VetoAlert
          applicantName={applicantName}
          reason={rejectionReason}
          viewerRole={viewerRole}
        />
      )}

      {/* Vote summary */}
      {summary && (
        <VoteSummaryPanel summary={summary} canSeeDetails={canSeeDetails} />
      )}

      {/* Already voted notice */}
      {!isTerminated && alreadyVoted && (
        <div className="rounded-lg border border-blue-200 bg-blue-50 px-4 py-3 text-sm text-blue-700">
          Bu aşamada oyunuzu kullandınız. Sonuç yukarıda görüntülenmektedir.
        </div>
      )}

      {/* Vote buttons */}
      {votingOpen && (
        <div className="space-y-3">
          <p className="text-sm font-medium text-slate-600">Oyunuzu kullanın:</p>
          <div className="flex flex-wrap gap-3">
            <button
              onClick={() => openConfirm('approve')}
              className="inline-flex items-center gap-2 rounded-lg border border-green-200 bg-green-50 px-4 py-2 text-sm font-medium text-green-700 transition hover:bg-green-100 disabled:opacity-50"
            >
              <CheckCircle2 className="h-4 w-4" />
              Onaylıyorum
            </button>
            <button
              onClick={() => openConfirm('abstain')}
              className="inline-flex items-center gap-2 rounded-lg border border-yellow-200 bg-yellow-50 px-4 py-2 text-sm font-medium text-yellow-700 transition hover:bg-yellow-100 disabled:opacity-50"
            >
              <MinusCircle className="h-4 w-4" />
              Çekimser
            </button>
            <button
              onClick={() => openConfirm('reject')}
              className="inline-flex items-center gap-2 rounded-lg border border-red-200 bg-red-50 px-4 py-2 text-sm font-medium text-red-700 transition hover:bg-red-100 disabled:opacity-50"
            >
              <XCircle className="h-4 w-4" />
              Reddediyorum (Veto)
            </button>
          </div>
        </div>
      )}

      {/* Confirmation modal */}
      {showConfirm && selectedVote && (
        <ConfirmDialog
          voteType={selectedVote}
          applicantName={applicantName}
          reason={reason}
          reasonError={reasonError}
          apiError={apiError}
          isPending={pending === selectedVote}
          onReasonChange={(v) => {
            setReason(v);
            if (v.length >= 20) setReasonError('');
          }}
          onConfirm={handleConfirm}
          onCancel={closeConfirm}
        />
      )}
    </div>
  );
}

// ─── ConfirmDialog ────────────────────────────────────────────────────────────

const voteLabels: Record<VoteType, { label: string; desc: string; color: string }> = {
  approve: {
    label: 'Onaylıyorum',
    desc: 'Adayın üyeliğini onaylamak istediğinizi doğrulayın.',
    color: 'text-green-700',
  },
  abstain: {
    label: 'Çekimser',
    desc: 'Bu aşamada çekimser kalmak istediğinizi doğrulayın.',
    color: 'text-yellow-700',
  },
  reject: {
    label: 'Reddediyorum (Veto)',
    desc: 'DİKKAT: Bu işlem geri alınamaz. Başvuru kalıcı olarak reddedilecektir.',
    color: 'text-red-700',
  },
};

function ConfirmDialog({
  voteType,
  applicantName,
  reason,
  reasonError,
  apiError,
  isPending,
  onReasonChange,
  onConfirm,
  onCancel,
}: {
  voteType: VoteType;
  applicantName: string;
  reason: string;
  reasonError: string;
  apiError: string;
  isPending: boolean;
  onReasonChange: (v: string) => void;
  onConfirm: () => void;
  onCancel: () => void;
}) {
  const cfg = voteLabels[voteType];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
        <h3 className={`text-base font-semibold ${cfg.color}`}>{cfg.label}</h3>
        <p className="mt-2 text-sm text-slate-600">
          <span className="font-medium">{applicantName}</span> için: {cfg.desc}
        </p>

        {/* Reason textarea for reject */}
        {voteType === 'reject' && (
          <div className="mt-4 space-y-1">
            <label className="text-xs font-medium text-slate-600">
              Red Gerekçesi <span className="text-red-500">*</span>
            </label>
            <textarea
              rows={3}
              value={reason}
              onChange={(e) => onReasonChange(e.target.value)}
              placeholder="Reddetme gerekçenizi giriniz (en az 20 karakter)…"
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none placeholder:text-slate-400 focus:border-red-400 focus:ring-1 focus:ring-red-400"
            />
            <div className="flex justify-between">
              {reasonError ? (
                <p className="text-xs text-red-600">{reasonError}</p>
              ) : (
                <span />
              )}
              <span className="text-xs text-slate-400">{reason.length} / 20+</span>
            </div>
          </div>
        )}

        {/* API error */}
        {apiError && (
          <p className="mt-3 rounded-md bg-red-50 px-3 py-2 text-xs text-red-600">
            {apiError}
          </p>
        )}

        {/* Actions */}
        <div className="mt-5 flex justify-end gap-2">
          <button
            onClick={onCancel}
            disabled={isPending}
            className="rounded-lg border border-slate-200 px-4 py-2 text-sm text-slate-600 hover:bg-slate-50 disabled:opacity-50"
          >
            İptal
          </button>
          <button
            onClick={onConfirm}
            disabled={isPending}
            className={`inline-flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium text-white disabled:opacity-50 ${
              voteType === 'reject'
                ? 'bg-red-600 hover:bg-red-700'
                : voteType === 'approve'
                ? 'bg-green-600 hover:bg-green-700'
                : 'bg-yellow-500 hover:bg-yellow-600'
            }`}
          >
            {isPending && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
            Onayla
          </button>
        </div>
      </div>
    </div>
  );
}
