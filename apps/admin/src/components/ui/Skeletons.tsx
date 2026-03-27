'use client';

export function TableSkeleton({ rows = 5 }: { rows?: number }) {
  return (
    <div className="animate-pulse">
      {/* Header skeleton */}
      <div className="h-12 bg-gray-100 rounded-t-lg" />
      {/* Rows */}
      <div className="divide-y divide-gray-100 border-x border-b border-gray-200 rounded-b-lg bg-white">
        {Array.from({ length: rows }).map((_, i) => (
          <div key={i} className="flex items-center gap-4 px-4 py-4">
            <div className="h-4 w-32 rounded bg-gray-200" />
            <div className="h-4 w-48 rounded bg-gray-200" />
            <div className="h-4 w-20 rounded bg-gray-200" />
            <div className="h-4 w-24 rounded bg-gray-200" />
          </div>
        ))}
      </div>
    </div>
  );
}

export function CardSkeleton() {
  return (
    <div className="animate-pulse rounded-xl border border-gray-200 bg-white p-6 space-y-4">
      <div className="h-5 w-40 rounded bg-gray-200" />
      <div className="h-4 w-64 rounded bg-gray-200" />
      <div className="grid grid-cols-3 gap-4">
        <div className="h-16 rounded-lg bg-gray-100" />
        <div className="h-16 rounded-lg bg-gray-100" />
        <div className="h-16 rounded-lg bg-gray-100" />
      </div>
    </div>
  );
}

export function StatsSkeleton() {
  return (
    <div className="animate-pulse flex gap-4">
      {Array.from({ length: 4 }).map((_, i) => (
        <div
          key={i}
          className="flex-1 rounded-lg border border-gray-200 bg-white p-4"
        >
          <div className="h-6 w-12 rounded bg-gray-200 mb-2" />
          <div className="h-4 w-20 rounded bg-gray-100" />
        </div>
      ))}
    </div>
  );
}
