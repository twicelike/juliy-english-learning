"use client";

import React, { useEffect, useState } from "react";
import apiClient from "@/lib/api-client";
import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer, Legend } from "recharts";

import { Button } from "@/components/ui/Button";

const COLORS = {
  pending: "#eab308", // Yellow
  processing: "#3b82f6", // Blue
  completed: "#22c55e", // Green
  failed: "#ef4444", // Red
};

export default function TranslationJobsAdminPage() {
  const [stats, setStats] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [retrying, setRetrying] = useState(false);

  const fetchStats = async () => {
    try {
      const res = await apiClient.get("/admin/translation-jobs/stats");
      if (res.data?.success) {
        const rawStats = res.data.data || {};
        
        const chartData = Object.keys(rawStats).map((key) => ({
          name: key.charAt(0).toUpperCase() + key.slice(1),
          value: rawStats[key],
          statusKey: key,
        }));
        
        setStats(chartData);
      }
    } catch (err) {
      console.error("Failed to fetch translation jobs stats", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();
    
    // Refresh every 5 seconds
    const interval = setInterval(fetchStats, 5000);
    return () => clearInterval(interval);
  }, []);

  const handleRetryFailed = async () => {
    try {
      setRetrying(true);
      const res = await apiClient.post("/admin/translation-jobs/retry-failed");
      if (res.data?.success) {
        alert(`Successfully queued ${res.data.data?.count || 0} failed jobs for retry!`);
        fetchStats();
      }
    } catch (err) {
      console.error("Failed to retry jobs", err);
      alert("Failed to retry jobs. Check console for details.");
    } finally {
      setRetrying(false);
    }
  };

  if (loading) {
    return <div className="flex h-64 items-center justify-center">Loading stats...</div>;
  }

  const total = stats.reduce((acc, curr) => acc + curr.value, 0);
  const failedCount = stats.find(s => s.statusKey === "failed")?.value || 0;

  return (
    <div className="bg-white/5 border border-white/10 rounded-3xl p-8">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold">Job Status Distribution (Total: {total})</h2>
        {failedCount > 0 && (
          <Button 
            onClick={handleRetryFailed} 
            disabled={retrying}
            className="bg-red-500 hover:bg-red-600 text-white"
          >
            {retrying ? "Retrying..." : `Retry ${failedCount} Failed Jobs`}
          </Button>
        )}
      </div>
      
      {stats.length > 0 ? (
        <div className="h-80 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={stats}
                cx="50%"
                cy="50%"
                innerRadius={80}
                outerRadius={120}
                paddingAngle={5}
                dataKey="value"
              >
                {stats.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={COLORS[entry.statusKey as keyof typeof COLORS] || "#8884d8"} />
                ))}
              </Pie>
              <Tooltip 
                formatter={(value: any) => [`${value} jobs`, "Count"]}
                contentStyle={{ backgroundColor: "#1f2937", border: "none", borderRadius: "8px", color: "#fff" }}
              />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
        </div>
      ) : (
        <div className="text-center py-12 text-gray-500">No data available</div>
      )}
    </div>
  );
}
