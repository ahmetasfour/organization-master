'use client';

import { useState } from 'react';
import { useForm, useFieldArray } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useReputationStatus, useAddReputationContacts } from '@/lib/hooks/useReputation';
import { ContactQueryList } from './ContactQueryList';

// ─── Types ────────────────────────────────────────────────────────────────────

interface ReputationPanelProps {
  applicationId: string;
  membershipType: string;
}

// ─── Add contacts form schema ─────────────────────────────────────────────────

const contactSchema = z.object({
  name: z.string().min(2, 'İsim en az 2 karakter olmalıdır'),
  email: z.string().email('Geçerli bir e-posta adresi giriniz'),
});

const addContactsSchema = z.object({
  contacts: z
    .array(contactSchema)
    .length(10, 'Tam olarak 10 kişi eklenmesi zorunludur'),
});

type AddContactsForm = z.infer<typeof addContactsSchema>;

const EMPTY_CONTACT = { name: '', email: '' };
const DEFAULT_CONTACTS = Array(10).fill(EMPTY_CONTACT).map(() => ({ ...EMPTY_CONTACT }));

// ─── Component ────────────────────────────────────────────────────────────────

export function ReputationPanel({ applicationId, membershipType }: ReputationPanelProps) {
  const [showAddForm, setShowAddForm] = useState(false);
  const [addError, setAddError] = useState<string | null>(null);

  const reputationTypes = ['asil', 'akademik'];
  if (!reputationTypes.includes(membershipType)) {
    return null; // Hidden for non-asil/akademik types
  }

  const { data: status, isLoading, isError } = useReputationStatus(applicationId);
  const addMutation = useAddReputationContacts(applicationId);

  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<AddContactsForm>({
    resolver: zodResolver(addContactsSchema),
    defaultValues: { contacts: DEFAULT_CONTACTS },
  });

  const { fields } = useFieldArray({ control, name: 'contacts' });

  const onSubmit = async (values: AddContactsForm) => {
    setAddError(null);
    try {
      await addMutation.mutateAsync({ contacts: values.contacts });
      setShowAddForm(false);
      reset();
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: { message?: string } } } })?.response?.data
          ?.error?.message ?? 'Bir hata oluştu. Lütfen tekrar deneyin.';
      setAddError(message);
    }
  };

  // ─── Loading / Error states ───────────────────────────────────────────────
  if (isLoading) {
    return (
      <div className="py-6 text-center text-sm text-gray-400 animate-pulse">
        İtibar tarama durumu yükleniyor…
      </div>
    );
  }

  if (isError) {
    return (
      <div className="py-4 text-center text-sm text-red-600">
        İtibar tarama bilgisi yüklenirken hata oluştu.
      </div>
    );
  }

  // ─── Progress bar helper ──────────────────────────────────────────────────
  const total = status?.total_contacts ?? 0;
  const responded = status?.responded ?? 0;
  const clean = status?.clean ?? 0;
  const flagged = status?.flagged ?? 0;
  const progressPercent = total > 0 ? Math.round((responded / total) * 100) : 0;

  return (
    <div className="space-y-6">
      {/* ── Summary card ──────────────────────────────────────────────── */}
      {total > 0 && status && (
        <div className="bg-white border border-gray-200 rounded-xl p-6 space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-base font-semibold text-gray-900">İtibar Tarama İlerlemesi</h3>
            <span className="text-sm text-gray-500">
              {responded}/{total} yanıt alındı
            </span>
          </div>

          {/* Progress bar */}
          <div className="w-full bg-gray-100 rounded-full h-2.5 overflow-hidden">
            <div
              className="bg-blue-600 h-2.5 rounded-full transition-all duration-500"
              style={{ width: `${progressPercent}%` }}
            />
          </div>

          {/* Stats row */}
          <div className="grid grid-cols-3 gap-3 text-center text-sm">
            <div className="bg-gray-50 rounded-lg py-3">
              <p className="text-2xl font-bold text-gray-700">{total - responded}</p>
              <p className="text-xs text-gray-500 mt-0.5">Bekliyor</p>
            </div>
            <div className="bg-green-50 rounded-lg py-3">
              <p className="text-2xl font-bold text-green-700">{clean}</p>
              <p className="text-xs text-green-600 mt-0.5">Temiz</p>
            </div>
            <div className="bg-red-50 rounded-lg py-3">
              <p className="text-2xl font-bold text-red-700">{flagged}</p>
              <p className="text-xs text-red-600 mt-0.5">Olumsuz</p>
            </div>
          </div>

          {flagged > 0 && (
            <div className="flex items-start gap-2 bg-red-50 border border-red-200 rounded-lg p-3 text-sm text-red-700">
              <span className="text-base leading-none">⚠️</span>
              <span>
                <strong>{flagged}</strong> olumsuz yanıt alındı. YK incelemesi gerekebilir.
              </span>
            </div>
          )}
        </div>
      )}

      {/* ── Contact list ──────────────────────────────────────────────── */}
      {total > 0 && status && (
        <div className="bg-white border border-gray-200 rounded-xl overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-100">
            <h3 className="text-sm font-semibold text-gray-900">İletişim Kişileri</h3>
            <p className="text-xs text-gray-500 mt-0.5">
              E-posta adresleri gizlilik nedeniyle maskelenmiştir.
            </p>
          </div>
          <ContactQueryList contacts={status.contacts} />
        </div>
      )}

      {/* ── Add contacts form ─────────────────────────────────────────── */}
      {total === 0 && !showAddForm && (
        <div className="text-center py-8 border-2 border-dashed border-gray-200 rounded-xl">
          <p className="text-sm text-gray-500 mb-3">
            Bu başvuru için henüz itibar tarama kişisi eklenmemiştir.
          </p>
          <button
            onClick={() => setShowAddForm(true)}
            className="inline-flex items-center gap-2 px-4 py-2 bg-slate-800 text-white text-sm font-medium rounded-lg hover:bg-slate-700 transition-colors"
          >
            <span>+</span> 10 Kişi Ekle
          </button>
        </div>
      )}

      {showAddForm && (
        <div className="bg-white border border-gray-200 rounded-xl p-6">
          <div className="flex items-center justify-between mb-5">
            <div>
              <h3 className="text-base font-semibold text-gray-900">
                İtibar Tarama Kişileri Ekle
              </h3>
              <p className="text-xs text-gray-500 mt-0.5">
                Tam olarak 10 kişi girilmesi zorunludur.
              </p>
            </div>
            <button
              onClick={() => { setShowAddForm(false); setAddError(null); }}
              className="text-gray-400 hover:text-gray-600 text-xl leading-none"
            >
              ×
            </button>
          </div>

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-3">
            {fields.map((field, idx) => (
              <div key={field.id} className="grid grid-cols-2 gap-3 items-start">
                <div>
                  <label className="block text-xs text-gray-500 mb-1">
                    #{idx + 1} İsim Soyisim
                  </label>
                  <input
                    {...register(`contacts.${idx}.name`)}
                    type="text"
                    placeholder="Ad Soyad"
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"
                  />
                  {errors.contacts?.[idx]?.name && (
                    <p className="text-xs text-red-600 mt-0.5">
                      {errors.contacts[idx]?.name?.message}
                    </p>
                  )}
                </div>
                <div>
                  <label className="block text-xs text-gray-500 mb-1">E-posta</label>
                  <input
                    {...register(`contacts.${idx}.email`)}
                    type="email"
                    placeholder="ornek@domain.com"
                    className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-500"
                  />
                  {errors.contacts?.[idx]?.email && (
                    <p className="text-xs text-red-600 mt-0.5">
                      {errors.contacts[idx]?.email?.message}
                    </p>
                  )}
                </div>
              </div>
            ))}

            {errors.contacts?.message && (
              <p className="text-sm text-red-600">{errors.contacts.message}</p>
            )}

            {addError && (
              <div className="rounded-md bg-red-50 border border-red-200 px-4 py-3">
                <p className="text-sm text-red-700">{addError}</p>
              </div>
            )}

            <div className="flex gap-3 pt-2">
              <button
                type="submit"
                disabled={addMutation.isPending}
                className="flex-1 py-2.5 bg-slate-800 text-white text-sm font-medium rounded-lg hover:bg-slate-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {addMutation.isPending ? 'Gönderiliyor…' : 'Talepleri Gönder'}
              </button>
              <button
                type="button"
                onClick={() => { setShowAddForm(false); setAddError(null); reset(); }}
                className="px-4 py-2.5 border border-gray-300 text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors"
              >
                İptal
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}
