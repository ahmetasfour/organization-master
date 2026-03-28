"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { AlertCircle, CheckCircle, Loader2, ArrowRight } from "lucide-react";
import api from "@/lib/api/client";
import { ApplicationStatus, MembershipType } from "@membership/shared-types";

interface ActionPanelProps {
  applicationId: string;
  currentStatus: string;
  membershipType: string;
  userRole: string;
}

interface StatusAction {
  label: string;
  targetStatus: string;
  description: string;
  variant: "primary" | "success" | "warning";
  allowedRoles: string[];
}

// Define manual transition actions that koordinator/admin can trigger
const STATUS_ACTIONS: Record<string, StatusAction[]> = {
  // Asil & Akademik
  referans_tamamlandı: [
    {
      label: "YK Ön İncelemeye Gönder",
      targetStatus: ApplicationStatus.YKOnIncelemede,
      description: "Referans süreci tamamlandı, YK ön incelemesine başla",
      variant: "primary",
      allowedRoles: ["koordinator", "admin", "yk"],
    },
  ],
  itibar_temiz: [
    {
      label: "Gündeme Al",
      targetStatus: ApplicationStatus.Gundemde,
      description: "İtibar taraması temiz, YK final oylamasına al",
      variant: "success",
      allowedRoles: ["koordinator", "admin", "yk"],
    },
  ],

  // Profesyonel & Öğrenci - no manual actions needed (auto-advances)
  
  // Onursal
  öneri_alındı: [
    {
      label: "YK Ön İncelemeye Gönder",
      targetStatus: ApplicationStatus.YKOnIncelemede,
      description: "Onursal öneriyi YK ön incelemesine gönder",
      variant: "primary",
      allowedRoles: ["koordinator", "admin", "yk"],
    },
  ],
  ön_onaylandı_onursal: [
    {
      label: "YİK Değerlendirmesine Gönder",
      targetStatus: ApplicationStatus.YIKDegerlendirmede,
      description: "YK ön onayı tamamlandı, YİK değerlendirmesine başla",
      variant: "primary",
      allowedRoles: ["koordinator", "admin", "yk"],
    },
  ],
};

// Statuses that are waiting for automatic progression (show info only)
const WAITING_STATUSES: Record<string, { message: string; icon: "clock" | "vote" | "check" }> = {
  referans_bekleniyor: {
    message: "Referans yanıtları bekleniyor. Tüm referanslar yanıt verdiğinde otomatik ilerleyecek.",
    icon: "clock",
  },
  yk_ön_incelemede: {
    message: "YK ön inceleme oylaması devam ediyor. Tüm oylar toplandığında ilerleyecek.",
    icon: "vote",
  },
  danışma_sürecinde: {
    message: "Danışma yanıtları bekleniyor. Tüm danışmalar tamamlandığında ilerleyecek.",
    icon: "clock",
  },
  itibar_taramasında: {
    message: "İtibar taraması devam ediyor. Tüm yanıtlar toplandığında ilerleyecek.",
    icon: "clock",
  },
  gündemde: {
    message: "YK final oylaması bekleniyor.",
    icon: "vote",
  },
  yik_değerlendirmede: {
    message: "YİK değerlendirme süresi devam ediyor. Süre bittiğinde veya itiraz geldiyse ilerleyecek.",
    icon: "clock",
  },
  ön_onaylandı: {
    message: "İtibar taraması kişilerini ekleyin. Reputasyon sekmesinden 10 kişi ekleyebilirsiniz.",
    icon: "check",
  },
};

