"use client";

import { useEffect, useState, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { AuthAPI, BookmarkDTO } from "@/lib/auth-api";
import { authAPI, useAuthStore, SessionResponse } from "@/features/auth";
import { ProfileAPI, DashboardResponse, SkillAssessmentDTO } from "@/lib/profile-api";
import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { Progress } from "@/components/ui/Progress";
import { Badge } from "@/components/ui/Badge";
type Tab = "overview" | "skills" | "goals" | "assessment" | "account";

function ProfileContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { user, clearAuth } = useAuthStore();
  const [activeTab, setActiveTab] = useState<Tab>("overview");
  
  // Account state
  const [bookmarks, setBookmarks] = useState<BookmarkDTO[]>([]);
  const [sessions, setSessions] = useState<SessionResponse[]>([]);
  const [message, setMessage] = useState({ type: "", text: "" });
  const [resending, setResending] = useState(false);
  
  // DE state
  const [dashboard, setDashboard] = useState<DashboardResponse | null>(null);
  const [loading, setLoading] = useState(true);

  // Goals & Focus form state
  const [savingPreferences, setSavingPreferences] = useState(false);
  const [selectedGoals, setSelectedGoals] = useState<string[]>([]);
  const [selectedFocus, setSelectedFocus] = useState<string[]>([]);

  useEffect(() => {
    if (!user) {
      AuthAPI.getProfile()
        .then(() => fetchAllData())
        .catch(() => router.push("/login"));
    } else {
      fetchAllData();
    }
  }, [user, router]);

  const fetchAllData = async () => {
    try {
      const [dash, bmarks, sess] = await Promise.all([
        ProfileAPI.getDashboard(),
        AuthAPI.getBookmarks(),
        authAPI.getSessions(),
      ]);
      setDashboard(dash);
      setBookmarks(bmarks);
      setSessions(sess);
      
      // Initialize preferences
      setSelectedGoals(dash.profile.goals || []);
      setSelectedFocus(dash.profile.focus_areas || []);
    } catch (err) {
      console.error("Failed to load profile data:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = async () => {
    try {
      await authAPI.logout();
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

  const handleRevokeSession = async (sessionId: string) => {
    try {
      await authAPI.revokeSession(sessionId);
      setSessions((prev) => prev.filter((s) => s.id !== sessionId));
    } catch (err) {
      console.error("Failed to revoke session:", err);
    }
  };

  const handleSavePreferences = async () => {
    setSavingPreferences(true);
    try {
      await ProfileAPI.updatePreferences({
        goals: selectedGoals,
        focus_areas: selectedFocus,
      });
      setMessage({ type: "success", text: "Learning preferences updated!" });
      
      // Refresh dashboard to get new recommendations
      const dash = await ProfileAPI.getDashboard();
      setDashboard(dash);
      
      setTimeout(() => setMessage({ type: "", text: "" }), 3000);
    } catch (err) {
      setMessage({ type: "error", text: "Failed to update preferences." });
    } finally {
      setSavingPreferences(false);
    }
  };

  const toggleSelection = (item: string, list: string[], setList: (l: string[]) => void) => {
    if (list.includes(item)) {
      setList(list.filter((i) => i !== item));
    } else {
      setList([...list, item]);
    }
  };

  if (!user || loading) {
    return (
      <div className="min-h-[calc(100vh-4rem)] flex items-center justify-center">
        <div className="w-8 h-8 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  const profile = dashboard?.profile;
  const recommendation = dashboard?.recommendation;

  const renderTabContent = () => {
    switch (activeTab) {
      case "overview":
        return (
          <div className="space-y-6">
            <Card>
              <h2 className="text-xl font-bold mb-2">Your English Progress</h2>
              <div className="text-indigo-400 font-medium mb-6">Level: {profile?.completed_placement ? "Assessed" : "Pending Assessment"}</div>
              
              <div className="space-y-4">
                {profile?.skills?.map(skill => (
                  <div key={skill.skill}>
                    <div className="flex justify-between text-sm mb-1">
                      <span className="capitalize">{skill.skill}</span>
                      <span className="opacity-70">{Math.round(skill.score)}/100 ({skill.level})</span>
                    </div>
                    <Progress value={skill.score} />
                  </div>
                ))}
                {(!profile?.skills || profile.skills.length === 0) && (
                  <p className="opacity-50 text-sm">Keep learning to unlock your skill assessments!</p>
                )}
              </div>
            </Card>

            {recommendation && recommendation.action !== "done" && (
              <div className="bg-gradient-to-br from-indigo-500/20 to-purple-500/20 border border-indigo-500/30 rounded-2xl p-6">
                <h2 className="text-lg font-bold mb-1">Recommended for you</h2>
                <h3 className="text-xl text-indigo-500 font-semibold mb-3">{recommendation.title}</h3>
                <p className="opacity-70 mb-4">{recommendation.description}</p>
                <div className="bg-black/20 p-3 rounded-xl border border-white/5 text-sm italic opacity-60 mb-4">
                  "{recommendation.ai_message}"
                </div>
                <Link 
                  href="/learn"
                  className="inline-block px-6 py-2.5 bg-indigo-500 hover:bg-indigo-600 text-white font-medium rounded-xl transition-colors"
                >
                  Start Practice
                </Link>
              </div>
            )}
          </div>
        );

      case "skills":
        return (
          <div className="space-y-6">
            <h2 className="text-xl font-bold mb-4">Skill Assessment Details</h2>
            <p className="opacity-50 mb-6">Your abilities are continuously updated from your learning activity.</p>
            
            {profile?.skills?.map(skill => (
              <div key={skill.skill} className="bg-white/5 border border-white/10 rounded-2xl p-6 group">
                <div className="flex justify-between items-start mb-4">
                  <div className="mb-4">
                    <h3 className="text-xl font-bold capitalize">{skill.skill}</h3>
                    <div className="flex items-center gap-2 mt-1">
                      <Badge variant={skill.level === 'strong' ? 'new' : skill.level === 'average' ? 'primary' : 'default'}>
                        {skill.level}
                      </Badge>
                      <span className="opacity-30">•</span>
                      <span className="opacity-60 text-sm flex items-center gap-1">
                        {skill.trend === 'improving' ? '↑ Improving' : skill.trend === 'declining' ? '↓ Declining' : '→ Stable'}
                      </span>
                    </div>
                  </div>
                </div>

                <div className="h-2 w-full bg-white/10 rounded-full overflow-hidden mb-6">
                  <div 
                    className={`h-full rounded-full ${skill.level === 'strong' ? 'bg-green-500' : skill.level === 'average' ? 'bg-yellow-500' : 'bg-red-500'}`}
                    style={{ width: `${skill.score}%` }}
                  ></div>
                </div>

                <div className="bg-black/20 rounded-xl p-4 border border-white/5">
                  <div className="flex justify-between items-center mb-1">
                    <span className="text-sm font-medium opacity-70">Assessment Confidence</span>
                    <span className="text-xs opacity-50">Based on {skill.evidence_count} activities</span>
                  </div>
                  <div className="h-1.5 w-full bg-white/10 rounded-full overflow-hidden">
                    <div 
                      className={`h-full rounded-full ${skill.confidence > 70 ? 'bg-indigo-400' : skill.confidence > 40 ? 'bg-indigo-400/60' : 'bg-indigo-400/30'}`}
                      style={{ width: `${Math.min(100, skill.confidence)}%` }}
                    ></div>
                  </div>
                  <details className="mt-4 text-sm opacity-50">
                    <summary className="cursor-pointer hover:opacity-100 transition-opacity">How is this calculated?</summary>
                    <div className="mt-2 p-3 bg-white/5 rounded-lg border border-white/5 text-xs space-y-2">
                      <p>Your score combines different time horizons:</p>
                      <ul className="list-disc pl-4 space-y-1">
                        <li><strong>Recent performance (60%):</strong> Your learning activity in the last 7 days.</li>
                        <li><strong>Retention (25%):</strong> Your ability to remember things after a few weeks.</li>
                        <li><strong>Long-term (15%):</strong> Historical stability over months.</li>
                      </ul>
                      <p className="pt-1">This ensures your score reflects both fast recent progress and solid long-term ability.</p>
                    </div>
                  </details>
                </div>
              </div>
            ))}
            
            {(!profile?.skills || profile.skills.length === 0) && (
              <div className="text-center p-10 bg-white/5 border border-white/10 rounded-2xl">
                <p className="opacity-50">You haven't completed enough learning activities to assess your skills yet.</p>
                <Link href="/learn" className="inline-block mt-4 text-indigo-400 hover:text-indigo-300">Start learning →</Link>
              </div>
            )}
          </div>
        );

      case "goals":
        const GOALS = ["Pass IELTS", "Pass TOEIC", "Improve English for university", "Speak more confidently", "Business English", "Travel"];
        const FOCUS_AREAS = ["Vocabulary", "Grammar", "Reading", "Listening", "Writing", "Speaking"];

        return (
          <div className="space-y-8">
            <div>
              <h2 className="text-xl font-bold mb-2">Learning Goals</h2>
              <p className="opacity-50 mb-4">Tell us what you want to achieve.</p>
              <div className="flex flex-wrap gap-3">
                {GOALS.map(goal => (
                  <button
                    key={goal}
                    onClick={() => toggleSelection(goal, selectedGoals, setSelectedGoals)}
                    className={`px-4 py-2 rounded-xl border text-sm transition-all ${
                      selectedGoals.includes(goal) 
                        ? "bg-indigo-500 border-indigo-400 text-white" 
                        : "bg-white/5 border-white/10 opacity-70 hover:opacity-100"
                    }`}
                  >
                    {goal}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <h2 className="text-xl font-bold mb-2">Focus Areas</h2>
              <p className="opacity-50 mb-4">What specific skills do you want to improve?</p>
              <div className="flex flex-wrap gap-3">
                {FOCUS_AREAS.map(area => (
                  <button
                    key={area}
                    onClick={() => toggleSelection(area, selectedFocus, setSelectedFocus)}
                    className={`px-4 py-2 rounded-xl border text-sm transition-all ${
                      selectedFocus.includes(area) 
                        ? "bg-purple-500 border-purple-400 text-white" 
                        : "bg-white/5 border-white/10 opacity-70 hover:opacity-100"
                    }`}
                  >
                    {area}
                  </button>
                ))}
              </div>
            </div>

            {message.text && (
              <div className={`p-4 rounded-xl text-sm ${message.type === 'success' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
                {message.text}
              </div>
            )}

            <Button
              onClick={handleSavePreferences}
              disabled={savingPreferences}
              variant="primary"
            >
              {savingPreferences ? "Saving..." : "Save Preferences"}
            </Button>
          </div>
        );

      case "assessment":
        return (
          <div className="space-y-6">
            <Card>
              <h2 className="text-xl font-bold mb-6">Periodic Assessment</h2>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
                <div className="bg-black/20 p-5 rounded-xl border border-white/5">
                  <div className="text-sm opacity-50 mb-1">Last Assessed</div>
                  <div className="text-lg font-medium">
                    {profile?.last_assessment_at ? new Date(profile.last_assessment_at).toLocaleDateString() : "Never"}
                  </div>
                </div>
                <div className="bg-black/20 p-5 rounded-xl border border-white/5">
                  <div className="text-sm opacity-50 mb-1">Next Reassessment</div>
                  <div className="text-lg font-medium">
                    {profile?.last_assessment_at 
                      ? "In " + Math.max(1, 12 - Math.floor((new Date().getTime() - new Date(profile.last_assessment_at).getTime()) / (1000 * 3600 * 24 * 7))) + " weeks"
                      : "Available now"}
                  </div>
                </div>
              </div>

              {(!profile?.last_assessment_at || (new Date().getTime() - new Date(profile.last_assessment_at).getTime()) / (1000 * 3600 * 24 * 7) >= 12) ? (
                <div className="p-5 bg-indigo-500/10 border border-indigo-500/20 rounded-xl flex flex-col md:flex-row items-center justify-between gap-4">
                  <div>
                    <h3 className="font-semibold text-indigo-300 mb-1">Time to check your progress</h3>
                    <p className="text-sm opacity-70">Take a short assessment test to recalibrate your learning baseline.</p>
                  </div>
                  <Link 
                    href="/placement"
                    className="shrink-0 px-6 py-2.5 bg-indigo-500 hover:bg-indigo-600 text-white font-medium rounded-xl transition-colors"
                  >
                    Start Assessment
                  </Link>
                </div>
              ) : (
                <div className="p-5 bg-green-500/10 border border-green-500/20 rounded-xl flex flex-col md:flex-row items-center justify-between gap-4">
                  <div>
                    <h3 className="font-semibold text-green-400 mb-1">Assessment up to date</h3>
                    <p className="text-sm opacity-70">Your skill profile is calibrated. Keep learning!</p>
                  </div>
                </div>
              )}
            </Card>
          </div>
        );

      case "account":
        return (
          <div className="space-y-12">
            {/* Bookmarks Section */}
            <div>
              <h2 className="text-xl font-bold mb-6">Saved Words ({bookmarks.length})</h2>
              {bookmarks.length === 0 ? (
                <div className="p-10 border border-white/10 rounded-2xl bg-white/5 text-center">
                  <p className="opacity-50 mb-4">You haven't saved any words yet.</p>
                  <Link href="/" className="inline-block px-6 py-2 bg-indigo-500 text-white rounded-xl">Explore Dictionary</Link>
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {bookmarks.map((b) => (
                    <div key={b.word} className="p-4 bg-white/5 border border-white/10 rounded-2xl flex justify-between items-center">
                      <Link href={`/dictionary/${b.word}`} className="font-medium hover:text-indigo-500">{b.word}</Link>
                      <button onClick={() => removeBookmark(b.word)} className="opacity-30 hover:text-red-500 hover:opacity-100">✕</button>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Sessions Section */}
            <div className="mt-8">
              <h2 className="text-xl font-bold mb-6">Active Sessions</h2>
              <div className="space-y-4">
                {sessions.map((sess) => (
                  <div key={sess.id} className="flex justify-between items-center p-4 bg-white/5 border border-white/10 rounded-2xl">
                    <div>
                      <div className="flex items-center gap-3">
                        <span className="font-medium">{sess.user_agent || "Unknown Device"}</span>
                        {sess.is_current && <span className="px-2 py-0.5 rounded text-xs bg-green-500/20 text-green-400">Current</span>}
                      </div>
                      <div className="text-sm opacity-50 mt-1">IP: {sess.ip_address} • Signed in: {new Date(sess.created_at).toLocaleDateString()}</div>
                    </div>
                    {!sess.is_current && (
                      <button onClick={() => handleRevokeSession(sess.id)} className="px-4 py-2 bg-red-500/10 text-red-400 rounded-xl text-sm">Revoke</button>
                    )}
                  </div>
                ))}
              </div>
            </div>
          </div>
        );
    }
  };

  const tabs: { id: Tab; label: string }[] = [
    { id: "overview", label: "Overview" },
    { id: "skills", label: "Skills" },
    { id: "goals", label: "Goals & Focus" },
    { id: "assessment", label: "Assessment" },
    { id: "account", label: "Account" },
  ];

  return (
    <div className="max-w-4xl mx-auto p-6 pt-12">
      <div className="flex flex-col md:flex-row md:items-center justify-between mb-8 gap-6">
        <div>
          <h1 className="text-3xl font-bold mb-2">Profile</h1>
          <p className="opacity-50">{user.email}</p>
        </div>
        <Button
          onClick={handleLogout}
          variant="outline"
          className="text-red-400 border-red-500/30 hover:bg-red-500/10"
        >
          Logout
        </Button>
      </div>

      <div className="flex border-b border-white/10 mb-8 overflow-x-auto">
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-6 py-3 font-medium text-sm whitespace-nowrap transition-colors border-b-2 ${
              activeTab === tab.id 
                ? "border-indigo-500 text-indigo-500" 
                : "border-transparent opacity-50 hover:opacity-80"
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      <div className="min-h-[400px]">
        {renderTabContent()}
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
