import { useState, useRef, useEffect, useCallback } from "react";

export type RecordingStatus =
  | "idle"
  | "requesting_permission"
  | "recording"
  | "paused"
  | "stopped"
  | "error";

interface UseAudioRecorderOptions {
  onVolumeChange?: (volume: number) => void;
}

export function useAudioRecorder(options?: UseAudioRecorderOptions) {
  const [status, setStatus] = useState<RecordingStatus>("idle");
  const [error, setError] = useState<string | null>(null);
  const [duration, setDuration] = useState(0); // in seconds
  const [audioBlob, setAudioBlob] = useState<Blob | null>(null);

  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const audioContextRef = useRef<AudioContext | null>(null);
  const analyserRef = useRef<AnalyserNode | null>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const animationFrameRef = useRef<number | null>(null);
  
  // Timer tracking
  const timerIntervalRef = useRef<NodeJS.Timeout | null>(null);
  
  // Clean up
  const cleanup = useCallback(() => {
    if (animationFrameRef.current) cancelAnimationFrame(animationFrameRef.current);
    if (timerIntervalRef.current) clearInterval(timerIntervalRef.current);
    
    if (mediaRecorderRef.current && mediaRecorderRef.current.state !== "inactive") {
      mediaRecorderRef.current.stop();
    }
    
    if (streamRef.current) {
      streamRef.current.getTracks().forEach((track) => track.stop());
    }
    
    if (audioContextRef.current && audioContextRef.current.state !== "closed") {
      audioContextRef.current.close();
    }

    mediaRecorderRef.current = null;
    audioContextRef.current = null;
    analyserRef.current = null;
    streamRef.current = null;
  }, []);

  useEffect(() => {
    return cleanup;
  }, [cleanup]);

  const updateVolume = useCallback(() => {
    if (!analyserRef.current || status !== "recording") return;

    const dataArray = new Uint8Array(analyserRef.current.frequencyBinCount);
    analyserRef.current.getByteFrequencyData(dataArray);

    let sum = 0;
    for (let i = 0; i < dataArray.length; i++) {
      sum += dataArray[i];
    }
    const average = sum / dataArray.length;
    // Normalize volume between 0 and 1
    const normalizedVolume = Math.min(average / 128, 1);
    
    if (options?.onVolumeChange) {
      options.onVolumeChange(normalizedVolume);
    }

    animationFrameRef.current = requestAnimationFrame(updateVolume);
  }, [status, options]);

  // Start timer interval
  const startTimer = useCallback(() => {
    if (timerIntervalRef.current) clearInterval(timerIntervalRef.current);
    timerIntervalRef.current = setInterval(() => {
      setDuration((prev) => prev + 1);
    }, 1000);
  }, []);

  // Stop timer interval
  const stopTimer = useCallback(() => {
    if (timerIntervalRef.current) clearInterval(timerIntervalRef.current);
  }, []);

  const start = async () => {
    try {
      setStatus("requesting_permission");
      setError(null);
      setDuration(0);
      setAudioBlob(null);

      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      streamRef.current = stream;

      // Audio Context for Volume Meter
      const audioContext = new (window.AudioContext || (window as any).webkitAudioContext)();
      const analyser = audioContext.createAnalyser();
      analyser.fftSize = 256;
      const source = audioContext.createMediaStreamSource(stream);
      source.connect(analyser);
      
      audioContextRef.current = audioContext;
      analyserRef.current = analyser;

      // MediaRecorder setup with fallback types
      const mimeTypes = [
        "audio/webm;codecs=opus",
        "audio/webm",
        "audio/mp4",
        "audio/ogg",
      ];
      
      let selectedMimeType = "";
      for (const type of mimeTypes) {
        if (MediaRecorder.isTypeSupported(type)) {
          selectedMimeType = type;
          break;
        }
      }

      const options = selectedMimeType ? { mimeType: selectedMimeType } : undefined;
      const mediaRecorder = new MediaRecorder(stream, options);
      mediaRecorderRef.current = mediaRecorder;

      const audioChunks: BlobPart[] = [];
      mediaRecorder.ondataavailable = (e) => {
        if (e.data.size > 0) {
          audioChunks.push(e.data);
        }
      };

      mediaRecorder.onstop = () => {
        const audioBlob = new Blob(audioChunks, { type: selectedMimeType || "audio/webm" });
        setAudioBlob(audioBlob);
      };

      mediaRecorder.start();
      setStatus("recording");
      startTimer();
      updateVolume();

    } catch (err: any) {
      console.error("Failed to start recording", err);
      setStatus("error");
      if (err.name === "NotAllowedError" || err.name === "PermissionDeniedError") {
        setError("Microphone access is required to record your speaking practice.");
      } else if (err.name === "NotFoundError" || err.name === "DevicesNotFoundError") {
        setError("No microphone detected on your device.");
      } else {
        setError("Audio recording isn't supported or failed to start.");
      }
      cleanup();
    }
  };

  const pause = () => {
    if (mediaRecorderRef.current && mediaRecorderRef.current.state === "recording") {
      mediaRecorderRef.current.pause();
      setStatus("paused");
      stopTimer();
      if (animationFrameRef.current) cancelAnimationFrame(animationFrameRef.current);
    }
  };

  const resume = () => {
    if (mediaRecorderRef.current && mediaRecorderRef.current.state === "paused") {
      mediaRecorderRef.current.resume();
      setStatus("recording");
      startTimer();
      updateVolume();
    }
  };

  const stop = () => {
    if (mediaRecorderRef.current && mediaRecorderRef.current.state !== "inactive") {
      mediaRecorderRef.current.stop();
      setStatus("stopped");
      stopTimer();
      cleanup();
    }
  };

  const reset = () => {
    cleanup();
    setStatus("idle");
    setDuration(0);
    setAudioBlob(null);
    setError(null);
  };

  // Ensure volume updates start/stop correctly with status changes
  useEffect(() => {
    if (status === "recording") {
      updateVolume();
    } else if (animationFrameRef.current) {
      cancelAnimationFrame(animationFrameRef.current);
    }
  }, [status, updateVolume]);

  return {
    status,
    duration,
    audioBlob,
    error,
    start,
    pause,
    resume,
    stop,
    reset,
  };
}
