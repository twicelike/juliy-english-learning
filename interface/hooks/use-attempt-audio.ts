import { useState, useEffect } from "react";
import apiClient from "@/lib/api-client";

export function useAttemptAudio(attemptId: string | null) {
  const [audioUrl, setAudioUrl] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let objectUrl: string | null = null;
    let active = true;

    const fetchAudio = async () => {
      if (!attemptId) return;
      
      setIsLoading(true);
      setError(null);
      
      try {
        const response = await apiClient.get(`/speaking/attempts/${attemptId}/audio`, {
          responseType: "blob",
        });
        
        if (!active) return;
        
        const blob = response.data;
        objectUrl = URL.createObjectURL(blob);
        setAudioUrl(objectUrl);
      } catch (err: any) {
        if (!active) return;
        console.error("Failed to load audio:", err);
        setError("Failed to load recording audio");
      } finally {
        if (active) {
          setIsLoading(false);
        }
      }
    };

    fetchAudio();

    return () => {
      active = false;
      if (objectUrl) {
        URL.revokeObjectURL(objectUrl);
      }
    };
  }, [attemptId]);

  return { audioUrl, isLoading, error };
}
