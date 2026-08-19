import apiClient from "./api-client";

export interface SkillAssessmentDTO {
  skill: string;
  score: number;
  level: "weak" | "average" | "strong";
  confidence: number;
  evidence_count: number;
  last_evaluated_at: string;
  trend: "improving" | "stable" | "declining" | "insufficient_data";
}

export interface UserProfileDTO {
  current_cefr?: string;
  daily_goal: number;
  completed_placement: boolean;
  strengths: string[];
  weaknesses: string[];
  skills: SkillAssessmentDTO[];
  goals: string[];
  focus_areas: string[];
  last_assessment_at?: string;
}

export type RecommendationAction = "review" | "learn_new" | "grammar" | "fill_blank" | "done";

export interface RecommendationResponse {
  action: RecommendationAction;
  title: string;
  description: string;
  count: number;
  ai_message: string;
}

export interface DashboardResponse {
  profile: UserProfileDTO;
  recommendation: RecommendationResponse;
  statistics?: any;
  gamification?: any;
}

export type QuestionType = "synonym" | "fill_blank" | "grammar";

export interface PlacementQuestionDTO {
  id: string;
  type: QuestionType;
  question: string;
  options: string[];
}

export interface PlacementTestResponse {
  questions: PlacementQuestionDTO[];
}

export interface SubmitTestRequest {
  answers: Record<string, string>;
}

export interface SubmitTestResponse {
  score: number;
  cefr: string;
}

export interface UpdatePreferencesRequest {
  goals: string[];
  focus_areas: string[];
}

export const ProfileAPI = {
  getDashboard: () =>
    apiClient.get<{ data: DashboardResponse }>("/profile").then((r) => r.data.data),

  getPlacementTest: () =>
    apiClient.get<{ data: PlacementTestResponse }>("/profile/placement-test").then((r) => r.data.data),

  submitPlacementTest: (req: SubmitTestRequest) =>
    apiClient.post<{ data: SubmitTestResponse }>("/profile/placement-test", req).then((r) => r.data.data),

  updatePreferences: (req: UpdatePreferencesRequest) =>
    apiClient.patch<{ success: boolean }>("/profile/preferences", req).then((r) => r.data),
};
