"use client";

import React, { useEffect } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useAuthStore } from "@/features/auth";
import { hasAdminAccess } from "@/features/auth/types";

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const { user, status } = useAuthStore();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (status === "unauthenticated" || (user && !hasAdminAccess(user.role))) {
      router.push("/");
    }
  }, [user, status, router]);

  if (status === "loading" || !user) {
    return <div className="flex h-screen items-center justify-center">Loading...</div>;
  }

  const tabs = [
    { id: "/admin/translation-jobs", label: "Translation Jobs" },
    { id: "/admin/users", label: "Users Management" },
    { id: "/admin/settings", label: "Settings" }
  ];

  return (
    <div className="max-w-6xl mx-auto p-6 pt-12">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Admin Dashboard</h1>
        <p className="opacity-50">Manage your application data and background jobs.</p>
      </div>

      <div className="flex border-b border-white/10 mb-8 overflow-x-auto">
        {tabs.map(tab => {
          const isActive = pathname.startsWith(tab.id);
          return (
            <Link
              key={tab.id}
              href={tab.id}
              className={`px-6 py-3 font-medium text-sm whitespace-nowrap transition-colors border-b-2 ${
                isActive
                  ? "border-amber-500 text-amber-500"
                  : "border-transparent opacity-50 hover:opacity-80"
              }`}
            >
              {tab.label}
            </Link>
          );
        })}
      </div>

      <div className="min-h-[400px]">
        {children}
      </div>
    </div>
  );
}
