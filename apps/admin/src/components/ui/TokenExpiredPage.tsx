'use client';

import { Clock } from 'lucide-react';

/**
 * TokenExpiredPage is shown when a user accesses a response form with an expired token.
 * Returns HTTP 410 Gone equivalent in the UI.
 */
export function TokenExpiredPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
      <div className="text-center max-w-md">
        <div className="mb-6 flex justify-center">
          <div className="rounded-full bg-orange-100 p-4">
            <Clock className="h-10 w-10 text-orange-600" />
          </div>
        </div>
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Link Süresi Dolmuş</h1>
        <p className="text-gray-600 mb-6">
          Bu bağlantının geçerlilik süresi dolmuştur. Lütfen ilgili kişiyle iletişime geçerek yeni bir bağlantı talep edin.
        </p>
        <p className="text-sm text-gray-500">
          Hata Kodu: 410
        </p>
      </div>
    </div>
  );
}

export default TokenExpiredPage;
