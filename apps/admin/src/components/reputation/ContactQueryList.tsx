'use client';

import type { ContactStatus } from '@/lib/api/reputation';

interface ContactQueryListProps {
  contacts: ContactStatus[];
}

const statusConfig: Record<
  ContactStatus['status'],
  { label: string; className: string }
> = {
  pending: {
    label: 'Bekliyor',
    className: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  },
  clean: {
    label: 'Temiz',
    className: 'bg-green-100 text-green-800 border-green-200',
  },
  flagged: {
    label: 'Olumsuz',
    className: 'bg-red-100 text-red-800 border-red-200',
  },
};

function StatusChip({ status }: { status: ContactStatus['status'] }) {
  const cfg = statusConfig[status];
  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${cfg.className}`}
    >
      {cfg.label}
    </span>
  );
}

export function ContactQueryList({ contacts }: ContactQueryListProps) {
  if (contacts.length === 0) {
    return (
      <p className="text-sm text-gray-500 italic">Henüz iletişim kişisi eklenmemiştir.</p>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200 text-sm">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              İsim
            </th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              E-posta
            </th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Durum
            </th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Yanıt Tarihi
            </th>
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-100">
          {contacts.map((contact) => (
            <tr key={contact.id} className="hover:bg-gray-50 transition-colors">
              <td className="px-4 py-3 font-medium text-gray-900">{contact.contact_name}</td>
              <td className="px-4 py-3 text-gray-500 font-mono text-xs">{contact.email}</td>
              <td className="px-4 py-3">
                <StatusChip status={contact.status} />
              </td>
              <td className="px-4 py-3 text-gray-500">
                {contact.responded_at
                  ? new Date(contact.responded_at).toLocaleDateString('tr-TR', {
                      day: '2-digit',
                      month: '2-digit',
                      year: 'numeric',
                      hour: '2-digit',
                      minute: '2-digit',
                    })
                  : '—'}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
