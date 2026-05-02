import { useEffect, useState } from "react";
import { EventsOn, EventsOff } from "../../wailsjs/runtime/runtime";

interface DownloadProgress {
  downloaded: number;
  total: number;
  percent: number;
}

export function useOllamaDownload() {
  const [progress, setProgress] = useState<DownloadProgress | null>(null);
  const [isDownloading, setIsDownloading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    EventsOn("ollama:download:progress", (data: DownloadProgress) => {
      setProgress(data);
      setIsDownloading(true);
    });

    EventsOn("ollama:download:complete", () => {
      setIsDownloading(false);
      setProgress(null);
    });

    EventsOn("ollama:download:error", (err: string) => {
      setIsDownloading(false);
      setError(err);
      console.error("Download error:", err);
    });

    return () => {
      EventsOff("ollama:download:progress");
      EventsOff("ollama:download:complete");
      EventsOff("ollama:download:error");
    };
  }, []);

  return { progress, isDownloading, error };
}