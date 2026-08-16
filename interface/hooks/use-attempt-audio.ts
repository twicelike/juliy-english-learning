import { useState, useEffect } from "react";

const isServer = typeof window === "undefined";
const baseURL = isServer
  ? process.env.NEXT_SERVER_API_URL || "http://127.0.0.1:8080/api/v1"
  : process.env.NEXT_PUBLIC_API_URL || "/api/v1";

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
        const response = await fetch(`${baseURL}/speaking/attempts/${attemptId}/audio`, {
          credentials: "include",
        });
        
        if (!active) return;
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const blob = await response.blob();
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
