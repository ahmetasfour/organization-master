"use client";

import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { CheckCircle, Mail, ArrowRight } from "lucide-react";

const MEMBERSHIP_TYPES = {
  asil: "Asil Üyelik",
  akademik: "Akademik Üyelik",
  profesyonel: "Profesyonel Üyelik",
  ogrenci: "Öğrenci Üyelik",
} as const;

export default function ApplicationSuccessPage() {
  const searchParams = useSearchParams();
  const applicationId = searchParams.get("id");
  const membershipType = searchParams.get("type") as keyof typeof MEMBERSHIP_TYPES;

  const requiresReferences =
    membershipType === "asil" || membershipType === "akademik";

  return (
    <div className="min-h-screen bg-gradient-to-br from-green-50 to-emerald-50 flex items-center justify-center py-12 px-4">
      <div className="max-w-2xl w-full">
        {/* Success Icon */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-20 h-20 bg-green-500 rounded-full mb-4">
            <CheckCircle className="w-12 h-12 text-white" />
          </div>
          <h1 className="text-3xl font-bold text-slate-900 mb-2">
            Başvurunuz Alındı!
          </h1>
          <p className="text-lg text-slate-600">
            {MEMBERSHIP_TYPES[membershipType]} başvurunuz başarıyla kaydedildi.
          </p>
        </div>

        {/* Info Card */}
        <div className="bg-white rounded-lg shadow-lg p-8 mb-6">
          <div className="space-y-6">
            {/* Application ID */}
            <div className="border-b border-slate-200 pb-4">
              <p className="text-sm text-slate-600 mb-1">Başvuru Numaranız</p>
              <p className="text-2xl font-mono font-semibold text-slate-900">
                {applicationId}
              </p>
              <p className="text-xs text-slate-500 mt-1">
                Bu numarayı kaydediniz.
              </p>
            </div>

            {/* Next Steps */}
            <div>
              <h2 className="text-lg font-semibold text-slate-900 mb-3 flex items-center">
                <Mail className="w-5 h-5 mr-2 text-blue-500" />
                Sonraki Adımlar
              </h2>
              <ul className="space-y-3">
                {requiresReferences ? (
                  <>
                    <li className="flex items-start">
                      <span className="flex-shrink-0 w-6 h-6 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center text-sm font-semibold mr-3">
                        1
                      </span>
                      <div>
                        <p className="text-slate-700">
                          Referanslarınıza e-posta gönderildi.
                        </p>
                        <p className="text-sm text-slate-500 mt-1">
                          Belirttiğiniz referanslar değerlendirme formunu
                          doldurmaları için bilgilendirildi.
                        </p>
                      </div>
                    </li>
                    <li className="flex items-start">
                      <span className="flex-shrink-0 w-6 h-6 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center text-sm font-semibold mr-3">
                        2
                      </span>
                      <div>
                        <p className="text-slate-700">
                          Referans süreci tamamlandığında bilgilendirileceksiniz.
                        </p>
                        <p className="text-sm text-slate-500 mt-1">
                          Tüm referanslar yanıt verdikten sonra başvurunuz
                          değerlendirme aşamasına geçecektir.
                        </p>
                      </div>
                    </li>
                  </>
                ) : (
                  <>
                    <li className="flex items-start">
                      <span className="flex-shrink-0 w-6 h-6 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center text-sm font-semibold mr-3">
                        1
                      </span>
                      <div>
                        <p className="text-slate-700">
                          Başvurunuz danışma sürecine alındı.
                        </p>
                        <p className="text-sm text-slate-500 mt-1">
                          Üyelerimiz tarafından değerlendirilecek ve size
                          bildirim gönderilecektir.
                        </p>
                      </div>
                    </li>
                  </>
                )}
                <li className="flex items-start">
                  <span className="flex-shrink-0 w-6 h-6 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center text-sm font-semibold mr-3">
                    {requiresReferences ? "3" : "2"}
                  </span>
                  <div>
                    <p className="text-slate-700">
                      Başvuru durumunuz hakkında e-posta ile bilgilendirileceksiniz.
                    </p>
                    <p className="text-sm text-slate-500 mt-1">
                      Sürecin her aşamasında güncel bilgiler size
                      iletilecektir.
                    </p>
                  </div>
                </li>
              </ul>
            </div>

            {/* Important Note */}
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <p className="text-sm text-blue-900">
                <strong>Not:</strong> Lütfen spam klasörünüzü kontrol etmeyi
                unutmayın. Başvurunuzla ilgili tüm bildirimler kayıtlı
                e-posta adresinize gönderilecektir.
              </p>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="text-center space-y-3">
          <Link
            href="/apply"
            className="inline-flex items-center justify-center px-6 py-3 bg-slate-900 text-white font-medium rounded-lg hover:bg-slate-800 transition-colors"
          >
            Yeni Başvuru Yap
            <ArrowRight className="w-4 h-4 ml-2" />
          </Link>
          <p className="text-sm text-slate-600">
            Sorularınız için:{" "}
            <a
              href="mailto:info@example.com"
              className="text-blue-600 hover:underline"
            >
              info@example.com
            </a>
          </p>
        </div>
      </div>
    </div>
  );
}