export default function ActionPanel({
  applicationId,
  currentStatus,
  membershipType,
  userRole,
}: ActionPanelProps) {
  const [selectedAction, setSelectedAction] = useState<StatusAction | null>(null);
  const queryClient = useQueryClient();

  const advanceMutation = useMutation({
    mutationFn: async (targetStatus: string) => {
      const response = await api.patch(`/applications/${applicationId}/advance`, {
        target_status: targetStatus,
      });
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["application", applicationId] });
      queryClient.invalidateQueries({ queryKey: ["applications"] });
      setSelectedAction(null);
    },
  });

  // Get available actions for current status
  const availableActions = STATUS_ACTIONS[currentStatus]?.filter((action) =>
    action.allowedRoles.includes(userRole)
  ) || [];

  // Get waiting message if applicable
  const waitingInfo = WAITING_STATUSES[currentStatus];

  // Don't render anything if no actions and no waiting message
  if (availableActions.length === 0 && !waitingInfo) {
    return null;
  }

  const handleAdvance = (action: StatusAction) => {
    if (selectedAction?.targetStatus === action.targetStatus) {
      // Confirm and execute
      advanceMutation.mutate(action.targetStatus);
    } else {
      // First click - show confirmation
      setSelectedAction(action);
    }
  };

  const handleCancel = () => {
    setSelectedAction(null);
  };

  return (
    <div className="bg-white border border-slate-200 rounded-lg p-6 mb-6">
      {/* Waiting Status Info */}
      {waitingInfo && availableActions.length === 0 && (
        <div className="flex items-start space-x-3">
          <div className="flex-shrink-0">
            {waitingInfo.icon === "clock" && (
              <div className="w-10 h-10 bg-blue-100 rounded-full flex items-center justify-center">
                <Loader2 className="w-5 h-5 text-blue-600 animate-spin" />
              </div>
            )}
            {waitingInfo.icon === "vote" && (
              <div className="w-10 h-10 bg-purple-100 rounded-full flex items-center justify-center">
                <CheckCircle className="w-5 h-5 text-purple-600" />
              </div>
            )}
            {waitingInfo.icon === "check" && (
              <div className="w-10 h-10 bg-green-100 rounded-full flex items-center justify-center">
                <CheckCircle className="w-5 h-5 text-green-600" />
              </div>
            )}
          </div>
          <div className="flex-1">
            <h3 className="text-sm font-semibold text-slate-900 mb-1">
              Otomatik İlerleme Bekleniyor
            </h3>
            <p className="text-sm text-slate-600">{waitingInfo.message}</p>
          </div>
        </div>
      )}

      {/* Available Actions */}
      {availableActions.length > 0 && (
        <div className="space-y-4">
          <div className="flex items-center space-x-2 mb-3">
            <ArrowRight className="w-5 h-5 text-slate-500" />
            <h3 className="text-sm font-semibold text-slate-900">
              Durum İlerletme İşlemleri
            </h3>
          </div>

          {/* Action Buttons */}
          <div className="space-y-3">
            {availableActions.map((action) => {
              const isSelected = selectedAction?.targetStatus === action.targetStatus;
              const isLoading =
                advanceMutation.isPending &&
                selectedAction?.targetStatus === action.targetStatus;

              return (
                <div key={action.targetStatus} className="space-y-2">
                  <button
                    onClick={() => handleAdvance(action)}
                    disabled={advanceMutation.isPending}
                    className={`
                      w-full flex items-center justify-between p-4 rounded-lg border-2 transition-all
                      ${
                        isSelected
                          ? "border-blue-500 bg-blue-50"
                          : "border-slate-200 hover:border-blue-300 hover:bg-slate-50"
                      }
                      ${advanceMutation.isPending ? "opacity-50 cursor-not-allowed" : "cursor-pointer"}
                    `}
                  >
                    <div className="flex-1 text-left">
                      <div className="flex items-center space-x-2">
                        <span className="font-semibold text-slate-900">
                          {action.label}
                        </span>
                        {isLoading && (
                          <Loader2 className="w-4 h-4 text-blue-600 animate-spin" />
                        )}
                      </div>
                      <p className="text-sm text-slate-600 mt-1">
                        {action.description}
                      </p>
                    </div>
                    <ArrowRight
                      className={`w-5 h-5 flex-shrink-0 ml-3 ${
                        isSelected ? "text-blue-600" : "text-slate-400"
                      }`}
                    />
                  </button>

                  {/* Confirmation Panel */}
                  {isSelected && !isLoading && (
                    <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 flex items-start justify-between">
                      <div className="flex items-start space-x-3">
                        <AlertCircle className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" />
                        <div>
                          <p className="text-sm font-medium text-amber-900">
                            Bu işlemi onaylıyor musunuz?
                          </p>
                          <p className="text-sm text-amber-700 mt-1">
                            Başvuru durumu{" "}
                            <span className="font-semibold">
                              {action.targetStatus}
                            </span>{" "}
                            olarak güncellenecek.
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center space-x-2 ml-4">
                        <button
                          onClick={handleCancel}
                          className="px-3 py-1.5 text-sm font-medium text-slate-700 bg-white border border-slate-300 rounded-md hover:bg-slate-50"
                        >
                          İptal
                        </button>
                        <button
                          onClick={() => handleAdvance(action)}
                          className="px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700"
                        >
                          Onayla
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>

          {/* Error Message */}
          {advanceMutation.isError && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-4 flex items-start space-x-3">
              <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
              <div>
                <p className="text-sm font-medium text-red-900">İşlem Hatası</p>
                <p className="text-sm text-red-700 mt-1">
                  {(advanceMutation.error as any)?.response?.data?.error?.message ||
                    "Durum güncellenirken bir hata oluştu."}
                </p>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
