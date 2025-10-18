'use client';

import { useEffect, useState } from 'react';
import { User, MessageSquare } from 'lucide-react';

interface Activity {
  id: string;
  message: string;
  timestamp: number;
}

const sampleActivities = [
  "Sarah just discovered marketing tools from YCombinator",
  "John asked about startup funding strategies",
  "Emma learned about content creation tools",
  "Michael explored SaaS pricing models",
  "Lisa asked about productivity apps",
  "David discovered growth hacking tips",
  "Anna explored email marketing strategies",
  "Tom asked about design resources"
];

export function LiveActivityFeed() {
  const [activities, setActivities] = useState<Activity[]>([]);

  useEffect(() => {
    const addActivity = () => {
      const newActivity: Activity = {
        id: Date.now().toString(),
        message: sampleActivities[Math.floor(Math.random() * sampleActivities.length)],
        timestamp: Date.now()
      };

      setActivities((prev) => [newActivity, ...prev.slice(0, 2)]);
    };

    addActivity();
    const interval = setInterval(addActivity, 8000);

    return () => clearInterval(interval);
  }, []);

  return (
    <div className="fixed bottom-8 left-8 z-50 hidden lg:block">
      <div className="space-y-2 w-80">
        {activities.map((activity, index) => (
          <div
            key={activity.id}
            className="animate-slide-in-left bg-white/95 backdrop-blur-md rounded-2xl shadow-xl border border-pink-100/50 p-4"
            style={{ animationDelay: `${index * 100}ms` }}
          >
            <div className="flex items-start gap-3">
              <div className="h-10 w-10 bg-gradient-to-br from-[#FB7185] to-[#EF4444] rounded-xl flex items-center justify-center shrink-0">
                <User className="h-5 w-5 text-white" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm text-slate-700 leading-snug">{activity.message}</p>
                <div className="flex items-center gap-1 mt-1">
                  <MessageSquare className="h-3 w-3 text-[#EF4444]" />
                  <span className="text-xs text-slate-500">Just now</span>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
