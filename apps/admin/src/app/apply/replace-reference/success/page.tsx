"use client";

import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { Suspense } from "react";

function SuccessContent() {
  const searchParams = useSearchParams();
  const applicationId = searchParams.get("id");

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 p-4">
      <div className="w-full max-w-2xl rounded-lg bg-white p-8 shadow-md">
        <div className="mb-6 flex h-16 w-16 items-center justify-center rounded-full bg-green-100">
          <svg
            className="h-8 w-8 text-green-600"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M5 13l4 4L19 7"
            />
          </svg>
        </div>

        <h1 className="mb-4 text-2xl font-bold text-gray-900">
          Referans Bilgisi Kaydedildi
        </h1>

        <div className="mb-6 space-y-4 text-gray-700">
          <p>
            Yeni referans bilginiz başarıyla kaydedilmiştir. Referansınıza
            e-posta ile bilgilendirme gönderilmiştir.
          </p>

          {applicationId && (
            <div className="rounded-md border border-blue-100 bg-blue-50 p-4">
              <p className="text-sm text-blue-800">
                <strong>Başvuru No:</strong> {applicationId}
              </p>
            </div>
          )}

          <div className="rounded-md border border-gray-200 bg-gray-50 p-4">
            <h3 className="mb-2 font-semibold text-gray-900">
              Sonraki Adımlar:
            </h3>
            <ul className="list-inside list-disc space-y-1 text-sm text-gray-700">
              <li>
                Referansınız e-posta üzerinden gelen bağlantıya tıklayarak
                görüşünü bildirecektir
              </li>
              <li>
                Referans süreci tamamlandığında size bilgilendirme
                yapılacaktır
              </li>
              <li>
                Başvurunuzun durumunu takip etmek için lütfen bekleyiniz
              </li>
            </ul>
          </div>
        </div>

        <div className="flex flex-col gap-3 sm:flex-row">
          <Link
            href="/"
            className="flex-1 rounded-md bg-blue-600 px-4 py-3 text-center font-semibold text-white transition-colors hover:bg-blue-700"
          >
            Ana Sayfaya Dön
          </Link>
        </div>
      </div>
    </div>
  );
}

export default function ReplaceReferenceSuccessPage() {
  return (
    <Suspense fallback={<div className="flex min-h-screen items-center justify-center">Yükleniyor...</div>}>
      <SuccessContent />
    </Suspense>
  );
}
