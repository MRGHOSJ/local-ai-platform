export namespace hardware {
	
	export class CPUInfo {
	    model: string;
	    cores: number;
	    threads: number;
	    architecture: string;
	    features: string[];
	    frequencyMHz: number;
	
	    static createFrom(source: any = {}) {
	        return new CPUInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.cores = source["cores"];
	        this.threads = source["threads"];
	        this.architecture = source["architecture"];
	        this.features = source["features"];
	        this.frequencyMHz = source["frequencyMHz"];
	    }
	}
	export class GPUInfo {
	    index: number;
	    vendor: string;
	    model: string;
	    vramBytes: number;
	    driver: string;
	    backend: string;
	    isSupported: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GPUInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.vendor = source["vendor"];
	        this.model = source["model"];
	        this.vramBytes = source["vramBytes"];
	        this.driver = source["driver"];
	        this.backend = source["backend"];
	        this.isSupported = source["isSupported"];
	    }
	}
	export class MemoryInfo {
	    totalBytes: number;
	    usedBytes: number;
	    freeBytes: number;
	    usedPercent: number;
	
	    static createFrom(source: any = {}) {
	        return new MemoryInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalBytes = source["totalBytes"];
	        this.usedBytes = source["usedBytes"];
	        this.freeBytes = source["freeBytes"];
	        this.usedPercent = source["usedPercent"];
	    }
	}
	export class OSInfo {
	    platform: string;
	    arch: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new OSInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.platform = source["platform"];
	        this.arch = source["arch"];
	        this.version = source["version"];
	    }
	}
	export class StorageInfo {
	    totalBytes: number;
	    freeBytes: number;
	    usedPercent: number;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new StorageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalBytes = source["totalBytes"];
	        this.freeBytes = source["freeBytes"];
	        this.usedPercent = source["usedPercent"];
	        this.path = source["path"];
	    }
	}
	export class SystemSpecs {
	    cpu: CPUInfo;
	    gpus: GPUInfo[];
	    memory: MemoryInfo;
	    storage: StorageInfo;
	    os: OSInfo;
	    score: number;
	    // Go type: time
	    detectedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new SystemSpecs(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cpu = this.convertValues(source["cpu"], CPUInfo);
	        this.gpus = this.convertValues(source["gpus"], GPUInfo);
	        this.memory = this.convertValues(source["memory"], MemoryInfo);
	        this.storage = this.convertValues(source["storage"], StorageInfo);
	        this.os = this.convertValues(source["os"], OSInfo);
	        this.score = source["score"];
	        this.detectedAt = this.convertValues(source["detectedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace hcl {
	
	export class ModelRecommendation {
	    name: string;
	    sizeGB: number;
	    quantization: string;
	    vramRequired: number;
	    ramRequired: number;
	    reason: string;
	    recommended: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ModelRecommendation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.sizeGB = source["sizeGB"];
	        this.quantization = source["quantization"];
	        this.vramRequired = source["vramRequired"];
	        this.ramRequired = source["ramRequired"];
	        this.reason = source["reason"];
	        this.recommended = source["recommended"];
	    }
	}

}

export namespace models {
	
	export class ModelTag {
	    tag: string;
	    quantization: string;
	    sizeGB: number;
	    vramRequiredGB: number;
	    ramRequiredGB: number;
	    sourceFile: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelTag(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tag = source["tag"];
	        this.quantization = source["quantization"];
	        this.sizeGB = source["sizeGB"];
	        this.vramRequiredGB = source["vramRequiredGB"];
	        this.ramRequiredGB = source["ramRequiredGB"];
	        this.sourceFile = source["sourceFile"];
	    }
	}
	export class CatalogModel {
	    id: string;
	    name: string;
	    description: string;
	    provider: string;
	    totalPulls: number;
	    tags: ModelTag[];
	
	    static createFrom(source: any = {}) {
	        return new CatalogModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.provider = source["provider"];
	        this.totalPulls = source["totalPulls"];
	        this.tags = this.convertValues(source["tags"], ModelTag);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace ollama {
	
	export class ModelInfo {
	    name: string;
	    size: number;
	    modified_at: string;
	    digest: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.size = source["size"];
	        this.modified_at = source["modified_at"];
	        this.digest = source["digest"];
	    }
	}

}

