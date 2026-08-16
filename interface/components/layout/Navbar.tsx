"use client";

import Link from "next/link";
import { useAuthStore } from "@/store/auth-store";

export default function Navbar() {
  const { user } = useAuthStore();

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

        {/* CTA */}
        <div className="flex items-center gap-4">
          {user ? (
            <>
              <Link href="/learn" className="text-sm text-white/70 hover:text-white transition-colors">
                Study
              </Link>
              <Link href="/speaking" className="text-sm text-white/70 hover:text-white transition-colors">
                Speaking
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
