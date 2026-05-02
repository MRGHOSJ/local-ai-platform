import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import {
  Cpu,
  HardDrive,
  Zap,
  Play,
  Download,
  Bot,
  CheckCircle,
  XCircle,
} from "lucide-react";
import {
  getCachedHardware,
  formatBytes,
  ensureOllama,
  ollamaStart,
  ollamaStatus,
  getCacheSize,
  getCurrentModel,
  type SystemSpecs,
} from "@/lib/hardware";

export default function OverviewPage() {
  const [specs, setSpecs] = useState<SystemSpecs | null>(null);
  const [loading, setLoading] = useState(true);
  const [isOllamaRunning, setIsOllamaRunning] = useState(false);
  const [currentModel, setCurrentModel] = useState<string>("");
  const [cacheSize, setCacheSize] = useState<number>(0);
  const [starting, setStarting] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [hardwareData, running, model, size] = await Promise.all([
        getCachedHardware(),
        ollamaStatus(),
        getCurrentModel(),
        getCacheSize(),
      ]);
      setSpecs(hardwareData);
      setIsOllamaRunning(running);
      setCurrentModel(model || "");
      setCacheSize(size);
    } catch (err) {
      console.error("Failed to load overview data:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleStartOllama = async () => {
    setStarting(true);
    try {
      await ensureOllama();
      await ollamaStart();
      await loadData();
    } catch (err) {
      console.error("Failed to start Ollama:", err);
    } finally {
      setStarting(false);
    }
  };

  if (loading) {
    return (
      <div className="p-6 flex items-center justify-center h-full">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    );
  }

  const isHealthy = specs ? specs.score >= 40 : false;

  return (
    <div className="p-6 space-y-6 max-w-5xl mx-auto">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Overview</h1>
        <p className="text-muted-foreground">Your local AI agent platform</p>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        {/* System Status Card */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base flex items-center gap-2">
              <Cpu className="h-4 w-4" />
              System Status
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-4xl font-bold">{specs?.score || 0}</span>
              <Badge
                variant={isHealthy ? "default" : "destructive"}
                className={isHealthy ? "bg-green-500" : ""}
              >
                {isHealthy ? "Healthy" : "Limited"}
              </Badge>
            </div>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">CPU</span>
                <span className="font-medium">{specs?.cpu.cores || 0} cores</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">RAM</span>
                <span className="font-medium">
                  {specs ? formatBytes(specs.memory.totalBytes) : "0 B"}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">GPU</span>
                <span className="font-medium">
                  {specs?.gpus?.[0]?.model || "None"}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Quick Actions Card */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base flex items-center gap-2">
              <Zap className="h-4 w-4" />
              Quick Actions
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {isOllamaRunning ? (
              <div className="flex items-center gap-2 text-green-500 text-sm">
                <CheckCircle className="h-4 w-4" />
                Ollama Running
              </div>
            ) : (
              <Button
                className="w-full"
                onClick={handleStartOllama}
                disabled={starting}
              >
                <Play className="mr-2 h-4 w-4" />
                {starting ? "Starting..." : "Start Ollama"}
              </Button>
            )}
            <Button
              className="w-full"
              variant="outline"
              onClick={() => window.location.hash = "/models"}
              disabled={!isOllamaRunning}
            >
              <Download className="mr-2 h-4 w-4" />
              Download Model
            </Button>
            <Button
              className="w-full"
              variant="outline"
              disabled
            >
              <Bot className="mr-2 h-4 w-4" />
              Create Agent
            </Button>
          </CardContent>
        </Card>

        {/* Active Model Card */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base flex items-center gap-2">
              <HardDrive className="h-4 w-4" />
              Active Model
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {currentModel ? (
              <div>
                <p className="font-mono font-medium text-lg">{currentModel}</p>
                <p className="text-xs text-muted-foreground mt-1">
                  Cache: {formatBytes(cacheSize)}
                </p>
              </div>
            ) : (
              <div className="flex items-center gap-2 text-muted-foreground text-sm">
                <XCircle className="h-4 w-4" />
                No model loaded
              </div>
            )}
            <Progress value={isOllamaRunning ? 100 : 0} className="h-2" />
            <p className="text-xs text-muted-foreground">
              {isOllamaRunning ? "Ready for inference" : "Start Ollama to enable"}
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}