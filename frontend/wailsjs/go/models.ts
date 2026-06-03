export namespace logs {
	
	export class Entry {
	    id: number;
	    timestamp: string;
	    level: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.level = source["level"];
	        this.message = source["message"];
	    }
	}

}

export namespace main {
	
	export class RuntimeDirectoryEntry {
	    name: string;
	    path: string;
	    isDir: boolean;
	    size: number;
	    modifiedTime?: string;
	
	    static createFrom(source: any = {}) {
	        return new RuntimeDirectoryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.isDir = source["isDir"];
	        this.size = source["size"];
	        this.modifiedTime = source["modifiedTime"];
	    }
	}
	export class RuntimeExecOptions {
	    workingDirectory?: string;
	    timeoutSeconds?: number;
	    env?: {[key: string]: string};
	    input?: string;
	
	    static createFrom(source: any = {}) {
	        return new RuntimeExecOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workingDirectory = source["workingDirectory"];
	        this.timeoutSeconds = source["timeoutSeconds"];
	        this.env = source["env"];
	        this.input = source["input"];
	    }
	}
	export class RuntimeExecResult {
	    stdout: string;
	    stderr: string;
	    exitCode: number;
	
	    static createFrom(source: any = {}) {
	        return new RuntimeExecResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stdout = source["stdout"];
	        this.stderr = source["stderr"];
	        this.exitCode = source["exitCode"];
	    }
	}
	export class RuntimeFileFilter {
	    displayName: string;
	    pattern: string;
	
	    static createFrom(source: any = {}) {
	        return new RuntimeFileFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.displayName = source["displayName"];
	        this.pattern = source["pattern"];
	    }
	}
	export class RuntimeFileInfo {
	    path: string;
	    exists: boolean;
	    isDir: boolean;
	    size: number;
	    modifiedTime?: string;
	
	    static createFrom(source: any = {}) {
	        return new RuntimeFileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.exists = source["exists"];
	        this.isDir = source["isDir"];
	        this.size = source["size"];
	        this.modifiedTime = source["modifiedTime"];
	    }
	}
	export class RuntimeOSInfo {
	    platform: string;
	    arch: string;
	    family: string;
	    pathSeparator: string;
	    listSeparator: string;
	    lineSeparator: string;
	    homeDir?: string;
	    configDir?: string;
	    cacheDir?: string;
	    tempDir: string;
	
	    static createFrom(source: any = {}) {
	        return new RuntimeOSInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.platform = source["platform"];
	        this.arch = source["arch"];
	        this.family = source["family"];
	        this.pathSeparator = source["pathSeparator"];
	        this.listSeparator = source["listSeparator"];
	        this.lineSeparator = source["lineSeparator"];
	        this.homeDir = source["homeDir"];
	        this.configDir = source["configDir"];
	        this.cacheDir = source["cacheDir"];
	        this.tempDir = source["tempDir"];
	    }
	}
	export class RuntimeOpenFileDialogOptions {
	    title?: string;
	    filters?: RuntimeFileFilter[];
	
