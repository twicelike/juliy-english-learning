import apiClient from "./api-client";

export interface UserProfileDTO {
  current_cefr: string;
  daily_goal: number;
  completed_placement: boolean;
  strengths: string[];
  weaknesses: string[];
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
  today_progress: number;
  streak: number;
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

export const ProfileAPI = {
  getDashboard: () =>
    apiClient.get<{ data: DashboardResponse }>("/profile").then((r) => r.data.data),

  getPlacementTest: () =>
    apiClient.get<{ data: PlacementTestResponse }>("/profile/placement-test").then((r) => r.data.data),

  submitPlacementTest: (req: SubmitTestRequest) =>
    apiClient.post<{ data: SubmitTestResponse }>("/profile/placement-test", req).then((r) => r.data.data),
};
