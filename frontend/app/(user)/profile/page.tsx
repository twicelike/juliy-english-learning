"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { AuthAPI, BookmarkDTO } from "@/lib/auth-api";
import { useAuthStore } from "@/store/auth-store";

export default function ProfilePage() {
  const router = useRouter();
  const { user, clearAuth } = useAuthStore();
  const [bookmarks, setBookmarks] = useState<BookmarkDTO[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!user) {
      // Allow API interceptor to attempt refresh if needed
      AuthAPI.getProfile()
        .then(() => fetchBookmarks())
        .catch(() => {
          router.push("/login");
        });
    } else {
      fetchBookmarks();
    }
  }, [user, router]);

  const fetchBookmarks = async () => {
    try {
      const data = await AuthAPI.getBookmarks();
      setBookmarks(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = async () => {
    try {
      await AuthAPI.logout();
    } catch (err) {
      console.error(err);
    } finally {
      clearAuth();
      router.push("/login");
    }
  };

  const removeBookmark = async (word: string) => {
    try {
      await AuthAPI.removeBookmark(word);
      setBookmarks((prev) => prev.filter((b) => b.word !== word));
    } catch (err) {
      console.error(err);
    }
  };

  if (!user || loading) {
    return (
      <div className="min-h-[calc(100vh-4rem)] flex items-center justify-center">
        <div className="w-8 h-8 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto p-6 pt-12">
      <div className="flex flex-col md:flex-row md:items-center justify-between mb-12 gap-6">
        <div>
          <h1 className="text-3xl font-bold text-white mb-2">My Profile</h1>
          <p className="text-white/50">Welcome back, {user.name}</p>
        </div>
        <button
          onClick={handleLogout}
          className="px-6 py-2.5 rounded-xl border border-red-500/30 text-red-400 hover:bg-red-500/10 font-medium transition-colors"
        >
          Logout
        </button>
      </div>

      <div>
        <h2 className="text-xl font-bold text-white mb-6 flex items-center gap-2">
          <span className="w-6 h-1 bg-indigo-500 rounded-full"></span>
          Saved Words ({bookmarks.length})
        </h2>

        {bookmarks.length === 0 ? (
          <div className="p-10 border border-white/10 rounded-2xl bg-white/5 text-center">
            <p className="text-white/50 mb-4">You haven't saved any words yet.</p>
            <Link 
              href="/"
              className="inline-block px-6 py-2.5 bg-indigo-500 text-white rounded-xl hover:bg-indigo-600 transition-colors"
            >
              Explore Dictionary
            </Link>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {bookmarks.map((b) => (
              <div key={b.word} className="group relative p-4 bg-white/5 border border-white/10 rounded-2xl hover:bg-white/10 transition-colors flex items-center justify-between">
                <Link href={`/dictionary/${b.word}`} className="text-white font-medium hover:text-indigo-300">
                  {b.word}
                </Link>
                <button
                  onClick={() => removeBookmark(b.word)}
                  className="w-8 h-8 flex items-center justify-center rounded-full hover:bg-red-500/20 text-white/30 hover:text-red-400 transition-colors"
                  title="Remove bookmark"
                >
                  <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M3.172 5.172a4 4 0 015.656 0L10 6.343l1.172-1.171a4 4 0 115.656 5.656L10 17.657l-6.828-6.829a4 4 0 010-5.656z" clipRule="evenodd" />
                  </svg>
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
