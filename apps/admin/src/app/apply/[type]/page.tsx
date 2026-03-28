"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm, useFieldArray } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { AlertCircle, Plus, Trash2, ArrowLeft } from "lucide-react";
import Link from "next/link";
import { z } from "zod";
import {
  getApplicationSchemaByType,
  type ApplicationInput,
} from "@membership/validators";
import api from "@/lib/api/client";

const MEMBERSHIP_TYPES = {
  asil: "Asil Üyelik",
  akademik: "Akademik Üyelik",
  profesyonel: "Profesyonel Üyelik",
  ogrenci: "Öğrenci Üyelik",
} as const;

type MembershipType = keyof typeof MEMBERSHIP_TYPES;

export default function ApplicationFormPage({
  params,
}: {
  params: { type: string };
}) {
  const router = useRouter();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const membershipType = params.type as MembershipType;

  // Validate membership type
  if (!MEMBERSHIP_TYPES[membershipType]) {
    router.push("/apply");
    return null;
  }

  const requiresReferences =
    membershipType === "asil" || membershipType === "akademik";
  const requiresPhoto =
    membershipType === "profesyonel" || membershipType === "ogrenci";

  const schema = getApplicationSchemaByType(membershipType);

  const {
    register,
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<any>({
    resolver: zodResolver(schema as any),
    defaultValues: {
      membership_type: membershipType,
      applicant_name: "",
      applicant_email: "",
      applicant_phone: "",
      linkedin_url: "",
      photo_url: "",
      references: requiresReferences
        ? [
            { referee_name: "", referee_email: "" },
            { referee_name: "", referee_email: "" },
            { referee_name: "", referee_email: "" },
          ]
        : undefined,
    },
  });

  const { fields, append, remove } = useFieldArray({
    control,
    name: "references" as any,
  });

  const onSubmit = async (data: any) => {
    try {
      setIsSubmitting(true);
      setError(null);

      const response = await api.post("/applications", data);

      // Redirect to success page with application ID
      router.push(
        `/apply/success?id=${response.data.data.application.id}&type=${membershipType}`
      );
    } catch (err: any) {
      console.error("Submit error:", err);
      setError(
        err.response?.data?.error?.message ||
          "Başvurunuz gönderilirken bir hata oluştu. Lütfen tekrar deneyin."
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-50 py-12">
      <div className="container mx-auto px-4 max-w-3xl">
        {/* Header */}
        <div className="mb-8">
          <Link
            href="/apply"
            className="inline-flex items-center text-slate-600 hover:text-slate-900 mb-4"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Geri Dön
          </Link>
          <h1 className="text-3xl font-bold text-slate-900">
            {MEMBERSHIP_TYPES[membershipType]} Başvurusu
          </h1>
          <p className="text-slate-600 mt-2">
            Lütfen aşağıdaki formu eksiksiz doldurunuz.
          </p>
        </div>

        {/* Error Alert */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6 flex items-start">
            <AlertCircle className="w-5 h-5 text-red-500 mr-3 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          </div>
        )}

        {/* Form */}
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-8">
          {/* Personal Information */}
          <div className="bg-white rounded-lg shadow-md p-6">
            <h2 className="text-xl font-semibold text-slate-900 mb-4">
              Kişisel Bilgiler
            </h2>
            <div className="space-y-4">
              {/* Full Name */}
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  Ad Soyad <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  {...register("applicant_name")}
                  className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="Adınız ve Soyadınız"
                />
                {errors.applicant_name && (
                  <p className="mt-1 text-sm text-red-600">
                    {errors.applicant_name.message as string}
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
                  {...register("applicant_email")}
                  className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="ornek@email.com"
                />
                {errors.applicant_email && (
                  <p className="mt-1 text-sm text-red-600">
                    {errors.applicant_email.message as string}
                  </p>
                )}
              </div>

              {/* Phone (optional) */}
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  Telefon Numarası
                </label>
                <input
                  type="tel"
                  {...register("applicant_phone")}
                  className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="+90 5XX XXX XX XX"
                />
                {errors.applicant_phone && (
                  <p className="mt-1 text-sm text-red-600">
                    {errors.applicant_phone.message as string}
                  </p>
                )}
              </div>

              {/* LinkedIn URL (required for asil/akademik) */}
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  LinkedIn Profil Linki{" "}
                  {requiresReferences && <span className="text-red-500">*</span>}
                </label>
                <input
                  type="url"
                  {...register("linkedin_url")}
                  className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="https://www.linkedin.com/in/username"
                />
                {errors.linkedin_url && (
                  <p className="mt-1 text-sm text-red-600">
                    {errors.linkedin_url.message as string}
                  </p>
                )}
              </div>

              {/* Photo URL (required for profesyonel/ogrenci) */}
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">
                  Fotoğraf Linki{" "}
                  {requiresPhoto && <span className="text-red-500">*</span>}
                </label>
                <input
                  type="url"
                  {...register("photo_url")}
                  className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="https://example.com/photo.jpg"
                />
                {errors.photo_url && (
                  <p className="mt-1 text-sm text-red-600">
                    {errors.photo_url.message as string}
                  </p>
                )}
                {requiresPhoto && (
                  <p className="mt-1 text-xs text-slate-500">
                    Fotoğrafınızı bir dosya paylaşım servisine yükleyip linkini
                    buraya yapıştırabilirsiniz.
                  </p>
                )}
              </div>
            </div>
          </div>

          {/* References (for asil/akademik only) */}
          {requiresReferences && (
            <div className="bg-white rounded-lg shadow-md p-6">
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h2 className="text-xl font-semibold text-slate-900">
                    Referanslar <span className="text-red-500">*</span>
                  </h2>
                  <p className="text-sm text-slate-600 mt-1">
                    En az 3 referans bilgisi giriniz
                  </p>
                </div>
                <button
                  type="button"
                  onClick={() =>
                    append({ referee_name: "", referee_email: "" })
                  }
                  className="flex items-center px-3 py-2 text-sm font-medium text-blue-600 hover:text-blue-700 border border-blue-300 rounded-lg hover:bg-blue-50"
                >
                  <Plus className="w-4 h-4 mr-1" />
                  Ekle
                </button>
              </div>

              <div className="space-y-4">
                {fields.map((field, index) => (
                  <div
                    key={field.id}
                    className="p-4 border border-slate-200 rounded-lg"
                  >
                    <div className="flex items-start justify-between mb-3">
                      <h3 className="text-sm font-medium text-slate-700">
                        Referans {index + 1}
                      </h3>
                      {fields.length > 3 && (
                        <button
                          type="button"
                          onClick={() => remove(index)}
                          className="text-red-600 hover:text-red-700"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      )}
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                      <div>
                        <label className="block text-xs font-medium text-slate-600 mb-1">
                          Ad Soyad
                        </label>
                        <input
                          type="text"
                          {...register(
                            `references.${index}.referee_name` as any
                          )}
                          className="w-full px-3 py-2 text-sm border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                          placeholder="Referans adı"
                        />
                        {(errors.references as any)?.[index]?.referee_name && (
                          <p className="mt-1 text-xs text-red-600">
                            {
                              (errors.references as any)[index]?.referee_name
                                ?.message as string
                            }
                          </p>
                        )}
                      </div>
                      <div>
                        <label className="block text-xs font-medium text-slate-600 mb-1">
                          E-posta
                        </label>
                        <input
                          type="email"
                          {...register(
                            `references.${index}.referee_email` as any
                          )}
                          className="w-full px-3 py-2 text-sm border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                          placeholder="referans@email.com"
                        />
                        {(errors.references as any)?.[index]?.referee_email && (
                          <p className="mt-1 text-xs text-red-600">
                            {
                              (errors.references as any)[index]?.referee_email
                                ?.message as string
                            }
                          </p>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>

              {errors.references && !Array.isArray(errors.references) && (
                <p className="mt-2 text-sm text-red-600">
                  {errors.references.message as string}
                </p>
              )}
            </div>
          )}

          {/* Submit Button */}
          <div className="flex justify-end">
            <button
              type="submit"
              disabled={isSubmitting}
              className="px-8 py-3 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 disabled:bg-slate-300 disabled:cursor-not-allowed transition-colors"
            >
              {isSubmitting ? "Gönderiliyor..." : "Başvuruyu Gönder"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
