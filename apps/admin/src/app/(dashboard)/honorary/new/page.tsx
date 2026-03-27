'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useAuthStore } from '@/lib/store/auth.store';

const proposalSchema = z.object({
  nominee_name: z.string().min(2, 'Ad Soyad en az 2 karakter olmalı'),
  nominee_linkedin: z.string().url('Geçerli bir LinkedIn URL giriniz'),
  proposal_reason: z.string().min(100, 'Gerekçe en az 100 karakter olmalı'),
});

type ProposalFormData = z.infer<typeof proposalSchema>;

export default function NewHonoraryProposalPage() {
  const router = useRouter();
  const { isAuthenticated, user } = useAuthStore();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const canPropose = user?.role === 'asil_uye' || user?.role === 'yik_uye';

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<ProposalFormData>({
    resolver: zodResolver(proposalSchema),
  });

  const proposalReason = watch('proposal_reason', '');
  const remainingChars = 100 - (proposalReason?.length || 0);

  useEffect(() => {
    if (!isAuthenticated) {
      router.replace('/login');
      return;
    }

    if (!canPropose) {
      router.replace('/applications');
    }
  }, [isAuthenticated, canPropose, router]);

  const onSubmit = async (data: ProposalFormData) => {
    setIsLoading(true);
    setError(null);

    try {
      const token = useAuthStore.getState().accessToken;
      const response = await fetch('http://localhost:8080/api/v1/honorary/propose', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error?.message || 'Öneri oluşturulamadı');
      }

      // Redirect to the newly created application
      const applicationId = result.data?.application_id;
      if (applicationId) {
        router.push(`/applications/${applicationId}`);
      } else {
        router.push('/honorary');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Bir hata oluştu');
    } finally {
      setIsLoading(false);
    }
  };

  if (!canPropose) return null;

  return (
    <div className="p-8">
      <div className="max-w-2xl mx-auto">
        <div className="mb-6">
          <button
            onClick={() => router.back()}
            className="text-sm text-gray-600 hover:text-gray-900 flex items-center"
          >
            <svg className="w-4 h-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            Geri Dön
          </button>
        </div>

        <div className="bg-white shadow-md rounded-lg p-8">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Onursal Üye Önerisi</h1>
          <p className="text-sm text-gray-600 mb-6">
            Derneğe önemli katkılarda bulunabilecek bir kişiyi onursal üyeliğe önerin
          </p>

          {error && (
            <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg text-red-800">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
            <div>
              <label htmlFor="nominee_name" className="block text-sm font-medium text-gray-700 mb-1">
                Aday Ad Soyad <span className="text-red-500">*</span>
              </label>
              <input
                {...register('nominee_name')}
                id="nominee_name"
                type="text"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Örn: Ahmet Yılmaz"
              />
              {errors.nominee_name && (
                <p className="mt-1 text-sm text-red-600">{errors.nominee_name.message}</p>
              )}
            </div>

            <div>
              <label htmlFor="nominee_linkedin" className="block text-sm font-medium text-gray-700 mb-1">
                LinkedIn Profil URL <span className="text-red-500">*</span>
              </label>
              <input
                {...register('nominee_linkedin')}
                id="nominee_linkedin"
                type="url"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="https://www.linkedin.com/in/username"
              />
              {errors.nominee_linkedin && (
                <p className="mt-1 text-sm text-red-600">{errors.nominee_linkedin.message}</p>
              )}
            </div>

            <div>
              <label htmlFor="proposal_reason" className="block text-sm font-medium text-gray-700 mb-1">
                Öneri Gerekçesi <span className="text-red-500">*</span>
              </label>
              <textarea
                {...register('proposal_reason')}
                id="proposal_reason"
                rows={6}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
                placeholder="Bu kişiyi neden onursal üyeliğe öneriyorsunuz? Lütfen detaylı açıklama yapınız (minimum 100 karakter)"
              />
              <div className="flex justify-between items-center mt-1">
                <div>
                  {errors.proposal_reason && (
                    <p className="text-sm text-red-600">{errors.proposal_reason.message}</p>
                  )}
                </div>
                <p className={`text-xs ${remainingChars > 0 ? 'text-red-500' : 'text-green-600'}`}>
                  {remainingChars > 0 ? `${remainingChars} karakter daha gerekli` : '✓ Yeterli'}
                </p>
              </div>
            </div>

            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <div className="flex">
                <svg className="h-5 w-5 text-blue-400 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-blue-800">Bilgilendirme</h3>
                  <div className="mt-2 text-sm text-blue-700">
                    <ul className="list-disc list-inside space-y-1">
                      <li>Öneri oluşturulduktan sonra tüm Yönetim Kurulu üyeleri bilgilendirilecektir</li>
                      <li>Öneri, normal başvuru sürecinden geçecektir (YK Ön → YİK → YK Final)</li>
                      <li>Aday LinkedIn profilinin benzersiz olması gerekmektedir</li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>

            <div className="flex justify-end space-x-3 pt-4">
              <button
                type="button"
                onClick={() => router.back()}
                className="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 transition-colors"
                disabled={isLoading}
              >
                İptal
              </button>
              <button
                type="submit"
                disabled={isLoading}
                className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isLoading ? (
                  <span className="flex items-center">
                    <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Oluşturuluyor...
                  </span>
                ) : (
                  'Öneriyi Oluştur'
                )}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}