"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { AlertCircle, ArrowLeft, Loader2 } from "lucide-react";
import { useCreateUser } from "@/lib/hooks/useUsers";
import { useAuthStore } from "@/lib/store/auth.store";

const ROLE_OPTIONS = [
  { value: "admin", label: "Sistem Yöneticisi" },
  { value: "yk", label: "YK Üyesi" },
  { value: "yik", label: "YİK Üyesi" },
  { value: "koordinator", label: "Koordinatör" },
  { value: "asil_uye", label: "Asil Üye" },
];

const createUserSchema = z.object({
  full_name: z
    .string()
    .min(2, "Ad soyad en az 2 karakter olmalıdır")
    .max(255, "Ad soyad en fazla 255 karakter olmalıdır"),
  email: z.string().email("Geçerli bir e-posta adresi giriniz"),
  password: z.string().min(8, "Şifre en az 8 karakter olmalıdır"),
  role: z.enum(["admin", "yk", "yik", "koordinator", "asil_uye"], {
    message: "Rol seçimi zorunludur",
  }),
});

type CreateUserForm = z.infer<typeof createUserSchema>;

export default function NewUserPage() {
  const router = useRouter();
  const { user } = useAuthStore();
  const createMutation = useCreateUser();
  const [error, setError] = useState<string | null>(null);

  // Redirect if not admin
  if (user?.role !== "admin") {
    router.push("/");
    return null;
  }

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<CreateUserForm>({
    resolver: zodResolver(createUserSchema),
  });

  const onSubmit = async (data: CreateUserForm) => {
    try {
      setError(null);
      await createMutation.mutateAsync(data);
      router.push("/users");
    } catch (err: any) {
      setError(
        err.response?.data?.error?.message ||
          "Kullanıcı oluşturulurken bir hata oluştu"
      );
    }
  };

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
          Yeni Kullanıcı Oluştur
        </h1>
        <p className="text-sm text-slate-600 mt-1">
          Sisteme yeni bir kullanıcı ekleyin
        </p>
      </div>

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
              placeholder="Kullanıcının tam adı"
            />
            {errors.full_name && (
              <p className="mt-1 text-sm text-red-600">
                {errors.full_name.message}
              </p>
            )}
          </div>

          {/* Email */}
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              E-posta Adresi <span className="text-red-500">*</span>
            </label>
            <input
              type="email"
              {...register("email")}
              className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="kullanici@example.com"
            />
            {errors.email && (
              <p className="mt-1 text-sm text-red-600">{errors.email.message}</p>
            )}
            <p className="mt-1 text-xs text-slate-500">
              Bu e-posta adresi giriş için kullanılacaktır
            </p>
          </div>

          {/* Password */}
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Şifre <span className="text-red-500">*</span>
            </label>
            <input
              type="password"
              {...register("password")}
              className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="Minimum 8 karakter"
            />
            {errors.password && (
              <p className="mt-1 text-sm text-red-600">
                {errors.password.message}
              </p>
            )}
          </div>

          {/* Role */}
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">
              Rol <span className="text-red-500">*</span>
            </label>
            <select
              {...register("role")}
              className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="">Bir rol seçin</option>
              {ROLE_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            {errors.role && (
              <p className="mt-1 text-sm text-red-600">{errors.role.message}</p>
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
            disabled={createMutation.isPending}
            className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-slate-300 disabled:cursor-not-allowed font-medium flex items-center"
          >
            {createMutation.isPending && (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            )}
            Kullanıcı Oluştur
          </button>
        </div>
      </form>
    </div>
  );
}
