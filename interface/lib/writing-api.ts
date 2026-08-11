import apiClient from "./api-client";

export interface DashboardStatsDTO {
  total_completed: number;
  completed_easy: number;
  completed_medium: number;
  completed_hard: number;
  average_score: number;
  mastery_mastered: number;
  mastery_familiar: number;
  mastery_beginner: number;
  recent_submissions: MissionHistoryDTO[];
}

export interface MissionHistoryDTO {
  mission_id: string;
  topic: string;
  scenario: string;
  prompt_vi: string[];
  user_content: string;
  corrected_version: string;
  detailed_feedback: string;
  score: number;
  completed_at: string;
}

export interface MissionDTO {
  id: string;
  topic: string;
  scenario: string;
  difficulty: string;
  prompt_vi: string[];
  target_vocabulary: string[];
  grammar_focus: string[];
  estimated_time: number;
}

export interface CategoryScore {
  score: number;
  feedback: string;
}

export interface EvaluationBreakdown {
  grammar: CategoryScore;
  vocabulary: CategoryScore;
  naturalness: CategoryScore;
  completeness: CategoryScore;
}

export interface EvaluationResponseDTO {
  score: number;
  breakdown: EvaluationBreakdown;
  corrected_version: string;
  detailed_feedback: string;
}

export const WritingAPI = {
  getDashboardStats: async (): Promise<DashboardStatsDTO> => {
    const res = await apiClient.get<{ data: DashboardStatsDTO }>("/writing/dashboard");
    return res.data.data;
  },

  getNewMission: async (topic?: string, difficulty?: string): Promise<MissionDTO> => {
    const params: any = {};
    if (topic) params.topic = topic;
    if (difficulty) params.difficulty = difficulty;
    
    const res = await apiClient.get<{ data: MissionDTO }>("/writing/missions/new", { params });
    return res.data.data;
  },

  getReviewMission: async (topic?: string, difficulty?: string): Promise<MissionDTO> => {
    const params: any = {};
    if (topic) params.topic = topic;
    if (difficulty) params.difficulty = difficulty;

    const res = await apiClient.get<{ data: MissionDTO }>("/writing/missions/review", { params });
    return res.data.data;
  },

  submitMission: async (missionId: string, content: string): Promise<EvaluationResponseDTO> => {
    const res = await apiClient.post<{ data: EvaluationResponseDTO }>(`/writing/missions/${missionId}/submit`, {
      content,
    });
    return res.data.data;
  },
};
