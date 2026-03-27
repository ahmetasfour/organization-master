'use client';

import { AlertCircle } from 'lucide-react';

interface ErrorPageProps {
  title?: string;
  message?: string;
  retry?: () => void;
}

export function ErrorPage({
  title = 'Hata Oluştu',
  message = 'Beklenmeyen bir hata oluştu. Lütfen daha sonra tekrar deneyin.',
  retry,
}: ErrorPageProps) {
  return (
    <div className="flex min-h-[400px] flex-col items-center justify-center p-6 text-center">
      <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-red-100">
        <AlertCircle className="h-8 w-8 text-red-600" />
      </div>
      <h2 className="text-lg font-semibold text-gray-900">{title}</h2>
      <p className="mt-2 max-w-md text-sm text-gray-500">{message}</p>
      {retry && (
        <button
          onClick={retry}
          className="mt-4 rounded-lg bg-gray-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-gray-800"
        >
          Tekrar Dene
        </button>
      )}
    </div>
  );
}
