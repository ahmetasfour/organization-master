'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

// ─── Types ────────────────────────────────────────────────────────────────────

interface ReputationFormData {
  contact_name: string;
  applicant_name: string;
  applicant_linkedin: string;
  expires_at: string;
}

// ─── Zod schema ───────────────────────────────────────────────────────────────

const responseSchema = z
  .object({
    response_type: z.enum(['clean', 'negative']),
    reason: z.string().optional(),
  })
  .superRefine((data, ctx) => {
    if (data.response_type === 'negative') {
      if (!data.reason || data.reason.trim().length < 30) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ['reason'],
          message: 'Olumsuz yanıt için en az 30 karakter gereklidir.',
        });
      }
    }
  });

type ResponseFormValues = z.infer<typeof responseSchema>;

// ─── Sub-pages ────────────────────────────────────────────────────────────────

function StatusCard({
  icon,
  title,
  message,
  variant,
}: {
  icon: string;
  title: string;
  message: string;
  variant: 'success' | 'warning' | 'info' | 'error';
}) {
  const variantStyles = {
    success: 'bg-green-50 border-green-200 text-green-800',
    warning: 'bg-yellow-50 border-yellow-200 text-yellow-800',
    info: 'bg-blue-50 border-blue-200 text-blue-800',
    error: 'bg-red-50 border-red-200 text-red-800',
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div
        className={`max-w-md w-full border rounded-xl p-8 text-center shadow-sm ${variantStyles[variant]}`}
      >
        <div className="text-5xl mb-4">{icon}</div>
        <h1 className="text-xl font-semibold mb-2">{title}</h1>
        <p className="text-sm leading-relaxed">{message}</p>
      </div>
    </div>
  );
}

function TokenExpiredPage() {
  return (
    <StatusCard
      icon="⏰"
      title="Bağlantı Süresi Dolmuş"
      message="Bu itibar tarama bağlantısının geçerlilik süresi dolmuştur. Lütfen kurum ile iletişime geçiniz."
      variant="warning"
    />
  );
}

function TokenUsedPage() {
  return (
    <StatusCard
      icon="✅"
      title="Yanıt Zaten Alındı"
      message="Bu bağlantı daha önce kullanılmıştır. Görüşünüz sistemimizde kayıtlıdır."
      variant="info"
    />
  );
}

function ThankYouPage({ message }: { message: string }) {
  return (
    <StatusCard icon="🙏" title="Teşekkürler" message={message} variant="success" />
  );
}

// ─── Main page ────────────────────────────────────────────────────────────────

type PageState = 'loading' | 'form' | 'expired' | 'used' | 'submitted' | 'error';

