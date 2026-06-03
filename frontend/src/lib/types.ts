export type ViewName = 'home' | 'tools' | 'store' | 'details' | 'runtime' | 'settings' | 'logs' | 'about';

export interface ToolValidationResult {
  valid: boolean;
  message: string;
  errors: string[];
  warnings: string[];
}

export interface ToolManifest {
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
}

export interface ToolMetadata {
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
}

export interface ToolPackage {
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
}

export interface ToolStorePackage {
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
  manifest?: ToolManifest;
  validation: ToolValidationResult;
  installed: boolean;
  installed_version?: string;
  update_available: boolean;
  installed_tool_id?: string;
}

export interface LogEntry {
  id: number;
  timestamp: string;
  level: string;
  message: string;
}

export interface TCNode {
  type: string;
  props?: Record<string, unknown>;
  children?: Array<TCNode | string> | TCNode | string;
}

export interface RuntimeExecOptions {
  workingDirectory?: string;
  timeoutSeconds?: number;
  env?: Record<string, string>;
  input?: string;
}

export interface RuntimeExecResult {
  stdout: string;
  stderr: string;
  exitCode: number;
}

export interface RuntimeOSInfo {
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
}

export interface RuntimeStoragePaths {
  toolID?: string;
  configDir: string;
  dataDir: string;
  cacheDir: string;
  tempDir: string;
}

export interface RuntimeFileInfo {
  path: string;
  exists: boolean;
  isDir: boolean;
  size: number;
  modifiedTime?: string;
}

export interface RuntimeDirectoryEntry {
  name: string;
  path: string;
  isDir: boolean;
  size: number;
  modifiedTime?: string;
}

export interface RuntimeFileFilter {
  displayName: string;
  pattern: string;
}

export interface RuntimeOpenFileDialogOptions {
  title?: string;
  filters?: RuntimeFileFilter[];
}

export interface RuntimeSaveFileDialogOptions {
  title?: string;
  defaultFilename?: string;
  filters?: RuntimeFileFilter[];
}
