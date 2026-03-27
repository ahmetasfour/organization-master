'use client';

import { useState } from 'react';
import { Globe, Lock, Loader2, CheckCircle2 } from 'lucide-react';
import { useConsentStatus, useRecordConsent } from '@/lib/hooks/useWebPublish';
import { showToast } from '@/components/ui/Toaster';

interface WebPublishPanelProps {
  applicationId: string;
  applicantName: string;
  membershipType: string;
}

const membershipLabels: Record<string, string> = {
  asil: 'Asil Üye',
  akademik: 'Akademik Üye',
  profesyonel: 'Profesyonel Üye',
  öğrenci: 'Öğrenci Üye',
  onursal: 'Onursal Üye',
};

export function WebPublishPanel({
  applicationId,
  applicantName,
  membershipType,
}: WebPublishPanelProps) {
  const [selectedOption, setSelectedOption] = useState<boolean | null>(null);
  const { data: consentStatus, isLoading } = useConsentStatus(applicationId);
  const recordMutation = useRecordConsent(applicationId);

  const handleSubmit = async () => {
    if (selectedOption === null) return;

    try {
      await recordMutation.mutateAsync({ consented: selectedOption });
      showToast(
        selectedOption
          ? 'Üye web sitesinde yayınlanacak.'
          : 'Üye yalnızca iç listede kalacak.',
        'success'
      );
    } catch {
      showToast('İşlem başarısız oldu. Lütfen tekrar deneyin.', 'error');
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 py-8 text-sm text-gray-400">
        <Loader2 className="h-4 w-4 animate-spin" />
        Yükleniyor...
      </div>
    );
  }

  // If consent already recorded, show read-only state
  if (consentStatus?.recorded) {
    return (
      <div className="space-y-4">
        <div className="flex items-center gap-3">
          <CheckCircle2 className="h-5 w-5 text-green-600" />
          <h3 className="text-base font-semibold text-gray-900">
            Web Yayın Onayı Kaydedildi
          </h3>
        </div>

        <div className="rounded-lg border border-gray-200 bg-gray-50 p-4 space-y-3">
          <div className="flex justify-between">
            <span className="text-sm text-gray-500">Üye Adı</span>
            <span className="text-sm font-medium text-gray-900">{applicantName}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-sm text-gray-500">Üyelik Tipi</span>
            <span className="text-sm font-medium text-gray-900">
              {membershipLabels[membershipType] || membershipType}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-sm text-gray-500">Karar</span>
            <span
              className={`inline-flex items-center gap-1.5 text-sm font-medium ${
                consentStatus.consented
                  ? 'text-green-700'
                  : 'text-yellow-700'
              }`}
            >
              {consentStatus.consented ? (
                <>
                  <Globe className="h-4 w-4" />
                  Web&apos;de Yayınlanıyor
                </>
              ) : (
                <>
                  <Lock className="h-4 w-4" />
                  İç Listede
                </>
              )}
            </span>
          </div>
          {consentStatus.recorded_at && (
            <div className="flex justify-between">
              <span className="text-sm text-gray-500">Kayıt Tarihi</span>
              <span className="text-sm text-gray-700">
                {new Date(consentStatus.recorded_at).toLocaleString('tr-TR')}
              </span>
            </div>
          )}
        </div>
      </div>
    );
  }

  // Show consent form
  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-base font-semibold text-gray-900">
          Web Sitesinde Yayın Onayı
        </h3>
        <p className="mt-1 text-sm text-gray-500">
          Kabul edilen üyenin web sitesinde yayınlanıp yayınlanmayacağını belirleyin.
        </p>
      </div>

      {/* Member info */}
      <div className="rounded-lg border border-gray-200 bg-gray-50 p-4">
        <div className="flex items-center gap-4">
          <div className="h-12 w-12 rounded-full bg-blue-100 flex items-center justify-center text-blue-700 font-semibold text-lg">
            {applicantName.charAt(0).toUpperCase()}
          </div>
          <div>
            <p className="font-medium text-gray-900">{applicantName}</p>
            <p className="text-sm text-gray-500">
              {membershipLabels[membershipType] || membershipType}
            </p>
          </div>
        </div>
      </div>

      {/* Options */}
      <div className="space-y-3">
        <label
          className={`flex items-start gap-4 rounded-lg border-2 p-4 cursor-pointer transition-colors ${
            selectedOption === true
              ? 'border-green-500 bg-green-50'
              : 'border-gray-200 hover:border-gray-300'
          }`}
        >
          <input
            type="radio"
            name="consent"
            checked={selectedOption === true}
            onChange={() => setSelectedOption(true)}
            className="mt-1"
          />
          <div>
            <div className="flex items-center gap-2">
              <Globe className="h-5 w-5 text-green-600" />
              <span className="font-medium text-gray-900">
                Evet, alfabetik listede yayınlansın
              </span>
            </div>
            <p className="mt-1 text-sm text-gray-500">
              Üye web sitesindeki kamuya açık üye listesinde görünecektir.
            </p>
          </div>
        </label>

        <label
          className={`flex items-start gap-4 rounded-lg border-2 p-4 cursor-pointer transition-colors ${
            selectedOption === false
              ? 'border-yellow-500 bg-yellow-50'
              : 'border-gray-200 hover:border-gray-300'
          }`}
        >
          <input
            type="radio"
            name="consent"
            checked={selectedOption === false}
            onChange={() => setSelectedOption(false)}
            className="mt-1"
          />
          <div>
            <div className="flex items-center gap-2">
              <Lock className="h-5 w-5 text-yellow-600" />
              <span className="font-medium text-gray-900">
                Hayır, yalnızca iç listede kalsın
              </span>
            </div>
            <p className="mt-1 text-sm text-gray-500">
              Üye yalnızca dahili kayıtlarda görünecek, web sitesinde yayınlanmayacaktır.
            </p>
          </div>
        </label>
      </div>

      {/* Submit button */}
      <button
        onClick={handleSubmit}
        disabled={selectedOption === null || recordMutation.isPending}
        className="w-full inline-flex items-center justify-center gap-2 rounded-lg bg-gray-900 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {recordMutation.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
        Onayla
      </button>
    </div>
  );
}
