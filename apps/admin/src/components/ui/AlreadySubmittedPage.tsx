'use client';

import { CheckCircle } from 'lucide-react';

/**
 * AlreadySubmittedPage is shown when a user tries to respond to a token that was already used.
 * Returns HTTP 409 Conflict equivalent in the UI.
 */
export function AlreadySubmittedPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
      <div className="text-center max-w-md">
        <div className="mb-6 flex justify-center">
          <div className="rounded-full bg-green-100 p-4">
            <CheckCircle className="h-10 w-10 text-green-600" />
          </div>
        </div>
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Yanıtınız Alındı</h1>
        <p className="text-gray-600 mb-6">
          Bu form daha önce yanıtlanmıştır. Aynı bağlantı ile tekrar yanıt veremezsiniz.
        </p>
        <p className="text-sm text-gray-500">
          Teşekkür ederiz.
        </p>
      </div>
    </div>
  );
}

export default AlreadySubmittedPage;
