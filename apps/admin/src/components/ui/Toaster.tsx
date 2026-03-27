'use client';

import { useEffect, useState } from 'react';
import { X, CheckCircle2, AlertTriangle, Info } from 'lucide-react';
import { cn } from '@/lib/utils';

// Simple toast store using a global state pattern
interface Toast {
  id: string;
  type: 'success' | 'error' | 'info';
  message: string;
}

let toastListeners: ((toasts: Toast[]) => void)[] = [];
let toasts: Toast[] = [];

function notifyListeners() {
  toastListeners.forEach((listener) => listener([...toasts]));
}

export function showToast(message: string, type: 'success' | 'error' | 'info' = 'success') {
  const id = Math.random().toString(36).substring(2, 9);
  toasts.push({ id, type, message });
  notifyListeners();

  // Auto dismiss after 4 seconds
  setTimeout(() => {
    dismissToast(id);
  }, 4000);
}

export function dismissToast(id: string) {
  toasts = toasts.filter((t) => t.id !== id);
  notifyListeners();
}

// Hook to subscribe to toasts
function useToasts() {
  const [localToasts, setLocalToasts] = useState<Toast[]>([]);

  useEffect(() => {
    const listener = (newToasts: Toast[]) => setLocalToasts(newToasts);
    toastListeners.push(listener);
    return () => {
      toastListeners = toastListeners.filter((l) => l !== listener);
    };
  }, []);

  return localToasts;
}

const iconMap = {
  success: CheckCircle2,
  error: AlertTriangle,
  info: Info,
};

const colorMap = {
  success: 'bg-green-50 border-green-200 text-green-800',
  error: 'bg-red-50 border-red-200 text-red-800',
  info: 'bg-blue-50 border-blue-200 text-blue-800',
};

const iconColorMap = {
  success: 'text-green-500',
  error: 'text-red-500',
  info: 'text-blue-500',
};

export function Toaster() {
  const activeToasts = useToasts();

  if (activeToasts.length === 0) return null;

  return (
    <div className="fixed bottom-4 right-4 z-[100] flex flex-col gap-2">
      {activeToasts.map((toast) => {
        const Icon = iconMap[toast.type];
        return (
          <div
            key={toast.id}
            className={cn(
              'flex items-center gap-3 rounded-lg border px-4 py-3 shadow-lg animate-in slide-in-from-bottom-5',
              colorMap[toast.type]
            )}
          >
            <Icon className={cn('h-5 w-5 flex-shrink-0', iconColorMap[toast.type])} />
            <p className="text-sm font-medium flex-1">{toast.message}</p>
            <button
              onClick={() => dismissToast(toast.id)}
              className="flex-shrink-0 rounded-full p-1 hover:bg-black/5 transition-colors"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        );
      })}
    </div>
  );
}
