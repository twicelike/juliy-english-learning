import apiClient from "./api-client";
import { UserProfileDTO, RecommendationResponse } from "./profile-api";

export interface AchievementDef {
  id: string;
  title: string;
  description: string;
  icon: string;
}

export interface ChartDataPoint {
  date: string;
  xp: number;
  words: number;
  active: boolean;
}

export interface DailyProgressDTO {
  xp: number;
  reviews_completed: number;
  new_words_learned: number;
}

export interface DashboardStatistics {
  words_learned: number;
  reviews_due: number;
  accuracy: number;
}

export interface DashboardGamification {
  level: number;
  current_xp: number;
  next_level_xp: number;
  streak: number;
  daily_goal_met: boolean;
  today_progress: number;
  daily_goal: number;
  achievements: AchievementDef[];
}

export interface UnifiedDashboardResponse {
  profile: UserProfileDTO;
  statistics: DashboardStatistics;
  charts: ChartDataPoint[];
  gamification: DashboardGamification;
  recommendation: RecommendationResponse;
}

export const DashboardAPI = {
  getUnifiedDashboard: () =>
    apiClient.get<{ data: UnifiedDashboardResponse }>("/dashboard").then((r) => r.data.data),

  getChart: (period: number = 7) =>
    apiClient.get<{ data: ChartDataPoint[] }>(`/gamification/chart?period=${period}`).then((r) => r.data.data),
};
