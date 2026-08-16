import apiClient from "@/lib/api-client";

export type SpeakingTaskType = "quick_practice" | "scenario" | "free_talk";

export interface GetRandomTaskParams {
  topic?: string;
  type: SpeakingTaskType;
  difficulty?: string;
}

export interface SpeakingTask {
  id: string;
  type: SpeakingTaskType;
  topic: string;
  title: string;
  prompt: string;
  difficulty?: string;
  scenario?: string;
  instructions: string[];
  hints?: string[];
  suggested_vocabulary?: string[];
  duration_seconds: number;
}

export interface SpeakingTaskSnapshot {
  topic: string;
  task_type: SpeakingTaskType;
  difficulty?: string;
  prompt: string;
  scenario?: string;
  instructions: string[];
  hints?: string[];
  suggested_vocabulary?: string[];
  duration_seconds: number;
}

export interface SpeakingSession {
  id: string;
  task_id: string;
  task_snapshot: SpeakingTaskSnapshot;
  status: "in_progress" | "completed";
  started_at: string;
}

export interface SpeakingSessionDetails {
  session: SpeakingSession;
  attempts: SpeakingAttempt[];
}

export interface SpeakingHistoryItem {
  session: SpeakingSession;
  best_score: number | null;
  attempts_count: number;
}

export interface SpeakingHistoryResponse {
  items: SpeakingHistoryItem[];
  total: number;
}

export interface NewUsedWord {
  word: string;
  word_id?: string;
  reason: string;
}

export interface SpeakingEvaluation {
  overall: number;
  task_completion: number;
  fluency: number;
  grammar: number;
  vocabulary: number;
  pronunciation: number;
  strengths: string[];
  improvements: string[];
  new_used_vocabulary?: NewUsedWord[];
}

export interface SpeakingAttempt {
  id: string;
  session_id: string;
  status: "processing" | "completed" | "failed";
  transcript?: string;
  evaluation?: SpeakingEvaluation;
  error_code?: string;
  error_message?: string;
}

export const MOCK_RECENT_PRACTICES = [
  {
    id: "hist-1",
    title: "Travel Experience",
    type: "Quick Practice",
    score: 84,
    practicedAt: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(), // 1 day ago
  },
  {
    id: "hist-2",
    title: "Restaurant Complaint",
    type: "Scenario",
    score: 78,
    practicedAt: new Date(Date.now() - 1000 * 60 * 60 * 48).toISOString(), // 2 days ago
  },
];

export interface RecommendationReason {
  skill: string;
  focus: string;
  based_on_sessions: number;
}

export interface RecommendedTaskResponse {
  task: SpeakingTask;
  reason?: RecommendationReason;
}

export const SpeakingAPI = {
  /**
   * Fetch recommended task based on performance.
   */
  async getRecommendedTask(): Promise<RecommendedTaskResponse> {
    const res = await apiClient.get<RecommendedTaskResponse>(`/speaking/tasks/recommended`);
    return res.data;
  },

  /**
   * Fetch a random task from the backend.
   * If DB misses, backend AI will generate it.
   */
  async getRandomTask(params: GetRandomTaskParams): Promise<SpeakingTask> {
    const searchParams = new URLSearchParams();
    if (params.topic) searchParams.set("topic", params.topic);
    searchParams.set("type", params.type);
    if (params.difficulty) searchParams.set("difficulty", params.difficulty);

    const res = await apiClient.get<SpeakingTask>(
      `/speaking/tasks/random?${searchParams.toString()}`
    );
    return res.data; // Ensure apiClient returns { data: SpeakingTask } or just SpeakingTask based on setup
  },

  /**
   * Start a new speaking session for the given task.
   */
  async startSession(taskId: string): Promise<SpeakingSession> {
    const res = await apiClient.post<SpeakingSession>(`/speaking/sessions`, {
      task_id: taskId,
    });
    return res.data;
  },

  /**
   * Get session details including task snapshot.
   */
  async getSessionDetails(sessionId: string): Promise<SpeakingSessionDetails> {
    const res = await apiClient.get<{ data: SpeakingSessionDetails }>(`/speaking/sessions/${sessionId}`);
    // Backend handler returns c.JSON(http.StatusOK, gin.H{"session": session, "attempts": attempts})
    // Oh wait, in handler.go:
    // c.JSON(http.StatusOK, gin.H{ "session": session, "attempts": attempts })
    // So the response is directly the object, not wrapped in data.
    return res.data as unknown as SpeakingSessionDetails;
  },

  /**
   * Submit an audio attempt for a session
   */
  async submitAttempt(sessionId: string, audioBlob: Blob, durationMs: number): Promise<{attempt_id: string, status: string}> {
    const formData = new FormData();
    formData.append("audio", audioBlob, "recording.webm");
    formData.append("duration_ms", durationMs.toString());

    const res = await apiClient.post<{attempt_id: string, status: string}>(
      `/speaking/sessions/${sessionId}/attempts`,
      formData,
      { headers: { "Content-Type": "multipart/form-data" } }
    );
    return res.data;
  },

  /**
   * Poll for attempt details (processing, completed, failed)
   */
  async getAttempt(attemptId: string): Promise<SpeakingAttempt> {
    const res = await apiClient.get<SpeakingAttempt>(`/speaking/attempts/${attemptId}`);
    return res.data;
  },

  /**
   * Get recent practices for the home page (takes up to 2 items from history).
   */
  async getRecentPractices() {
    try {
      const historyItems = await this.getHistory();
      if (!historyItems || !Array.isArray(historyItems)) return [];
      
      // Take up to 2 items and map them to RecentPracticeItem format
      return historyItems.slice(0, 2).map((item) => ({
        id: item.session.id,
        title: item.session.task_snapshot?.title || item.session.task_snapshot?.topic || "Speaking Practice",
        type: item.session.task_snapshot?.type || "Quick Practice",
        score: item.best_score || 0,
        practicedAt: item.session.started_at,
      }));
    } catch (err) {
      console.error("Failed to fetch recent practices:", err);
      return [];
    }
  },

  /**
   * Get speaking session history.
   */
  async getHistory(): Promise<SpeakingHistoryResponse> {
    const res = await apiClient.get<SpeakingHistoryResponse>(`/speaking/sessions`);
    return res.data;
  },
};
