"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import {
  useReactTable,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  flexRender,
  type ColumnDef,
  type SortingState,
} from "@tanstack/react-table";
import { Plus, Search, UserCheck, UserX } from "lucide-react";
import { useUsers, type User } from "@/lib/hooks/useUsers";
import { useAuthStore } from "@/lib/store/auth.store";

const ROLE_LABELS: Record<string, string> = {
  admin: "Sistem Yöneticisi",
  yk: "YK Üyesi",
  yik: "YİK Üyesi",
  koordinator: "Koordinatör",
  asil_uye: "Asil Üye",
};

export default function UsersPage() {
  const router = useRouter();
  const { user } = useAuthStore();
  const [search, setSearch] = useState("");
  const [roleFilter, setRoleFilter] = useState<string>("");
  const [activeFilter, setActiveFilter] = useState<string>("all");
  const [sorting, setSorting] = useState<SortingState>([]);

  // Redirect if not admin
  if (user?.role !== "admin") {
    router.push("/");
    return null;
  }

  const isActiveFilter =
    activeFilter === "all"
      ? undefined
      : activeFilter === "active"
      ? true
      : false;

  const { data, isLoading } = useUsers({
    search,
    role: roleFilter || undefined,
    is_active: isActiveFilter,
  });

  const columns: ColumnDef<User>[] = [
    {
      accessorKey: "full_name",
      header: "Ad Soyad",
      cell: ({ row }) => (
        <div>
          <div className="font-medium text-slate-900">
            {row.original.full_name}
          </div>
          <div className="text-sm text-slate-500">{row.original.email}</div>
        </div>
      ),
    },
    {
      accessorKey: "role",
      header: "Rol",
      cell: ({ row }) => (
        <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800">
          {ROLE_LABELS[row.original.role] || row.original.role}
        </span>
      ),
    },
    {
      accessorKey: "is_active",
      header: "Durum",
      cell: ({ row }) =>
        row.original.is_active ? (
          <div className="flex items-center text-green-600">
            <UserCheck className="w-4 h-4 mr-1" />
            <span className="text-sm">Aktif</span>
          </div>
        ) : (
          <div className="flex items-center text-red-600">
            <UserX className="w-4 h-4 mr-1" />
            <span className="text-sm">Pasif</span>
          </div>
        ),
    },
    {
      accessorKey: "created_at",
      header: "Oluşturulma Tarihi",
      cell: ({ row }) =>
        new Date(row.original.created_at).toLocaleDateString("tr-TR", {
          year: "numeric",
          month: "short",
          day: "numeric",
        }),
    },
    {
      id: "actions",
      header: "İşlemler",
      cell: ({ row }) => (
        <button
          onClick={() => router.push(`/users/${row.original.id}/edit`)}
          className="text-sm text-blue-600 hover:text-blue-800 font-medium"
        >
          Düzenle
        </button>
      ),
    },
  ];

  const table = useReactTable({
    data: data?.data || [],
    columns,
    state: {
      sorting,
    },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">
            Kullanıcı Yönetimi
          </h1>
          <p className="text-sm text-slate-600 mt-1">
            Sistem kullanıcılarını yönetin
          </p>
        </div>
        <button
          onClick={() => router.push("/users/new")}
          className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium"
        >
          <Plus className="w-5 h-5 mr-2" />
          Yeni Kullanıcı
        </button>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg border border-slate-200 p-4">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {/* Search */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-400 w-5 h-5" />
            <input
              type="text"
              placeholder="Ad veya e-posta ile ara..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>

          {/* Role Filter */}
          <select
            value={roleFilter}
            onChange={(e) => setRoleFilter(e.target.value)}
            className="px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="">Tüm Roller</option>
            {Object.entries(ROLE_LABELS).map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </select>

          {/* Active Filter */}
          <select
            value={activeFilter}
            onChange={(e) => setActiveFilter(e.target.value)}
            className="px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="all">Tüm Durumlar</option>
            <option value="active">Sadece Aktif</option>
            <option value="inactive">Sadece Pasif</option>
          </select>
        </div>
      </div>

      {/* Table */}
      <div className="bg-white rounded-lg border border-slate-200 overflow-hidden">
        {isLoading ? (
          <div className="p-12 text-center text-slate-500">Yükleniyor...</div>
        ) : !data?.data || data.data.length === 0 ? (
          <div className="p-12 text-center text-slate-500">
            Kullanıcı bulunamadı
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-50 border-b border-slate-200">
                {table.getHeaderGroups().map((headerGroup) => (
                  <tr key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <th
                        key={header.id}
                        className="px-6 py-3 text-left text-xs font-semibold text-slate-700 uppercase tracking-wider"
                      >
                        {header.isPlaceholder
                          ? null
                          : flexRender(
                              header.column.columnDef.header,
                              header.getContext()
                            )}
                      </th>
                    ))}
                  </tr>
                ))}
              </thead>
              <tbody className="divide-y divide-slate-200">
                {table.getRowModel().rows.map((row) => (
                  <tr key={row.id} className="hover:bg-slate-50">
                    {row.getVisibleCells().map((cell) => (
                      <td key={cell.id} className="px-6 py-4 whitespace-nowrap">
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext()
                        )}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination Info */}
        {data?.total && (
          <div className="px-6 py-3 border-t border-slate-200 bg-slate-50 text-sm text-slate-600">
            Toplam {data.total} kullanıcı bulundu
          </div>
        )}
      </div>
    </div>
  );
}
