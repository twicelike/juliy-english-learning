"use client";

import { useEffect, useState, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { AuthAPI, BookmarkDTO } from "@/lib/auth-api";
import { authAPI, useAuthStore, SessionResponse } from "@/features/auth";

function ProfileContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { user, clearAuth } = useAuthStore();
  const [bookmarks, setBookmarks] = useState<BookmarkDTO[]>([]);
  const [sessions, setSessions] = useState<SessionResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState({ type: "", text: "" });
  const [resending, setResending] = useState(false);

  useEffect(() => {
    if (!user) {
      // Allow API interceptor to attempt refresh if needed
      AuthAPI.getProfile()
        .then(() => {
          fetchBookmarks();
          fetchSessions();
        })
        .catch(() => {
          router.push("/login");
        });
    } else {
      fetchBookmarks();
      fetchSessions();
    }
  }, [user, router]);

  useEffect(() => {
    const errParam = searchParams.get("error");
    const successParam = searchParams.get("success");
    if (errParam === "identity_already_linked") {
      setMessage({ type: "error", text: "This Google account is already linked to another user." });
    } else if (errParam === "linking_failed") {
      setMessage({ type: "error", text: "Failed to link Google account. Please try again." });
    } else if (successParam === "google_linked") {
      setMessage({ type: "success", text: "Google account successfully linked!" });
    }
  }, [searchParams]);

  const fetchBookmarks = async () => {
    try {
      const data = await AuthAPI.getBookmarks();
      setBookmarks(data);
    } catch (err) {
      console.error(err);
    } finally {
      // We will set loading false when sessions are also fetched, or just here is fine since they run parallel
    }
  };

  const fetchSessions = async () => {
    try {
      const data = await authAPI.getSessions();
      setSessions(data);
    } catch (err) {
      console.error("Failed to fetch sessions:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleRevokeSession = async (sessionId: string) => {
    try {
      await authAPI.revokeSession(sessionId);
      setSessions((prev) => prev.filter((s) => s.id !== sessionId));
      setMessage({ type: "success", text: "Session revoked successfully." });
    } catch (err) {
      console.error("Failed to revoke session:", err);
      setMessage({ type: "error", text: "Failed to revoke session." });
    }
  };

  const handleLogout = async () => {
    try {
      await authAPI.logout();
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

  const handleResendVerification = async () => {
    if (!user || user.email_verified) return;
    setResending(true);
    try {
      await authAPI.resendVerification(user.email);
      setMessage({ type: "success", text: "Verification email sent. Please check your inbox." });
    } catch (err: any) {
      setMessage({ type: "error", text: err.response?.data?.message || "Failed to send verification email." });
    } finally {
      setResending(false);
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
          <p className="text-white/50">Welcome back, {user.display_name}</p>
        </div>
        <div className="flex gap-3">
          {!user.has_google_linked && (
            <a
              href="/api/v1/auth/google"
              className="px-6 py-2.5 rounded-xl bg-white/5 border border-white/10 text-white hover:bg-white/10 font-medium transition-colors flex items-center gap-2"
            >
              <svg viewBox="0 0 24 24" width="16" height="16" xmlns="http://www.w3.org/2000/svg">
                <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
                <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
                <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/>
                <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
                <path d="M1 1h22v22H1z" fill="none"/>
              </svg>
              Link Google Account
            </a>
          )}
          <button
            onClick={handleLogout}
            className="px-6 py-2.5 rounded-xl border border-red-500/30 text-red-400 hover:bg-red-500/10 font-medium transition-colors"
          >
            Logout
          </button>
        </div>
      </div>

      {!user.email_verified && (
        <div className="mb-8 p-4 rounded-xl border border-yellow-500/20 bg-yellow-500/10 flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
          <div className="flex items-start gap-3">
            <svg className="w-6 h-6 text-yellow-400 shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <div>
              <h3 className="text-yellow-400 font-medium">Verify your email address</h3>
              <p className="text-yellow-400/80 text-sm mt-1">Please verify your email to secure your account. A verification link was sent during registration.</p>
            </div>
          </div>
          <button
            onClick={handleResendVerification}
            disabled={resending}
            className="shrink-0 px-4 py-2 bg-yellow-500/20 hover:bg-yellow-500/30 text-yellow-400 rounded-lg text-sm font-medium transition-colors disabled:opacity-50"
          >
            {resending ? "Sending..." : "Resend Email"}
          </button>
        </div>
      )}

      {message.text && (
        <div className={`mb-8 p-4 rounded-xl border text-sm ${
          message.type === "error" 
            ? "bg-red-500/10 border-red-500/20 text-red-400" 
            : "bg-green-500/10 border-green-500/20 text-green-400"
        }`}>
          {message.text}
        </div>
      )}

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

      <div className="mt-16">
        <h2 className="text-xl font-bold text-white mb-6 flex items-center gap-2">
          <span className="w-6 h-1 bg-red-500 rounded-full"></span>
          Active Sessions
        </h2>

        {sessions.length === 0 ? (
          <p className="text-white/50">No active sessions found.</p>
        ) : (
          <div className="space-y-4">
            {sessions.map((sess) => (
              <div key={sess.id} className="flex flex-col md:flex-row md:items-center justify-between p-4 bg-white/5 border border-white/10 rounded-2xl gap-4">
                <div>
                  <div className="flex items-center gap-3 mb-1">
                    <p className="text-white font-medium">{sess.user_agent || "Unknown Device"}</p>
                    {sess.is_current && (
                      <span className="px-2 py-0.5 rounded text-xs font-medium bg-green-500/20 text-green-400">
                        Current
                      </span>
                    )}
                  </div>
                  <div className="text-sm text-white/50 space-x-3">
                    <span>IP: {sess.ip_address}</span>
                    <span>•</span>
                    <span>Signed in: {new Date(sess.created_at).toLocaleDateString()}</span>
                  </div>
                </div>
                {!sess.is_current && (
                  <button
                    onClick={() => handleRevokeSession(sess.id)}
                    className="shrink-0 px-4 py-2 bg-red-500/10 hover:bg-red-500/20 text-red-400 rounded-xl text-sm font-medium transition-colors"
                  >
                    Revoke Access
                  </button>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default function ProfilePage() {
  return (
    <Suspense fallback={<div className="min-h-[calc(100vh-4rem)] flex items-center justify-center">Loading...</div>}>
      <ProfileContent />
    </Suspense>
  );
}
