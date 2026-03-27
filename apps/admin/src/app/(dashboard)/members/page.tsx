'use client';

import { usePublishedMembers } from '@/lib/hooks/useWebPublish';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Input } from '@/components/ui/input';
import { Loader2, Users, Search } from 'lucide-react';
import { useState, useMemo } from 'react';

const membershipTypeLabels: Record<string, string> = {
  asil: 'Asil Üye',
  akademik: 'Akademik Üye',
  profesyonel: 'Profesyonel Üye',
  ogrenci: 'Öğrenci Üye',
  onursal: 'Onursal Üye',
};

const membershipTypeColors: Record<string, string> = {
  asil: 'bg-blue-100 text-blue-800',
  akademik: 'bg-purple-100 text-purple-800',
  profesyonel: 'bg-green-100 text-green-800',
  ogrenci: 'bg-yellow-100 text-yellow-800',
  onursal: 'bg-pink-100 text-pink-800',
};

export default function MembersPage() {
  const { data: members, isLoading, error } = usePublishedMembers();
  const [searchQuery, setSearchQuery] = useState('');

  const filteredMembers = useMemo(() => {
    if (!members) return [];
    if (!searchQuery.trim()) return members;

    const query = searchQuery.toLowerCase();
    return members.filter(
      (member) =>
        member.full_name.toLowerCase().includes(query) ||
        membershipTypeLabels[member.membership_type]?.toLowerCase().includes(query)
    );
  }, [members, searchQuery]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <Card className="border-destructive">
          <CardHeader>
            <CardTitle className="text-destructive">Hata</CardTitle>
            <CardDescription>Üye listesi yüklenirken bir hata oluştu.</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      <div className="space-y-1">
        <h1 className="text-2xl font-bold tracking-tight">Web'de Yayınlanan Üyeler</h1>
        <p className="text-muted-foreground">Kamuya açık üye listesi (alfabetik sıralı)</p>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Users className="h-5 w-5 text-muted-foreground" />
              <CardTitle className="text-lg">Üye Listesi</CardTitle>
              <Badge variant="secondary">{members?.length || 0} üye</Badge>
            </div>
            <div className="relative w-64">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Üye ara..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-9"
              />
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {filteredMembers.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              {searchQuery ? (
                <>Aramanızla eşleşen üye bulunamadı.</>
              ) : (
                <>Henüz yayınlanan üye bulunmuyor.</>
              )}
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[50px]">#</TableHead>
                  <TableHead>Ad Soyad</TableHead>
                  <TableHead>Üyelik Tipi</TableHead>
                  <TableHead>Kabul Tarihi</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredMembers.map((member, index) => (
                  <TableRow key={`${member.full_name}-${index}`}>
                    <TableCell className="text-muted-foreground">
                      {index + 1}
                    </TableCell>
                    <TableCell className="font-medium">
                      {member.full_name}
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant="outline"
                        className={membershipTypeColors[member.membership_type]}
                      >
                        {membershipTypeLabels[member.membership_type] || member.membership_type}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {new Date(member.accepted_at).toLocaleDateString('tr-TR', {
                        year: 'numeric',
                        month: 'long',
                        day: 'numeric',
                      })}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}