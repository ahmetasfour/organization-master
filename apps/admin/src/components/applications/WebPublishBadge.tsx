import { Badge } from '@/components/ui/badge';
import { Globe, Lock, Clock } from 'lucide-react';

interface WebPublishBadgeProps {
  webPublishConsent: boolean | null;
  isPublished: boolean;
}

export function WebPublishBadge({ webPublishConsent, isPublished }: WebPublishBadgeProps) {
  if (webPublishConsent === null) {
    return (
      <Badge variant="outline" className="gap-1">
        <Clock className="h-3 w-3" />
        Karar Bekleniyor
      </Badge>
    );
  }

  if (isPublished) {
    return (
      <Badge variant="default" className="gap-1 bg-green-600 hover:bg-green-700">
        <Globe className="h-3 w-3" />
        Yayında
      </Badge>
    );
  }

  return (
    <Badge variant="secondary" className="gap-1">
      <Lock className="h-3 w-3" />
      İç Listede
    </Badge>
  );
}