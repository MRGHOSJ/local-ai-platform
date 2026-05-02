export interface SystemSpecs {
  cpu: CPUInfo;
  gpus: GPUInfo[];
  memory: MemoryInfo;
  storage: StorageInfo;
  os: OSInfo;
  score: number;
  detectedAt: string;
}

export interface CPUInfo {
  model: string;
  cores: number;
  threads: number;
  architecture: string;
  features: string[];
  frequencyMHz: number;
}

export interface GPUInfo {
  index: number;
  vendor: string;
  model: string;
  vramBytes: number;
  driver: string;
  backend: string;
  isSupported: boolean;
}

export interface MemoryInfo {
  totalBytes: number;
  usedBytes: number;
  freeBytes: number;
  usedPercent: number;
}

export interface StorageInfo {
  totalBytes: number;
  freeBytes: number;
  usedPercent: number;
  path: string;
}

export interface OSInfo {
  platform: string;
  arch: string;
  version: string;
}

export interface ModelInfo {
  name: string;
  size: number;
  modified_at: string;
  digest: string;
}

export interface ModelRecommendation {
  name: string;
  sizeGB: number;
  quantization: string;
  vramRequired: number;
  ramRequired: number;
  reason: string;
  recommended: boolean;
}

export interface CatalogModel {
  id: string;
  name: string;
  description: string;
  provider: string;
  totalPulls: number;
  tags: ModelTag[];
}

export interface ModelTag {
  tag: string;
  quantization: string;
  sizeGB: number;
  vramRequiredGB: number;
  ramRequiredGB: number;
  sourceFile: string;
}

export interface DownloadProgress {
  downloaded: number;
  total: number;
  percent: number;
}

declare global {
  interface Window {
    go: {
      main: {
        App: {
          HardwareDetect(): Promise<SystemSpecs>;
          HardwareGetCached(): Promise<SystemSpecs>;
          HardwareClearCache(): Promise<void>;
          EnsureOllama(): Promise<void>;
          GetOllamaVersion(): Promise<string>;
          OllamaStart(): Promise<void>;
          OllamaStop(): Promise<void>;
          OllamaStatus(): Promise<boolean>;
          ListModels(): Promise<ModelInfo[]>;
          PullModel(name: string): Promise<void>;
          DeleteModel(name: string): Promise<void>;
          GetCacheSize(): Promise<number>;
          GetCurrentModel(): Promise<string>;
          GetModelRecommendations(): Promise<ModelRecommendation[]>;
          SearchHFModels(query: string): Promise<CatalogModel[]>;
          GetPopularHFModels(): Promise<CatalogModel[]>;
          GetAllPopularHFModels(): Promise<CatalogModel[]>;
        };
      };
    };
  }
}

const detectHardware = (): Promise<SystemSpecs> => window.go.main.App.HardwareDetect();
const getCachedHardware = (): Promise<SystemSpecs> => window.go.main.App.HardwareGetCached();
const clearHardwareCache = (): Promise<void> => window.go.main.App.HardwareClearCache();

const ensureOllama = (): Promise<void> => window.go.main.App.EnsureOllama();
const getOllamaVersion = (): Promise<string> => window.go.main.App.GetOllamaVersion();
const ollamaStart = (): Promise<void> => window.go.main.App.OllamaStart();
const ollamaStop = (): Promise<void> => window.go.main.App.OllamaStop();
const ollamaStatus = (): Promise<boolean> => window.go.main.App.OllamaStatus();
const listModels = (): Promise<ModelInfo[]> => window.go.main.App.ListModels();
const pullModel = (name: string): Promise<void> => window.go.main.App.PullModel(name);
const deleteModel = (name: string): Promise<void> => window.go.main.App.DeleteModel(name);
const getCacheSize = (): Promise<number> => window.go.main.App.GetCacheSize();
const getCurrentModel = (): Promise<string> => window.go.main.App.GetCurrentModel();
const getModelRecommendations = (): Promise<ModelRecommendation[]> => window.go.main.App.GetModelRecommendations();
const searchHFModels = (query: string): Promise<CatalogModel[]> => window.go.main.App.SearchHFModels(query);
const getPopularHFModels = (): Promise<CatalogModel[]> => window.go.main.App.GetPopularHFModels();
const getAllPopularHFModels = (): Promise<CatalogModel[]> => window.go.main.App.GetAllPopularHFModels();

export { detectHardware, getCachedHardware, clearHardwareCache };
export { ensureOllama, getOllamaVersion, ollamaStart, ollamaStop, ollamaStatus, listModels, pullModel, deleteModel, getCacheSize, getCurrentModel, getModelRecommendations };
export { searchHFModels, getPopularHFModels, getAllPopularHFModels };

export function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(decimals)) + " " + sizes[i];
}

export function formatMHz(mhz: number): string {
  if (mhz >= 1000) return (mhz / 1000).toFixed(2) + " GHz";
  return mhz.toFixed(0) + " MHz";
}