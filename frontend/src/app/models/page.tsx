import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Input } from "@/components/ui/input";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Download,
  Trash2,
  HardDrive,
  Clock,
  Search,
  Cpu,
  CheckCircle2,
  Loader2,
  ExternalLink,
} from "lucide-react";
import {
  listModels,
  pullModel,
  deleteModel,
  getModelRecommendations,
  searchHFModels,
  getPopularHFModels,
  formatBytes,
  type ModelInfo,
  type ModelRecommendation,
  type CatalogModel,
} from "@/lib/hardware";

export default function ModelsPage() {
  const [installed, setInstalled] = useState<ModelInfo[]>([]);
  const [recommendations, setRecommendations] = useState<ModelRecommendation[]>([]);
  const [hfCatalog, setHFCatalog] = useState<CatalogModel[]>([]);
  const [pulling, setPulling] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [source, setSource] = useState<"ollama" | "huggingface">("ollama");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadAll();
  }, []);

  const loadAll = async () => {
    setLoading(true);
    try {
      const [models, recs] = await Promise.all([
        listModels(),
        getModelRecommendations(),
      ]);
      setInstalled(models || []);
      setRecommendations(recs || []);
    } catch (e) {
      console.error("Failed to load:", e);
    } finally {
      setLoading(false);
    }
  };

  const loadHFCatalog = async (query: string) => {
    setLoading(true);
    try {
      const data = query 
        ? await searchHFModels(query)
        : await getPopularHFModels();
      setHFCatalog(data || []);
    } catch (e) {
      console.error("Failed to load HF catalog:", e);
    } finally {
      setLoading(false);
    }
  };

  const handlePull = async (name: string) => {
    setPulling(name);
    try {
      await pullModel(name);
      await loadAll();
    } catch (e) {
      console.error("Pull failed:", e);
    } finally {
      setPulling(null);
    }
  };

  const handleDelete = async (name: string) => {
    try {
      await deleteModel(name);
      await loadAll();
    } catch (e) {
      console.error("Delete failed:", e);
    }
  };

  const handleSourceChange = (newSource: "ollama" | "huggingface") => {
    setSource(newSource);
    setSearchQuery("");
    if (newSource === "huggingface") {
      loadHFCatalog("");
    }
  };

  const handleSearch = (query: string) => {
    setSearchQuery(query);
    if (source === "huggingface") {
      loadHFCatalog(query);
    }
  };

  const isInstalled = (name: string) => {
    const baseName = name.split(":")[0].split("-")[0];
    return installed.some(m => m.name.startsWith(baseName));
  };

  const getHFModelURL = (id: string) => {
    const hfId = id.replace("hf--", "").replace(/--/g, "/");
    return `https://huggingface.co/${hfId}`;
  };

  return (
    <div className="h-full flex flex-col p-6 gap-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Model Library</h1>
          <p className="text-muted-foreground text-sm">
            {source === "ollama" 
              ? `${recommendations.length} curated models from Ollama registry`
              : `${hfCatalog.length} models from Hugging Face Hub`
            }
          </p>
        </div>
        <Badge variant="outline" className="gap-1">
          <HardDrive className="h-3 w-3" />
          {installed.length} installed
        </Badge>
      </div>

      <Tabs value={source} onValueChange={(v) => handleSourceChange(v as "ollama" | "huggingface")}>
        <TabsList className="grid w-full max-w-md grid-cols-2">
          <TabsTrigger value="ollama" className="gap-2">
            <CheckCircle2 className="h-4 w-4" />
            Ollama Registry
          </TabsTrigger>
          <TabsTrigger value="huggingface" className="gap-2">
            <ExternalLink className="h-4 w-4" />
            Hugging Face
          </TabsTrigger>
        </TabsList>

        <div className="relative max-w-md mt-4">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder={source === "ollama" 
              ? "Search Ollama models..." 
              : "Search Hugging Face (e.g., unsloth, qwen)..."
            }
            className="pl-10"
            value={searchQuery}
            onChange={(e) => handleSearch(e.target.value)}
          />
        </div>
      </Tabs>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 flex-1 min-h-0">
        {/* Recommendations / Catalog */}
        <Card className="lg:col-span-1 flex flex-col">
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <Cpu className="h-4 w-4 text-primary" />
              {source === "ollama" ? "Recommended for Your Hardware" : "Popular on Hugging Face"}
            </CardTitle>
          </CardHeader>
          <CardContent className="flex-1 overflow-hidden space-y-3">
            <ScrollArea className="h-full">
              <div className="space-y-3 pr-3">
                {source === "ollama" ? (
                  recommendations.map((rec) => (
                    <div
                      key={rec.name}
                      className="p-3 rounded-lg border bg-card/50 space-y-2"
                    >
                      <div className="flex items-center justify-between">
                        <span className="font-mono font-medium text-sm">{rec.name}</span>
                        <Badge variant="outline" className="text-xs">
                          {rec.quantization}
                        </Badge>
                      </div>
                      <p className="text-xs text-muted-foreground">{rec.reason}</p>
                      <div className="flex items-center gap-3 text-xs text-muted-foreground">
                        <span className="flex items-center gap-1">
                          <HardDrive className="h-3 w-3" />
                          {rec.sizeGB} GB
                        </span>
                        <span className="flex items-center gap-1">
                          <Cpu className="h-3 w-3" />
                          {formatBytes(rec.vramRequired)} VRAM
                        </span>
                      </div>
                      {isInstalled(rec.name) ? (
                        <Button size="sm" variant="outline" className="w-full gap-1" disabled>
                          <CheckCircle2 className="h-4 w-4" />
                          Installed
                        </Button>
                      ) : (
                        <Button
                          size="sm"
                          className="w-full"
                          disabled={pulling === rec.name}
                          onClick={() => handlePull(rec.name)}
                        >
                          {pulling === rec.name ? (
                            <>
                              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                              Downloading...
                            </>
                          ) : (
                            <>
                              <Download className="mr-2 h-4 w-4" />
                              Pull
                            </>
                          )}
                        </Button>
                      )}
                    </div>
                  ))
                ) : (
                  loading ? (
                    <div className="flex items-center justify-center py-8">
                      <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                    </div>
                  ) : (
                    hfCatalog.map((model) => (
                      <div
                        key={model.id}
                        className="p-3 rounded-lg border bg-card/50 space-y-2"
                      >
                        <div className="flex items-start justify-between">
                          <div>
                            <p className="font-mono font-medium text-sm">{model.name}</p>
                            <p className="text-xs text-muted-foreground">{model.description}</p>
                          </div>
                          <a
                            href={getHFModelURL(model.id)}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-xs text-muted-foreground hover:text-primary flex items-center gap-1"
                          >
                            <ExternalLink className="h-3 w-3" />
                            HF
                          </a>
                        </div>
                        {model.tags && model.tags.length > 0 && (
                          <div className="flex items-center gap-2 flex-wrap">
                            {model.tags.slice(0, 3).map((tag, idx) => (
                              <Badge key={idx} variant="secondary" className="text-xs">
                                {tag.tag} {tag.quantization}
                              </Badge>
                            ))}
                          </div>
                        )}
                        <Button
                          size="sm"
                          variant="outline"
                          className="w-full gap-1"
                          disabled
                        >
                          <ExternalLink className="h-3.5 w-3.5" />
                          Import Coming Soon
                        </Button>
                      </div>
                    ))
                  )
                )}
              </div>
            </ScrollArea>
          </CardContent>
        </Card>

        {/* Installed Models */}
        <Card className="lg:col-span-2 flex flex-col">
          <CardHeader className="pb-3">
            <CardTitle className="text-base flex items-center gap-2">
              <HardDrive className="h-4 w-4 text-primary" />
              Installed Models
            </CardTitle>
          </CardHeader>
          <CardContent className="flex-1 overflow-hidden">
            <ScrollArea className="h-full">
              {installed.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-48 text-muted-foreground gap-2">
                  <HardDrive className="h-10 w-10 opacity-40" />
                  <p className="font-medium">No models installed</p>
                  <p className="text-sm">Pull a model to get started</p>
                </div>
              ) : (
                <div className="space-y-2 pr-3">
                  {installed.map((model) => (
                    <div
                      key={model.name}
                      className="flex items-center justify-between p-4 rounded-lg border hover:bg-accent/5 transition-colors group"
                    >
                      <div className="space-y-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <p className="font-mono font-medium truncate">{model.name}</p>
                          <Badge variant="secondary" className="text-xs shrink-0">
                            {formatBytes(model.size)}
                          </Badge>
                        </div>
                        <div className="flex items-center gap-3 text-xs text-muted-foreground">
                          <span className="flex items-center gap-1">
                            <Clock className="h-3 w-3" />
                            {model.modified_at ? new Date(model.modified_at).toLocaleDateString() : "Unknown"}
                          </span>
                        </div>
                      </div>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="opacity-0 group-hover:opacity-100 transition-opacity text-destructive hover:text-destructive hover:bg-destructive/10"
                        onClick={() => handleDelete(model.name)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </ScrollArea>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}