	    static createFrom(source: any = {}) {
	        return new RuntimeOpenFileDialogOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.filters = this.convertValues(source["filters"], RuntimeFileFilter);
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
	export class RuntimeSaveFileDialogOptions {
	    title?: string;
	    defaultFilename?: string;
	    filters?: RuntimeFileFilter[];
	
	    static createFrom(source: any = {}) {
	        return new RuntimeSaveFileDialogOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.defaultFilename = source["defaultFilename"];
	        this.filters = this.convertValues(source["filters"], RuntimeFileFilter);
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
	export class RuntimeStoragePaths {
	    toolID?: string;
	    configDir: string;
	    dataDir: string;
	    cacheDir: string;
	    tempDir: string;
	
	    static createFrom(source: any = {}) {
	        return new RuntimeStoragePaths(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.toolID = source["toolID"];
	        this.configDir = source["configDir"];
	        this.dataDir = source["dataDir"];
	        this.cacheDir = source["cacheDir"];
	        this.tempDir = source["tempDir"];
	    }
	}
	export class ToolStorePackage {
	    url: string;
	    registry_key: string;
	    id: string;
	    name: string;
	    version: string;
	    author?: string;
	    description?: string;
	    entry?: string;
	    runtime?: string;
	    icon_data: string;
	    icon_name: string;
	    manifest?: tools.ToolManifest;
	    validation: tools.ToolValidationResult;
	    installed: boolean;
	    installed_version?: string;
	    update_available: boolean;
	    installed_tool_id?: string;
	
	    static createFrom(source: any = {}) {
	        return new ToolStorePackage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.registry_key = source["registry_key"];
	        this.id = source["id"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.author = source["author"];
	        this.description = source["description"];
	        this.entry = source["entry"];
	        this.runtime = source["runtime"];
	        this.icon_data = source["icon_data"];
	        this.icon_name = source["icon_name"];
	        this.manifest = this.convertValues(source["manifest"], tools.ToolManifest);
	        this.validation = this.convertValues(source["validation"], tools.ToolValidationResult);
	        this.installed = source["installed"];
	        this.installed_version = source["installed_version"];
	        this.update_available = source["update_available"];
	        this.installed_tool_id = source["installed_tool_id"];
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

export namespace tools {
	
	export class ToolManifest {
	    id: string;
	    name: string;
	    version: string;
	    author?: string;
	    description?: string;
	    entry: string;
	    runtime: string;
	    icon?: string;
	    requires_admin?: boolean;
	    permissions?: string[];
	    target_platforms?: string[];
	    min_grtbox_version?: string;
	    package_format_version?: string;
	
	    static createFrom(source: any = {}) {
	        return new ToolManifest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.author = source["author"];
	        this.description = source["description"];
	        this.entry = source["entry"];
	        this.runtime = source["runtime"];
	        this.icon = source["icon"];
	        this.requires_admin = source["requires_admin"];
	        this.permissions = source["permissions"];
	        this.target_platforms = source["target_platforms"];
	        this.min_grtbox_version = source["min_grtbox_version"];
	        this.package_format_version = source["package_format_version"];
	    }
	}
	export class ToolMetadata {
	    id: string;
	    name: string;
	    version: string;
	    author?: string;
	    description?: string;
	    entry: string;
	    runtime: string;
	    requires_admin?: boolean;
	    permissions?: string[];
	    target_platforms?: string[];
	    min_grtbox_version?: string;
	    package_format_version?: string;
	    package_location: string;
	    validation_status: string;
	    uses_default_tool_icon: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ToolMetadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.author = source["author"];
	        this.description = source["description"];
	        this.entry = source["entry"];
	        this.runtime = source["runtime"];
	        this.requires_admin = source["requires_admin"];
	        this.permissions = source["permissions"];
	        this.target_platforms = source["target_platforms"];
	        this.min_grtbox_version = source["min_grtbox_version"];
	        this.package_format_version = source["package_format_version"];
	        this.package_location = source["package_location"];
	        this.validation_status = source["validation_status"];
	        this.uses_default_tool_icon = source["uses_default_tool_icon"];
	    }
	}
	export class ToolValidationResult {
	    valid: boolean;
	    message: string;
	    errors: string[];
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new ToolValidationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.message = source["message"];
	        this.errors = source["errors"];
	        this.warnings = source["warnings"];
	    }
	}
	export class ToolPackage {
	    registry_key: string;
	    id: string;
	    name: string;
	    version: string;
	    author?: string;
	    description?: string;
	    entry?: string;
	    runtime?: string;
	    location: string;
	    extracted_path?: string;
	    icon_data: string;
	    icon_name: string;
	    manifest?: ToolManifest;
	    metadata: ToolMetadata;
	    validation: ToolValidationResult;
	
	    static createFrom(source: any = {}) {
	        return new ToolPackage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.registry_key = source["registry_key"];
	        this.id = source["id"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.author = source["author"];
	        this.description = source["description"];
	        this.entry = source["entry"];
	        this.runtime = source["runtime"];
	        this.location = source["location"];
	        this.extracted_path = source["extracted_path"];
	        this.icon_data = source["icon_data"];
	        this.icon_name = source["icon_name"];
	        this.manifest = this.convertValues(source["manifest"], ToolManifest);
	        this.metadata = this.convertValues(source["metadata"], ToolMetadata);
	        this.validation = this.convertValues(source["validation"], ToolValidationResult);
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

