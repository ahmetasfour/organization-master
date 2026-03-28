"use client";

import { useState, useEffect } from "react";
import { useRouter, useParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { AlertCircle, ArrowLeft, Loader2 } from "lucide-react";
import { useUser, useUpdateUser } from "@/lib/hooks/useUsers";
import { useAuthStore } from "@/lib/store/auth.store";

const ROLE_OPTIONS = [
  { value: "admin", label: "Sistem Yöneticisi" },
  { value: "yk", label: "YK Üyesi" },
  { value: "yik", label: "YİK Üyesi" },
  { value: "koordinator", label: "Koordinatör" },
  { value: "asil_uye", label: "Asil Üye" },
];

const updateUserSchema = z.object({
  full_name: z
    .string()
    .min(2, "Ad soyad en az 2 karakter olmalıdır")
    .max(255, "Ad soyad en fazla 255 karakter olmalıdır"),
  role: z.enum(["admin", "yk", "yik", "koordinator", "asil_uye"], {
    message: "Rol seçimi zorunludur",
  }),
  is_active: z.boolean(),
});

type UpdateUserForm = z.infer<typeof updateUserSchema>;

export default function EditUserPage() {
  const router = useRouter();
  const params = useParams();
  const userId = params?.id as string;
  const { user: currentUser } = useAuthStore();
  const { data: user, isLoading } = useUser(userId);
  const updateMutation = useUpdateUser(userId);
  const [error, setError] = useState<string | null>(null);

  // Redirect if not admin
  if (currentUser?.role !== "admin") {
    router.push("/");
    return null;
  }

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<UpdateUserForm>({
    resolver: zodResolver(updateUserSchema),
  });

  // Populate form when user data loads
  useEffect(() => {
    if (user) {
      reset({
        full_name: user.full_name,
        role: user.role as any,
        is_active: user.is_active,
      });
    }
  }, [user, reset]);

  const onSubmit = async (data: UpdateUserForm) => {
    try {
      setError(null);

      // Prevent editing own role
      if (userId === currentUser?.id && data.role !== currentUser.role) {
        setError("Kendi rolünüzü değiştiremezsiniz");
        return;
      }

      // Prevent deactivating self
      if (userId === currentUser?.id && !data.is_active) {
        setError("Kendi hesabınızı pasif yapamazsınız");
        return;
      }

      await updateMutation.mutateAsync(data);
      router.push("/users");
    } catch (err: any) {
      setError(
        err.response?.data?.error?.message ||
          "Kullanıcı güncellenirken bir hata oluştu"
      );
    }
  };

  if (isLoading) {
    return (
      <div className="p-6 text-center text-slate-500">Yükleniyor...</div>
    );
  }

  if (!user) {
    return (
      <div className="p-6 text-center text-red-500">Kullanıcı bulunamadı</div>
    );
  }

  const isSelf = userId === currentUser?.id;

  return (
    <div className="p-6 max-w-2xl mx-auto space-y-6">
      {/* Header */}
      <div>
        <button
          onClick={() => router.back()}
          className="inline-flex items-center text-slate-600 hover:text-slate-900 mb-4"
        >
          <ArrowLeft className="w-4 h-4 mr-2" />
          Geri Dön
        </button>
        <h1 className="text-2xl font-bold text-slate-900">
          Kullanıcı Düzenle
        </h1>
        <p className="text-sm text-slate-600 mt-1">{user.email}</p>
      </div>

      {/* Warning for self-edit */}
      {isSelf && (
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 flex items-start">
          <AlertCircle className="w-5 h-5 text-amber-600 mr-3 flex-shrink-0 mt-0.5" />
          <div className="flex-1">
            <p className="text-sm text-amber-900 font-medium">Dikkat</p>
            <p className="text-sm text-amber-800 mt-1">
              Kendi hesabınızı düzenliyorsunuz. Rolünüzü değiştiremez ve
              hesabınızı pasif yapamazsınız.
            </p>
          </div>
        </div>
      )}

      {/* Error Alert */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 flex items-start">
          <AlertCircle className="w-5 h-5 text-red-500 mr-3 flex-shrink-0 mt-0.5" />
          <div className="flex-1">
            <p className="text-sm text-red-800">{error}</p>
          </div>
        </div>
      )}

      {/* Form */}
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        <div className="bg-white rounded-lg border border-slate-200 p-6 space-y-4">
          {/* Full Name */}
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Ad Soyad <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              {...register("full_name")}
              className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            {errors.full_name && (
              <p className="mt-1 text-sm text-red-600">
                {errors.full_name.message}
              </p>
            )}
          </div>

          {/* Email (read-only) */}
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              E-posta Adresi
            </label>
            <input
              type="email"
              value={user.email}
              disabled
              className="w-full px-4 py-2 border border-slate-300 rounded-lg bg-slate-50 text-slate-500 cursor-not-allowed"
            />
            <p className="mt-1 text-xs text-slate-500">
              E-posta adresi değiştirilemez
            </p>
          </div>

          {/* Role */}
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Rol <span className="text-red-500">*</span>
            </label>
            <select
              {...register("role")}
              disabled={isSelf}
              className={`w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent ${
                isSelf ? "bg-slate-50 text-slate-500 cursor-not-allowed" : ""
              }`}
            >
              {ROLE_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            {errors.role && (
              <p className="mt-1 text-sm text-red-600">{errors.role.message}</p>
            )}
            {isSelf && (
              <p className="mt-1 text-xs text-slate-500">
                Kendi rolünüzü değiştiremezsiniz
              </p>
            )}
          </div>

          {/* Active Status */}
          <div>
            <label className="flex items-center space-x-3">
              <input
                type="checkbox"
                {...register("is_active")}
                disabled={isSelf}
                className={`w-5 h-5 text-blue-600 border-slate-300 rounded focus:ring-blue-500 ${
                  isSelf ? "cursor-not-allowed opacity-50" : ""
                }`}
              />
              <span className="text-sm font-medium text-slate-700">
                Kullanıcı aktif
              </span>
            </label>
            {isSelf && (
              <p className="mt-1 text-xs text-slate-500 ml-8">
                Kendi hesabınızı pasif yapamazsınız
              </p>
            )}
            {!isSelf && (
              <p className="mt-1 text-xs text-slate-500 ml-8">
                Pasif kullanıcılar sisteme giriş yapamaz
              </p>
            )}
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center justify-end space-x-3">
          <button
            type="button"
            onClick={() => router.back()}
            className="px-6 py-2 border border-slate-300 text-slate-700 rounded-lg hover:bg-slate-50 font-medium"
          >
            İptal
          </button>
          <button
            type="submit"
            disabled={updateMutation.isPending}
            className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-slate-300 disabled:cursor-not-allowed font-medium flex items-center"
          >
            {updateMutation.isPending && (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            )}
            Değişiklikleri Kaydet
          </button>
        </div>
      </form>
    </div>
  );
}
