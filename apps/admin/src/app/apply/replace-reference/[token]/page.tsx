"use client";

import { useParams, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import api from "@/lib/api/client";

const replacementSchema = z.object({
  referee_name: z.string().min(2, "Referans adı en az 2 karakter olmalıdır"),
  referee_email: z.string().email("Geçerli bir e-posta adresi giriniz"),
});

type ReplacementFormData = z.infer<typeof replacementSchema>;

interface ReplacementInfo {
  applicant_name: string;
  membership_type: string;
  unknown_referee_name: string;
  application_id: string;
}

export default function ReplaceReferencePage() {
  const params = useParams();
  const router = useRouter();
  const token = params.token as string;
  const [loading, setLoading] = useState(true);
  const [replacementInfo, setReplacementInfo] = useState<ReplacementInfo | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ReplacementFormData>({
    resolver: zodResolver(replacementSchema),
  });

  useEffect(() => {
    async function loadReplacementData() {
      try {
        const response = await api.get(`/ref/replace/${token}`);
        if (response.data.success) {
          setReplacementInfo(response.data.data);
        } else {
          setError("Geçersiz veya süresi dolmuş bağlantı");
        }
      } catch (err: any) {
        if (err.response?.data?.error?.code === "TOKEN_EXPIRED") {
          setError("Bu bağlantının süresi dolmuş. Lütfen yeni bir e-posta talebi için sistem yöneticinizle iletişime geçiniz.");
        } else if (err.response?.data?.error?.code === "TOKEN_USED") {
          setError("Bu bağlantı zaten kullanılmış.");
        } else {
          setError("Bağlantı yüklenemedi. Lütfen tekrar deneyiniz.");
        }
      } finally {
        setLoading(false);
      }
    }

    loadReplacementData();
  }, [token]);

  const onSubmit = async (data: ReplacementFormData) => {
    setSubmitting(true);
    try {
      const response = await api.post(`/ref/replace/${token}`, data);
      if (response.data.success) {
        router.push(`/apply/replace-reference/success?id=${replacementInfo?.application_id}`);
      } else {
        alert("Bir hata oluştu. Lütfen tekrar deneyiniz.");
      }
    } catch (err: any) {
      if (err.response?.data?.error?.code === "TOKEN_EXPIRED") {
        setError("Bu bağlantının süresi dolmuş.");
      } else if (err.response?.data?.error?.code === "TOKEN_USED") {
        setError("Bu bağlantı zaten kullanılmış.");
      } else {
        alert(err.response?.data?.error?.message || "Bir hata oluştu. Lütfen tekrar deneyiniz.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-center">
          <div className="mb-4 text-lg">Yükleniyor...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50 p-4">
        <div className="w-full max-w-md rounded-lg border border-red-200 bg-white p-8 shadow-md">
          <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-red-100">
            <svg
              className="h-6 w-6 text-red-600"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </div>
          <h2 className="mb-2 text-xl font-semibold text-gray-900">Hata</h2>
          <p className="text-gray-600">{error}</p>
        </div>
      </div>
    );
  }

  if (!replacementInfo) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-center">
          <div className="mb-4 text-lg">Bilgi bulunamadı</div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 p-4">
      <div className="w-full max-w-2xl rounded-lg bg-white p-8 shadow-md">
        <div className="mb-6">
          <h1 className="mb-2 text-2xl font-bold text-gray-900">
            Yeni Referans Bilgisi
          </h1>
          <div className="rounded-md border border-amber-200 bg-amber-50 p-4">
            <p className="mb-2 text-sm text-amber-800">
              <strong>{replacementInfo.unknown_referee_name}</strong> adlı
              referansınız sizi tanımadığını bildirmiştir.
            </p>
            <p className="text-sm text-amber-800">
              Başvurunuzun ({replacementInfo.membership_type}) değerlendirmeye
              devam edebilmesi için lütfen yeni bir referans bilgisi giriniz.
            </p>
          </div>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          <div>
            <label
              htmlFor="referee_name"
              className="mb-2 block text-sm font-medium text-gray-700"
            >
              Referans Adı Soyadı <span className="text-red-500">*</span>
            </label>
            <input
              id="referee_name"
              type="text"
              {...register("referee_name")}
              className="w-full rounded-md border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              placeholder="Örn: Prof. Dr. Ahmet Yılmaz"
            />
            {errors.referee_name && (
              <p className="mt-1 text-sm text-red-600">
                {errors.referee_name.message}
              </p>
            )}
          </div>

          <div>
            <label
              htmlFor="referee_email"
              className="mb-2 block text-sm font-medium text-gray-700"
            >
              Referans E-posta Adresi <span className="text-red-500">*</span>
            </label>
            <input
              id="referee_email"
              type="email"
              {...register("referee_email")}
              className="w-full rounded-md border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              placeholder="ornek@universite.edu.tr"
            />
            {errors.referee_email && (
              <p className="mt-1 text-sm text-red-600">
                {errors.referee_email.message}
              </p>
            )}
          </div>

          <div className="rounded-md border border-blue-100 bg-blue-50 p-4">
            <p className="text-sm text-blue-800">
              <strong>Not:</strong> Girdiğiniz referans e-posta adresine onay
              talebi gönderilecektir. Lütfen referansınızın geçerli bir e-posta
              adresi olduğundan emin olunuz.
            </p>
          </div>

          <div className="flex gap-4">
            <button
              type="submit"
              disabled={submitting}
              className="flex-1 rounded-md bg-blue-600 px-4 py-3 font-semibold text-white transition-colors hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed"
            >
              {submitting ? "Gönderiliyor..." : "Referans Bilgisini Kaydet"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
