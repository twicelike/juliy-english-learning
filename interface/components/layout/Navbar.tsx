"use client";

import Link from "next/link";
import { useAuthStore } from "@/store/auth-store";
import { LearningAPI } from "@/lib/learning-api";
import { useState } from "react";

export default function Navbar() {
  const { user } = useAuthStore();
  const [isAdvancing, setIsAdvancing] = useState(false);

  const handleAdvanceTime = async () => {
    if (isAdvancing) return;
    setIsAdvancing(true);
    try {
      await LearningAPI.debugAdvanceTime(1);
      alert("Time advanced by 1 day! Please refresh the page to see changes.");
      window.location.reload();
    } catch (e: any) {
      alert("Failed to advance time: " + e.message);
    } finally {
      setIsAdvancing(false);
    }
  };

  return (
    <nav className="fixed top-0 inset-x-0 z-50 border-b border-white/10 bg-black/30 backdrop-blur-md">
      <div className="mx-auto max-w-6xl px-6 h-16 flex items-center justify-between">
        {/* Logo */}
        <Link href="/" className="flex items-center gap-2 group">
          <span className="text-2xl">📖</span>
          <span className="font-bold text-white tracking-tight text-lg group-hover:text-indigo-300 transition-colors">
            JULIY
          </span>
        </Link>

        {/* Nav links */}
        <div className="hidden md:flex items-center gap-8 text-sm text-white/70">
          <Link href="#" className="hover:text-white transition-colors">Features</Link>
          <Link href="#" className="hover:text-white transition-colors">How it works</Link>
        </div>

        {/* CTA */}
        <div className="flex items-center gap-4">
          {user ? (
            <>
              {/* DEV ONLY BUTTON */}
              <button
                onClick={handleAdvanceTime}
                disabled={isAdvancing}
                className="hidden md:flex items-center gap-2 px-3 py-1.5 text-xs font-mono font-medium text-amber-300 bg-amber-500/10 hover:bg-amber-500/20 border border-amber-500/20 rounded-lg transition-colors"
                title="Fast forward 1 day (Dev Tool)"
              >
                {isAdvancing ? "..." : "⏩ +1 Day"}
              </button>

              <Link href="/learn" className="text-sm text-white/70 hover:text-white transition-colors">
                Study
              </Link>
              <Link href="/word-sets/explore" className="text-sm text-white/70 hover:text-white transition-colors">
                Explore
              </Link>
              <Link href="/word-sets" className="text-sm text-white/70 hover:text-white transition-colors">
                My Sets
              </Link>
              <Link
                href="/profile"
                className="px-4 py-2 text-sm font-medium text-white bg-white/10 hover:bg-white/20 rounded-full transition-colors"
              >
                Profile
              </Link>
            </>
          ) : (
            <Link
              href="/login"
              className="px-4 py-2 text-sm font-medium text-white bg-indigo-500 hover:bg-indigo-600 rounded-full transition-colors"
            >
              Sign In
            </Link>
          )}
        </div>
      </div>
    </nav>
  );
}
