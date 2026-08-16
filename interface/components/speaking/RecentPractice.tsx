import React from "react";
import { formatDistanceToNow } from "date-fns";

export interface RecentPracticeItem {
  id: string;
  title: string;
  type: string;
  score: number;
  practicedAt: string;
}

interface RecentPracticeProps {
  items: RecentPracticeItem[];
}

export function RecentPractice({ items }: RecentPracticeProps) {
  if (items.length === 0) return null;

  return (
    <div>
      <h2 className="text-xl font-bold text-white mb-4">Recent Practice</h2>
      
      <div className="flex flex-col gap-3">
        {items.map((item) => (
          <div
            key={item.id}
            className="flex items-center justify-between p-4 rounded-xl bg-white/5 border border-white/5 hover:bg-white/10 transition-colors cursor-pointer"
          >
            <div>
              <div className="text-white font-medium mb-1">{item.title}</div>
              <div className="text-white/40 text-xs flex gap-2">
                <span>{item.type}</span>
                <span>•</span>
                <span>{formatDistanceToNow(new Date(item.practicedAt), { addSuffix: true })}</span>
              </div>
            </div>
            
            <div className="flex flex-col items-end">
              <div className="text-lg font-bold text-white">
                {item.score}<span className="text-white/40 text-sm font-normal">/100</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
