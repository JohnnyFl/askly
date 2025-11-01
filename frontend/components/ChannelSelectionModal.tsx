'use client';

import { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';

interface Channel {
  id: string;
  name: string;
  description: string;
  thumbnail_url?: string;
}

export function ChannelSelectionModal({
  open,
  channels,
  onSelect,
  onClose
}: {
  open: boolean;
  channels: Channel[];
  onSelect: (channelId: string) => void;
  onClose?: () => void;
}) {
  const [selected, setSelected] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSelect = async () => {
    if (!selected) return;

    setLoading(true);
    try {
      await onSelect(selected);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="text-2xl">Choose Your Channel</DialogTitle>
          <DialogDescription className="text-base pt-2">
            Free users can access one channel. Choose wisely! Upgrade to Pro to access both channels.
          </DialogDescription>
        </DialogHeader>
        <div className="py-4">
          <div className="space-y-3 mb-6">
            {channels.map((channel) => (
              <Card
                key={channel.id}
                onClick={() => setSelected(channel.id)}
                className={`p-4 cursor-pointer transition-all hover:shadow-md ${
                  selected === channel.id
                    ? 'border-[#FB7185] ring-2 ring-[#FB7185]/20 bg-pink-50/50'
                    : 'hover:border-[#FB7185]/50'
                }`}
              >
                <div className="flex items-center gap-4">
                  {channel.thumbnail_url && (
                    <Avatar className="h-12 w-12 ring-2 ring-pink-100">
                      <AvatarImage src={channel.thumbnail_url} alt={channel.name} />
                      <AvatarFallback className="bg-gradient-to-br from-[#FB7185] to-[#EF4444] text-white font-bold">
                        {channel.name.substring(0, 2).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                  )}
                  <div className="flex-1">
                    <h3 className="font-semibold text-slate-900">{channel.name}</h3>
                    <p className="text-sm text-slate-600 line-clamp-2">{channel.description}</p>
                  </div>
                </div>
              </Card>
            ))}
          </div>

          <Button
            onClick={handleSelect}
            disabled={!selected || loading}
            className="w-full bg-gradient-to-r from-[#FB7185] to-[#EF4444] hover:from-[#EF4444] hover:to-[#DC2626] shadow-lg rounded-xl"
          >
            {loading ? 'Saving...' : 'Select Channel'}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
