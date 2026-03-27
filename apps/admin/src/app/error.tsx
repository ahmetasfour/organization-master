'use client';

import { useEffect } from 'react';
import { AlertCircle } from 'lucide-react';

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // Log to error reporting service
    console.error('Application error:', error);
  }, [error]);

  return (
    <div className="flex min-h-screen flex-col items-center justify-center p-6 text-center bg-gray-50">
      <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-red-100">
        <AlertCircle className="h-8 w-8 text-red-600" />
      </div>
      <h2 className="text-xl font-semibold text-gray-900">Bir Hata Oluştu</h2>
      <p className="mt-2 max-w-md text-sm text-gray-500">
        Beklenmeyen bir hata oluştu. Sorunu çözemiyorsak lütfen sistem yöneticisiyle iletişime geçin.
      </p>
      {error.digest && (
        <p className="mt-2 text-xs text-gray-400">
          Hata Kodu: {error.digest}
        </p>
      )}
      <button
        onClick={reset}
        className="mt-6 rounded-lg bg-gray-900 px-5 py-2.5 text-sm font-medium text-white transition hover:bg-gray-800"
      >
        Tekrar Dene
      </button>
    </div>
  );
}