export default function ReputationRespondPage() {
  const params = useParams();
  const token = params?.token as string;

  const [pageState, setPageState] = useState<PageState>('loading');
  const [formData, setFormData] = useState<ReputationFormData | null>(null);
  const [submitMessage, setSubmitMessage] = useState('');
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<ResponseFormValues>({
    resolver: zodResolver(responseSchema),
  });

  const selectedType = watch('response_type');

  // ── Fetch form data on mount ────────────────────────────────────────────
  useEffect(() => {
    if (!token) return;

    axios
      .get(`${API_URL}/reputation/respond/${token}`)
      .then((res) => {
        setFormData(res.data.data);
        setPageState('form');
      })
      .catch((err) => {
        const status = err.response?.status;
        if (status === 410) setPageState('expired');
        else if (status === 409) setPageState('used');
        else setPageState('error');
      });
  }, [token]);

  // ── Submit handler ──────────────────────────────────────────────────────
  const onSubmit = async (values: ResponseFormValues) => {
    setIsSubmitting(true);
    setSubmitError(null);

    try {
      const res = await axios.post(`${API_URL}/reputation/respond/${token}`, {
        response_type: values.response_type,
        reason: values.reason ?? '',
      });
      setSubmitMessage(
        res.data.data?.message ?? 'Yanıtınız alındı. Katkınız için teşekkür ederiz.'
      );
      setPageState('submitted');
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        const status = err.response?.status;
        if (status === 410) {
          setPageState('expired');
        } else if (status === 409) {
          setPageState('used');
        } else {
          const message =
            err.response?.data?.error?.message ?? 'Bir hata oluştu. Lütfen tekrar deneyin.';
          setSubmitError(message);
        }
      } else {
        setSubmitError('Bir hata oluştu. Lütfen tekrar deneyin.');
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  // ──────────────────────────────────────────────────────────────────────────
  if (pageState === 'loading') {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-gray-400 text-sm animate-pulse">Yükleniyor…</div>
      </div>
    );
  }

  if (pageState === 'expired') return <TokenExpiredPage />;
  if (pageState === 'used') return <TokenUsedPage />;
  if (pageState === 'submitted') return <ThankYouPage message={submitMessage} />;
  if (pageState === 'error') {
    return (
      <StatusCard
        icon="⚠️"
        title="Hata"
        message="Geçersiz veya bulunamayan bağlantı. Lütfen kurum ile iletişime geçiniz."
        variant="error"
      />
    );
  }

  // ── Form ──────────────────────────────────────────────────────────────────
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4 py-12">
      <div className="max-w-lg w-full bg-white rounded-xl shadow-sm border border-gray-200 p-8">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-2xl font-bold text-gray-900">İtibar Tarama Formu</h1>
          <p className="mt-1 text-sm text-gray-500">Üyelik Yönetim Sistemi</p>
        </div>

        {/* Context */}
        {formData && (
          <div className="mb-6 bg-gray-50 rounded-lg p-4 text-sm space-y-2">
            <div>
              <span className="font-medium text-gray-700">Sayın </span>
              <span className="text-gray-900 font-semibold">{formData.contact_name}</span>
            </div>
            <div>
              <span className="font-medium text-gray-700">Başvuran: </span>
              <span className="text-gray-900">{formData.applicant_name}</span>
            </div>
            {formData.applicant_linkedin && (
              <div>
                <span className="font-medium text-gray-700">LinkedIn: </span>
                <a
                  href={formData.applicant_linkedin}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 hover:underline break-all"
                >
                  {formData.applicant_linkedin}
                </a>
              </div>
            )}
            <div>
              <span className="font-medium text-gray-700">Geçerlilik: </span>
              <span className="text-gray-500">{formData.expires_at}</span>
            </div>
          </div>
        )}

        {/* Question */}
        <p className="mb-5 text-sm text-gray-700 font-medium">
          Bu kişi hakkında olumsuz bir bilginiz var mı?
        </p>

        {/* Form */}
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
          <fieldset>
            <div className="space-y-3">
              {/* Clean */}
              <label className="flex items-start gap-3 cursor-pointer group">
                <input
                  {...register('response_type')}
                  type="radio"
                  value="clean"
                  className="mt-0.5 h-4 w-4 text-green-600 border-gray-300 focus:ring-green-500"
                />
                <div>
                  <span className="block text-sm font-medium text-gray-800 group-hover:text-green-700">
                    Hayır, olumsuz bir bilgim yok
                  </span>
                  <span className="block text-xs text-gray-500">
                    Bu kişi hakkında herhangi bir olumsuz bilgim bulunmamaktadır.
                  </span>
                </div>
              </label>

              {/* Negative */}
              <label className="flex items-start gap-3 cursor-pointer group">
                <input
                  {...register('response_type')}
                  type="radio"
                  value="negative"
                  className="mt-0.5 h-4 w-4 text-red-600 border-gray-300 focus:ring-red-500"
                />
                <div>
                  <span className="block text-sm font-medium text-gray-800 group-hover:text-red-700">
                    Evet, olumsuz bilgim var
                  </span>
                  <span className="block text-xs text-gray-500">
                    Bu kişi hakkında paylaşmak istediğim olumsuz bilgi veya deneyim mevcuttur.
                  </span>
                </div>
              </label>
            </div>

            {errors.response_type && (
              <p className="mt-2 text-xs text-red-600">{errors.response_type.message}</p>
            )}
          </fieldset>

          {/* Reason textarea — only for negative */}
          {selectedType === 'negative' && (
            <div>
              <label htmlFor="reason" className="block text-sm font-medium text-gray-700 mb-1">
                Lütfen olumsuz bilginizi açıklayınız{' '}
                <span className="text-red-500">*</span>
              </label>
              <textarea
                {...register('reason')}
                id="reason"
                rows={4}
                placeholder="En az 30 karakter gereklidir…"
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-red-500 focus:border-red-500 resize-none"
              />
              {errors.reason && (
                <p className="mt-1 text-xs text-red-600">{errors.reason.message}</p>
              )}
            </div>
          )}

          {/* Server error */}
          {submitError && (
            <div className="rounded-md bg-red-50 border border-red-200 px-4 py-3">
              <p className="text-sm text-red-700">{submitError}</p>
            </div>
          )}

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full flex justify-center py-2.5 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-slate-800 hover:bg-slate-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-slate-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {isSubmitting ? 'Gönderiliyor…' : 'Yanıtımı Gönder'}
          </button>
        </form>

        <p className="mt-6 text-xs text-center text-gray-400">
          Bu e-posta otomatik olarak gönderilmiştir. Üyelik Yönetim Sistemi
        </p>
      </div>
    </div>
  );
}
