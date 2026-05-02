import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Separator } from "@/components/ui/separator";
import { Cpu, HardDrive, MemoryStick, Monitor, RefreshCw } from "lucide-react";
import {
  getCachedHardware,
  clearHardwareCache,
  formatBytes,
  formatMHz,
  type SystemSpecs,
} from "@/lib/hardware";

async function loadHardware(fresh: boolean): Promise<SystemSpecs | null> {
  try {
    if (fresh) {
      await clearHardwareCache();
    }
    return await getCachedHardware();
  } catch (err) {
    console.error("Hardware detection failed:", err);
    return null;
  }
}

export default function SystemPage() {
  const [specs, setSpecs] = useState<SystemSpecs | null>(null);
  const [loading, setLoading] = useState(true);
  const [scanning, setScanning] = useState(false);

  useEffect(() => {
    loadHardware(false).then((data) => {
      setSpecs(data);
      setLoading(false);
    });
  }, []);

  const handleRefresh = async () => {
    setScanning(true);
    const data = await loadHardware(true);
    setSpecs(data);
    setScanning(false);
  };

  if (loading) {
    return <SystemInfoSkeleton />;
  }

  if (!specs) {
    return (
      <div className="flex flex-col items-center justify-center h-96 gap-4">
        <p className="text-muted-foreground">Failed to detect hardware.</p>
        <Button onClick={handleRefresh} variant="outline">
          <RefreshCw className="mr-2 h-4 w-4" />
          Retry
        </Button>
      </div>
    );
  }

  const ramUsedPercent = specs.memory.usedPercent;

  return (
    <div className="space-y-6 max-w-4xl mx-auto p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">System Information</h1>
          <p className="text-muted-foreground text-sm">
            Last scanned: {new Date(specs.detectedAt).toLocaleString()}
          </p>
        </div>
        <Button onClick={handleRefresh} disabled={scanning} variant="outline">
          <RefreshCw className={`mr-2 h-4 w-4 ${scanning ? "animate-spin" : ""}`} />
          {scanning ? "Scanning..." : "Rescan"}
        </Button>
      </div>

      <Card className="border-primary/20 bg-primary/5">
        <CardContent className="pt-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Compatibility Score</p>
              <p className="text-4xl font-bold text-primary mt-1">{specs.score}/100</p>
            </div>
            <div className="text-right">
              <CompatibilityBadge score={specs.score} />
              <p className="text-xs text-muted-foreground mt-1">
                {getRecommendation(specs.score)}
              </p>
            </div>
          </div>
          <Progress value={specs.score} className="mt-4 h-2" />
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center gap-2 pb-2">
            <Cpu className="h-5 w-5 text-primary" />
            <CardTitle className="text-base">CPU</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <p className="font-medium">{specs.cpu.model}</p>
            <div className="flex gap-2 flex-wrap">
              <Badge variant="secondary">{specs.cpu.cores} Cores</Badge>
              <Badge variant="secondary">{specs.cpu.threads} Threads</Badge>
              <Badge variant="secondary">{formatMHz(specs.cpu.frequencyMHz)}</Badge>
              <Badge variant="secondary">{specs.cpu.architecture}</Badge>
            </div>
            <div className="flex gap-1 flex-wrap mt-2">
              {(specs.cpu.features || []).map((f) => (
                <Badge key={f} variant="outline" className="text-xs">
                  {f.toUpperCase()}
                </Badge>
              ))}
              {(specs.cpu.features || []).length === 0 && (
                <span className="text-xs text-muted-foreground">No SIMD features detected</span>
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center gap-2 pb-2">
            <MemoryStick className="h-5 w-5 text-primary" />
            <CardTitle className="text-base">Memory</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Total</span>
              <span className="font-medium">{formatBytes(specs.memory.totalBytes)}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Used</span>
              <span className="font-medium">{formatBytes(specs.memory.usedBytes)}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Free</span>
              <span className="font-medium">{formatBytes(specs.memory.freeBytes)}</span>
            </div>
            <Progress value={ramUsedPercent} className="h-2" />
            <p className="text-xs text-muted-foreground text-right">{ramUsedPercent.toFixed(1)}% used</p>
          </CardContent>
        </Card>

        <Card className="md:col-span-2">
          <CardHeader className="flex flex-row items-center gap-2 pb-2">
            <Monitor className="h-5 w-5 text-primary" />
            <CardTitle className="text-base">Graphics</CardTitle>
          </CardHeader>
          <CardContent>
            {(specs.gpus || []).length === 0 ? (
              <p className="text-sm text-muted-foreground">No GPU detected. CPU inference only.</p>
            ) : (
              <div className="space-y-4">
                {(specs.gpus || []).map((gpu) => (
                  <div key={gpu.index} className="flex items-start justify-between p-3 rounded-lg border bg-card/50">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <p className="font-medium">{gpu.model}</p>
                        <BackendBadge backend={gpu.backend} />
                      </div>
                      <div className="flex gap-2 text-xs text-muted-foreground">
                        <span>{gpu.vendor}</span>
                        {gpu.vramBytes > 0 && (
                          <>
                            <Separator orientation="vertical" className="h-3" />
                            <span>{formatBytes(gpu.vramBytes)} VRAM</span>
                          </>
                        )}
                        {gpu.driver && (
                          <>
                            <Separator orientation="vertical" className="h-3" />
                            <span>{gpu.driver}</span>
                          </>
                        )}
                      </div>
                    </div>
                    <Badge variant={gpu.isSupported ? "default" : "destructive"}>
                      {gpu.isSupported ? "Supported" : "Unsupported"}
                    </Badge>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="md:col-span-2">
          <CardHeader className="flex flex-row items-center gap-2 pb-2">
            <HardDrive className="h-5 w-5 text-primary" />
            <CardTitle className="text-base">Storage</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Path</span>
              <span className="font-mono text-xs">{specs.storage.path}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Total</span>
              <span className="font-medium">{formatBytes(specs.storage.totalBytes)}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Free</span>
              <span className="font-medium">{formatBytes(specs.storage.freeBytes)}</span>
            </div>
            <Progress value={specs.storage.usedPercent} className="h-2" />
            <p className="text-xs text-muted-foreground text-right">
              {specs.storage.usedPercent.toFixed(1)}% used
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function CompatibilityBadge({ score }: { score: number }) {
  if (score >= 80) return <Badge className="bg-green-500 text-white">Excellent</Badge>;
  if (score >= 60) return <Badge className="bg-yellow-500 text-white">Good</Badge>;
  if (score >= 40) return <Badge variant="secondary">Fair</Badge>;
  return <Badge variant="destructive">Limited</Badge>;
}

function BackendBadge({ backend }: { backend: string }) {
  const colors: Record<string, string> = {
    cuda: "bg-green-500/10 text-green-500 border-green-500/20",
    rocm: "bg-red-500/10 text-red-500 border-red-500/20",
    metal: "bg-blue-500/10 text-blue-500 border-blue-500/20",
  };
  return (
    <Badge variant="outline" className={colors[backend] || "text-muted-foreground"}>
      {backend.toUpperCase()}
    </Badge>
  );
}

function getRecommendation(score: number): string {
  if (score >= 80) return "Can run 70B models with distributed compute";
  if (score >= 60) return "Can run 13B-30B models locally";
  if (score >= 40) return "Can run 7B-8B models locally";
  if (score >= 20) return "Can run small 3B models";
  return "CPU-only inference recommended";
}

function SystemInfoSkeleton() {
  return (
    <div className="space-y-6 max-w-4xl mx-auto p-6">
      <Skeleton className="h-8 w-64" />
      <Skeleton className="h-32 w-full" />
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Skeleton className="h-48" />
        <Skeleton className="h-48" />
        <Skeleton className="h-48 md:col-span-2" />
      </div>
    </div>
  );
}