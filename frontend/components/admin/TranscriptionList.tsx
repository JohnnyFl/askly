'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {
  Search,
  MoreVertical,
  Download,
  Trash2,
  FileText,
  Clock,
  CheckCircle2,
  AlertCircle,
  Loader2,
} from 'lucide-react';

export type Transcription = {
  id: string;
  file_name: string;
  file_size: number;
  upload_date: string;
  status: 'processing' | 'completed' | 'failed';
  error_message?: string;
};

type TranscriptionListProps = {
  channelId: string;
  transcriptions: Transcription[];
  loading: boolean;
  onDelete?: (transcriptionId: string) => Promise<void>;
  onDownload?: (transcriptionId: string) => Promise<void>;
};

export function TranscriptionList({
  channelId,
  transcriptions,
  loading,
  onDelete,
  onDownload,
}: TranscriptionListProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [downloadingId, setDownloadingId] = useState<string | null>(null);

  const filteredTranscriptions = transcriptions.filter((t) =>
    t.file_name?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
  };

  const handleDelete = async (id: string) => {
    if (!onDelete) return;
    setDeletingId(id);
    try {
      await onDelete(id);
    } finally {
      setDeletingId(null);
      setDeleteId(null);
    }
  };

  const handleDownload = async (id: string) => {
    if (!onDownload) return;
    setDownloadingId(id);
    try {
      await onDownload(id);
    } finally {
      setDownloadingId(null);
    }
  };

  const getStatusBadge = (status: Transcription['status']) => {
    switch (status) {
      case 'completed':
        return (
          <Badge className="bg-green-500 hover:bg-green-600 gap-1">
            <CheckCircle2 className="h-3 w-3" />
            Completed
          </Badge>
        );
      case 'processing':
        return (
          <Badge className="bg-blue-500 hover:bg-blue-600 gap-1">
            <Loader2 className="h-3 w-3 animate-spin" />
            Processing
          </Badge>
        );
      case 'failed':
        return (
          <Badge variant="destructive" className="gap-1">
            <AlertCircle className="h-3 w-3" />
            Failed
          </Badge>
        );
    }
  };

  if (loading) {
    return (
      <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50">
        <CardContent className="p-12">
          <div className="flex items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-[#6D28D9]" />
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <Card className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-purple-100/50">
        <CardHeader className="p-6 pb-4">
          <div className="flex items-center justify-between mb-4">
            <div>
              <CardTitle className="text-2xl font-bold text-slate-900">
                Uploaded Transcriptions
              </CardTitle>
              <CardDescription className="text-base mt-1">
                {transcriptions.length} file{transcriptions.length !== 1 ? 's' : ''} uploaded
              </CardDescription>
            </div>
          </div>

          {/* Search Bar */}
          {transcriptions.length > 0 && (
            <div className="relative">
              <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
              <Input
                type="text"
                placeholder="Search files..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-11 h-12 rounded-xl border-2 border-purple-100 focus:border-[#6D28D9]"
              />
            </div>
          )}
        </CardHeader>

        <CardContent className="p-6 pt-0">
          {filteredTranscriptions.length === 0 ? (
            <div className="text-center py-12 text-slate-500">
              <FileText className="h-16 w-16 mx-auto mb-4 opacity-30 text-[#6D28D9]" />
              <p className="text-lg font-medium">
                {searchQuery ? 'No matching files found' : 'No transcriptions uploaded yet'}
              </p>
              <p className="text-sm mt-1">
                {searchQuery
                  ? 'Try adjusting your search terms'
                  : 'Upload transcription files to get started'}
              </p>
            </div>
          ) : (
            <div className="rounded-2xl border border-purple-100 overflow-hidden">
              <Table>
                <TableHeader>
                  <TableRow className="bg-purple-50/50 hover:bg-purple-50">
                    <TableHead className="font-bold text-slate-900">File Name</TableHead>
                    <TableHead className="font-bold text-slate-900">Size</TableHead>
                    <TableHead className="font-bold text-slate-900">Upload Date</TableHead>
                    <TableHead className="font-bold text-slate-900">Status</TableHead>
                    <TableHead className="font-bold text-slate-900 text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredTranscriptions.map((transcription) => (
                    <TableRow key={transcription.id} className="hover:bg-purple-50/30">
                      <TableCell>
                        <div className="flex items-center gap-3">
                          <div className="h-10 w-10 bg-gradient-to-br from-[#6D28D9] to-[#5B21B6] rounded-xl flex items-center justify-center shrink-0">
                            <FileText className="h-5 w-5 text-white" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="font-medium text-slate-900 truncate">
                              {transcription.file_name}
                            </p>
                            {transcription.status === 'failed' && transcription.error_message && (
                              <p className="text-xs text-red-600 truncate">
                                {transcription.error_message}
                              </p>
                            )}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell className="text-slate-600">
                        {formatFileSize(transcription.file_size)}
                      </TableCell>
                      <TableCell className="text-slate-600">
                        <div className="flex items-center gap-2">
                          <Clock className="h-4 w-4 text-slate-400" />
                          {new Date(transcription.upload_date).toLocaleDateString('en-US', {
                            month: 'short',
                            day: 'numeric',
                            year: 'numeric',
                            hour: '2-digit',
                            minute: '2-digit',
                          })}
                        </div>
                      </TableCell>
                      <TableCell>{getStatusBadge(transcription.status)}</TableCell>
                      <TableCell className="text-right">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8 rounded-lg hover:bg-purple-100"
                            >
                              <MoreVertical className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end" className="rounded-xl">
                            <DropdownMenuItem
                              onClick={() => handleDownload(transcription.id)}
                              disabled={
                                transcription.status !== 'completed' ||
                                downloadingId === transcription.id
                              }
                              className="gap-2 cursor-pointer rounded-lg"
                            >
                              {downloadingId === transcription.id ? (
                                <Loader2 className="h-4 w-4 animate-spin" />
                              ) : (
                                <Download className="h-4 w-4" />
                              )}
                              Download
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={() => setDeleteId(transcription.id)}
                              disabled={deletingId === transcription.id}
                              className="gap-2 cursor-pointer text-red-600 focus:text-red-600 rounded-lg"
                            >
                              {deletingId === transcription.id ? (
                                <Loader2 className="h-4 w-4 animate-spin" />
                              ) : (
                                <Trash2 className="h-4 w-4" />
                              )}
                              Delete
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={!!deleteId} onOpenChange={(open) => !open && setDeleteId(null)}>
        <AlertDialogContent className="rounded-3xl">
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Transcription</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete this transcription? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel className="rounded-xl">Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => deleteId && handleDelete(deleteId)}
              className="rounded-xl bg-red-600 hover:bg-red-700"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}