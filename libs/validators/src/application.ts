import { z } from "zod";

// ─── Reference Schema ──────────────────────────────────────────────────────────

export const referenceInputSchema = z.object({
  referee_name: z
    .string()
    .min(2, "Referans adı en az 2 karakter olmalıdır")
    .max(255, "Referans adı en fazla 255 karakter olmalıdır"),
  referee_email: z
    .string()
    .email("Geçerli bir e-posta adresi giriniz"),
});

// ─── Base Application Schema ───────────────────────────────────────────────────

const baseApplicationSchema = z.object({
  applicant_name: z
    .string()
    .min(2, "İsim en az 2 karakter olmalıdır")
    .max(255, "İsim en fazla 255 karakter olmalıdır"),
  applicant_email: z
    .string()
    .email("Geçerli bir e-posta adresi giriniz"),
  applicant_phone: z
    .string()
    .optional()
    .refine(
      (val) => !val || /^(\+90|0)?[0-9]{10}$/.test(val.replace(/\s/g, "")),
      "Geçerli bir Türkiye telefon numarası giriniz"
    ),
  photo_url: z
    .string()
    .url("Geçerli bir URL giriniz")
    .optional()
    .or(z.literal("")),
});

// ─── Asil Üye Application Schema ───────────────────────────────────────────────

export const asilApplicationSchema = baseApplicationSchema.extend({
  membership_type: z.literal("asil"),
  linkedin_url: z
    .string()
    .url("Geçerli bir LinkedIn URL'i giriniz")
    .refine(
      (url) =>
        url.startsWith("https://www.linkedin.com/") ||
        url.startsWith("https://linkedin.com/"),
      "LinkedIn profil bağlantısı gereklidir"
    ),
  references: z
    .array(referenceInputSchema)
    .min(3, "En az 3 referans gereklidir")
    .max(10, "En fazla 10 referans ekleyebilirsiniz"),
});

export type AsilApplicationInput = z.infer<typeof asilApplicationSchema>;

// ─── Akademik Üye Application Schema ───────────────────────────────────────────

export const akademikApplicationSchema = baseApplicationSchema.extend({
  membership_type: z.literal("akademik"),
  linkedin_url: z
    .string()
    .url("Geçerli bir LinkedIn URL'i giriniz")
    .refine(
      (url) =>
        url.startsWith("https://www.linkedin.com/") ||
        url.startsWith("https://linkedin.com/"),
      "LinkedIn profil bağlantısı gereklidir"
    ),
  references: z
    .array(referenceInputSchema)
    .min(3, "En az 3 referans gereklidir")
    .max(10, "En fazla 10 referans ekleyebilirsiniz"),
});

export type AkademikApplicationInput = z.infer<
  typeof akademikApplicationSchema
>;

// ─── Profesyonel Üye Application Schema ────────────────────────────────────────

export const profesyonelApplicationSchema = baseApplicationSchema.extend({
  membership_type: z.literal("profesyonel"),
  linkedin_url: z
    .string()
    .url("Geçerli bir LinkedIn URL'i giriniz")
    .optional()
    .or(z.literal("")),
  photo_url: z
    .string()
    .url("Geçerli bir fotoğraf URL'i giriniz")
    .min(1, "Profesyonel başvurular için fotoğraf zorunludur"),
});

export type ProfesyonelApplicationInput = z.infer<
  typeof profesyonelApplicationSchema
>;

// ─── Öğrenci Üye Application Schema ────────────────────────────────────────────

export const ogrenciApplicationSchema = baseApplicationSchema.extend({
  membership_type: z.literal("ogrenci"),
  linkedin_url: z
    .string()
    .url("Geçerli bir LinkedIn URL'i giriniz")
    .optional()
    .or(z.literal("")),
  photo_url: z
    .string()
    .url("Geçerli bir fotoğraf URL'i giriniz")
    .min(1, "Öğrenci başvurular için fotoğraf zorunludur"),
});

export type OgrenciApplicationInput = z.infer<typeof ogrenciApplicationSchema>;

// ─── Union Schema for Dynamic Forms ─────────────────────────────────────────────

export const applicationSchema = z.discriminatedUnion("membership_type", [
  asilApplicationSchema,
  akademikApplicationSchema,
  profesyonelApplicationSchema,
  ogrenciApplicationSchema,
]);

export type ApplicationInput = z.infer<typeof applicationSchema>;

// ─── Helper to get schema by type ───────────────────────────────────────────────

export function getApplicationSchemaByType(type: string) {
  switch (type) {
    case "asil":
      return asilApplicationSchema;
    case "akademik":
      return akademikApplicationSchema;
    case "profesyonel":
      return profesyonelApplicationSchema;
    case "ogrenci":
      return ogrenciApplicationSchema;
    default:
      throw new Error(`Unknown membership type: ${type}`);
  }
}
