import { ApplicationStatus, MembershipType } from '../../lib/api/applications';
import { cn } from '../../lib/utils';

interface Step {
  status: ApplicationStatus;
  label: string;
}

// Flow definitions per membership type
const flows: Record<MembershipType, Step[]> = {
  asil: [
    { status: 'başvuru_alındı',       label: 'Başvuru Alındı' },
    { status: 'referans_bekleniyor',  label: 'Referans Bekleniyor' },
    { status: 'referans_tamamlandı',  label: 'Referans Tamamlandı' },
    { status: 'yk_ön_incelemede',     label: 'YK Ön İnceleme' },
    { status: 'ön_onaylandı',         label: 'Ön Onaylandı' },
    { status: 'itibar_taramasında',   label: 'İtibar Taraması' },
    { status: 'itibar_temiz',         label: 'İtibar Temiz' },
    { status: 'gündemde',             label: 'Gündemde' },
    { status: 'kabul',                label: 'Kabul' },
  ],
  akademik: [
    { status: 'başvuru_alındı',       label: 'Başvuru Alındı' },
    { status: 'referans_bekleniyor',  label: 'Referans Bekleniyor' },
    { status: 'referans_tamamlandı',  label: 'Referans Tamamlandı' },
    { status: 'yk_ön_incelemede',     label: 'YK Ön İnceleme' },
    { status: 'ön_onaylandı',         label: 'Ön Onaylandı' },
    { status: 'itibar_taramasında',   label: 'İtibar Taraması' },
    { status: 'itibar_temiz',         label: 'İtibar Temiz' },
    { status: 'gündemde',             label: 'Gündemde' },
    { status: 'kabul',                label: 'Kabul' },
  ],
  profesyonel: [
    { status: 'başvuru_alındı',       label: 'Başvuru Alındı' },
    { status: 'danışma_sürecinde',    label: 'Danışma Süreci' },
    { status: 'gündemde',             label: 'Gündemde' },
    { status: 'kabul',                label: 'Kabul' },
  ],
  'öğrenci': [
    { status: 'başvuru_alındı',       label: 'Başvuru Alındı' },
    { status: 'danışma_sürecinde',    label: 'Danışma Süreci' },
    { status: 'gündemde',             label: 'Gündemde' },
    { status: 'kabul',                label: 'Kabul' },
  ],
  onursal: [
    { status: 'öneri_alındı',         label: 'Öneri Alındı' },
    { status: 'yk_ön_incelemede',     label: 'YK Ön İnceleme' },
    { status: 'ön_onaylandı',         label: 'Ön Onaylandı' },
    { status: 'yik_değerlendirmede',  label: 'YİK Değerlendirme' },
    { status: 'gündemde',             label: 'Gündemde' },
    { status: 'kabul',                label: 'Kabul' },
  ],
};

const terminalRed: ApplicationStatus[] = [
  'referans_red', 'yk_red', 'itibar_red', 'danışma_red', 'yik_red', 'reddedildi',
];

interface StatusTimelineProps {
  membershipType: MembershipType;
  currentStatus: ApplicationStatus;
}

export function StatusTimeline({ membershipType, currentStatus }: StatusTimelineProps) {
  const steps = flows[membershipType] ?? flows['asil'];
  const isRejected = terminalRed.includes(currentStatus);

  // Find current step index
  const currentIndex = steps.findIndex((s) => s.status === currentStatus);

  return (
    <div className="flex items-start gap-0 overflow-x-auto pb-2">
      {steps.map((step, idx) => {
        const isCompleted = !isRejected && idx < currentIndex;
        const isCurrent = step.status === currentStatus;
        const isPending = !isRejected && idx > currentIndex;

        return (
          <div key={step.status} className="flex flex-col items-center min-w-[80px] flex-1">
            {/* Connector line + circle row */}
            <div className="flex items-center w-full">
              {/* Left line */}
              <div className={cn(
                'flex-1 h-0.5',
                idx === 0 ? 'bg-transparent' : isCompleted || isCurrent ? 'bg-green-500' : 'bg-gray-200'
              )} />
              {/* Circle */}
              <div className={cn(
                'w-6 h-6 rounded-full border-2 flex items-center justify-center z-10 flex-shrink-0',
                isRejected && isCurrent
                  ? 'border-red-500 bg-red-500 text-white'
                  : isCompleted
                  ? 'border-green-500 bg-green-500'
                  : isCurrent
                  ? 'border-blue-500 bg-blue-500'
                  : 'border-gray-300 bg-white'
              )}>
                {isCompleted && (
                  <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                )}
                {isRejected && isCurrent && (
                  <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                )}
              </div>
              {/* Right line */}
              <div className={cn(
                'flex-1 h-0.5',
                idx === steps.length - 1 ? 'bg-transparent' : isCompleted ? 'bg-green-500' : 'bg-gray-200'
              )} />
            </div>
            {/* Label */}
            <span className={cn(
              'mt-1 text-[10px] text-center leading-tight',
              isCompleted ? 'text-green-700' : isCurrent ? 'text-blue-700 font-semibold' : 'text-gray-400'
            )}>
              {step.label}
            </span>
          </div>
        );
      })}

      {/* Show rejection state if terminated */}
      {isRejected && currentStatus !== 'reddedildi' && (
        <div className="flex flex-col items-center min-w-[80px]">
          <div className="flex items-center w-full">
            <div className="flex-1 h-0.5 bg-red-400" />
            <div className="w-6 h-6 rounded-full border-2 border-red-500 bg-red-500 flex items-center justify-center flex-shrink-0">
              <svg className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={3}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </div>
            <div className="flex-1 h-0.5 bg-transparent" />
          </div>
          <span className="mt-1 text-[10px] text-red-600 font-semibold text-center leading-tight">
            {currentStatus.replace('_', ' ')}
          </span>
        </div>
      )}
    </div>
  );
}
