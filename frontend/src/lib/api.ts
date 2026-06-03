import type {
  LogEntry,
  RuntimeDirectoryEntry,
  RuntimeExecOptions,
  RuntimeExecResult,
  RuntimeFileInfo,
  RuntimeOpenFileDialogOptions,
  RuntimeSaveFileDialogOptions,
  RuntimeOSInfo,
  RuntimeStoragePaths,
  ToolPackage,
  ToolStorePackage,
  ToolValidationResult
} from './types';

declare global {
  interface Window {
    go?: {
      main?: {
        App?: Record<string, (...args: unknown[]) => Promise<unknown>>;
      };
    };
  }
}

const defaultIcon = `data:image/svg+xml;utf8,${encodeURIComponent(
  '<svg xmlns="http://www.w3.org/2000/svg" width="96" height="96" viewBox="0 0 96 96" role="img" aria-label="Default Tool Icon"><rect width="96" height="96" rx="18" fill="#252525"/><rect x="30" y="34" width="36" height="28" rx="3" fill="#e8e8e8"/><path d="M35 40h26v4H35zm0 9h18v4H35z" fill="#2d2d2d"/><path d="M24 46h6v8h-6zm42 0h6v8h-6z" fill="#8d8d8d"/><path d="M42 25h12v9H42z" fill="#8d8d8d"/></svg>'
)}`;

const mockTools: ToolPackage[] = [
  {
    registry_key: 'sample_tool',
    id: 'sample_tool',
    name: 'Sample Tool',
    version: '1.0.0',
    author: 'GRTBox',
    description: 'Example TC tool package.',
    entry: 'main.tc',
    runtime: 'tc',
    location: 'mock/sample_tool.tl',
    icon_data: defaultIcon,
    icon_name: 'Default Tool Icon',
    manifest: {
      id: 'sample_tool',
      name: 'Sample Tool',
      version: '1.0.0',
      author: 'GRTBox',
      description: 'Example TC tool package.',
      entry: 'main.tc',
      runtime: 'tc',
      permissions: ['logs.write'],
      target_platforms: ['windows', 'linux', 'macos'],
      min_grtbox_version: '0.1.0',
      package_format_version: '1.0.0'
    },
    metadata: {
      id: 'sample_tool',
      name: 'Sample Tool',
      version: '1.0.0',
      author: 'GRTBox',
      description: 'Example TC tool package.',
      entry: 'main.tc',
      runtime: 'tc',
      permissions: ['logs.write'],
      target_platforms: ['windows', 'linux', 'macos'],
      min_grtbox_version: '0.1.0',
      package_format_version: '1.0.0',
      package_location: 'mock/sample_tool.tl',
      validation_status: 'Package Valid',
      uses_default_tool_icon: true
    },
    validation: {
      valid: true,
      message: 'Package Valid',
      errors: [],
      warnings: []
    }
  }
];

const mockStoreTools: ToolStorePackage[] = [];

let mockLogs: LogEntry[] = [
  {
    id: 1,
    timestamp: new Date().toISOString(),
    level: 'info',
    message: 'Frontend preview runtime loaded'
  }
];

function backendMethod(name: string) {
  return window.go?.main?.App?.[name];
}

async function callBackend<T>(
  name: string,
  args: unknown[],
  fallback: () => T | Promise<T>
): Promise<T> {
  const method = backendMethod(name);
  if (typeof method === 'function') {
    return (await method(...args)) as T;
  }
  return fallback();
}

function invalidNativeResult(): ToolValidationResult {
  return {
    valid: false,
    message: 'Native desktop runtime is not available.',
    errors: ['This action is available when GRTBox is running inside Wails.'],
    warnings: []
  };
}

export const api = {
  listTools() {
    return callBackend<ToolPackage[]>('ListTools', [], () => mockTools);
  },
  refreshTools() {
    return callBackend<ToolPackage[]>('RefreshTools', [], () => mockTools);
  },
  installTool(filePath: string) {
    return callBackend<ToolValidationResult>('InstallTool', [filePath], () => invalidNativeResult());
  },
  updateTool(filePath: string) {
    return callBackend<ToolValidationResult>('UpdateTool', [filePath], () => invalidNativeResult());
  },
  previewToolPackage(filePath: string) {
    return callBackend<ToolPackage>('PreviewToolPackage', [filePath], () => {
      throw new Error('Native desktop runtime is not available.');
    });
  },
  removeTool(toolID: string) {
    return callBackend<void>('RemoveTool', [toolID], () => {
      const index = mockTools.findIndex((tool) => tool.registry_key === toolID || tool.id === toolID);
      if (index >= 0) mockTools.splice(index, 1);
    });
  },
  getToolDetails(toolID: string) {
    return callBackend<ToolPackage>('GetToolDetails', [toolID], () => {
      const tool = mockTools.find((item) => item.registry_key === toolID || item.id === toolID);
      if (!tool) throw new Error('Tool not found');
      return tool;
    });
  },
  openTool(toolID: string) {
    return callBackend<ToolPackage>('OpenTool', [toolID], () => {
      const tool = mockTools.find((item) => item.registry_key === toolID || item.id === toolID);
      if (!tool) throw new Error('Tool not found');
      return tool;
    });
  },
  validateTool(filePath: string) {
    return callBackend<ToolValidationResult>('ValidateTool', [filePath], () => invalidNativeResult());
  },
  listToolStore() {
    return callBackend<ToolStorePackage[]>('ListToolStore', [], () => mockStoreTools);
  },
  installStoreTool(url: string) {
    return callBackend<ToolValidationResult>('InstallStoreTool', [url], () => invalidNativeResult());
  },
  updateStoreTool(url: string) {
    return callBackend<ToolValidationResult>('UpdateStoreTool', [url], () => invalidNativeResult());
  },
  getLogs() {
    return callBackend<LogEntry[]>('GetLogs', [], () => mockLogs);
  },
  selectToolPackage() {
    return callBackend<string>('SelectToolPackage', [], () => '');
  },
  runToolAction(toolID: string, action: string) {
    return callBackend<LogEntry>('RunToolAction', [toolID, action], () => {
      const entry = {
        id: mockLogs.length + 1,
        timestamp: new Date().toISOString(),
        level: 'info',
        message: `${toolID} executed action "${action}"`
      };
      mockLogs = [...mockLogs, entry];
      return entry;
    });
  },
  getToolsDirectory() {
    return callBackend<string>('GetToolsDirectory', [], () => 'AppData/Roaming/GRTBox/tools');
  },
  openExternalURL(url: string) {
    return callBackend<void>('OpenExternalURL', [url], () => {
      window.open(url, '_blank', 'noopener,noreferrer');
    });
  },
  readToolModule(toolID: string, relativePath: string) {
    return callBackend<string>('ReadToolModule', [toolID, relativePath], () => {
      if (relativePath.includes('ui.tc')) {
        return `export async function render(runtime) {
  const h = runtime.ui.createElement;
  runtime.ui.render(h('div', { className: 'tc-panel' }, [
    h('h1', {}, 'Sample Tool'),
    h('p', {}, 'This is a sample TC tool.'),
    h('button', { onClick: () => runtime.logs.write('Sample action executed.') }, 'Test Action')
  ]));
}`;
      }
      return `import { render } from './src/ui.tc';
export default async function main(runtime) {
  await render(runtime);
}`;
    });
  },
  runtimeOSPlatform() {
    return callBackend<string>('RuntimeOSPlatform', [], () => 'browser-preview');
  },
  runtimeOSInfo() {
    return callBackend<RuntimeOSInfo>('RuntimeOSInfo', [], () => ({
      platform: 'browser-preview',
      arch: 'unknown',
      family: 'browser',
      pathSeparator: '/',
      listSeparator: ';',
      lineSeparator: '\n',
      tempDir: ''
    }));
  },
  runtimeOSIsAdmin() {
    return callBackend<boolean>('RuntimeOSIsAdmin', [], () => false);
  },
  runtimeProcessExec(command: string, args: string[], options: RuntimeExecOptions = {}) {
    return callBackend<RuntimeExecResult>('RuntimeProcessExec', [command, args, options], () => ({
      stdout: '',
      stderr: 'Native desktop runtime is not available.',
      exitCode: 1
    }));
  },
  runtimeProcessWhich(command: string) {
    return callBackend<string>('RuntimeProcessWhich', [command], () => {
      throw new Error('Native desktop runtime is not available.');
    });
  },
  runtimeShellExec(command: string, options: RuntimeExecOptions = {}) {
    return callBackend<RuntimeExecResult>('RuntimeShellExec', [command, options], () => ({
      stdout: '',
      stderr: 'Native desktop runtime is not available.',
      exitCode: 1
    }));
  },
  runtimePowerShellExec(script: string, options: RuntimeExecOptions = {}) {
    return callBackend<RuntimeExecResult>('RuntimePowerShellExec', [script, options], () => ({
      stdout: '',
      stderr: 'Native desktop runtime is not available.',
      exitCode: 1
    }));
  },
  runtimeFilesystemReadFile(path: string) {
    return callBackend<string>('RuntimeFilesystemReadFile', [path], () => '');
  },
  runtimeFilesystemReadFileBase64(path: string) {
    return callBackend<string>('RuntimeFilesystemReadFileBase64', [path], () => '');
  },
  runtimeFilesystemWriteFile(path: string, content: string) {
    return callBackend<void>('RuntimeFilesystemWriteFile', [path, content], () => undefined);
  },
  runtimeFilesystemStat(path: string) {
    return callBackend<RuntimeFileInfo>('RuntimeFilesystemStat', [path], () => ({
      path,
      exists: false,
      isDir: false,
      size: 0
    }));
  },
  runtimeFilesystemListDir(path: string) {
    return callBackend<RuntimeDirectoryEntry[]>('RuntimeFilesystemListDir', [path], () => []);
  },
  runtimeFilesystemMkdirAll(path: string) {
    return callBackend<void>('RuntimeFilesystemMkdirAll', [path], () => undefined);
  },
  runtimeFilesystemRemoveFile(path: string) {
    return callBackend<void>('RuntimeFilesystemRemoveFile', [path], () => undefined);
  },
  runtimeFilesystemRemoveDir(path: string) {
    return callBackend<void>('RuntimeFilesystemRemoveDir', [path], () => undefined);
  },
  runtimeFilesystemExists(path: string) {
    return callBackend<boolean>('RuntimeFilesystemExists', [path], () => false);
  },
  runtimeEnvGet(name: string) {
    return callBackend<string>('RuntimeEnvGet', [name], () => '');
  },
  runtimeStoragePaths(toolID: string) {
    return callBackend<RuntimeStoragePaths>('RuntimeStoragePaths', [toolID], () => ({
      toolID,
      configDir: '',
      dataDir: '',
      cacheDir: '',
      tempDir: ''
    }));
  },
  runtimeStorageEnsure(toolID: string) {
    return callBackend<RuntimeStoragePaths>('RuntimeStorageEnsure', [toolID], () => ({
      toolID,
      configDir: '',
      dataDir: '',
      cacheDir: '',
      tempDir: ''
    }));
  },
  runtimePathJoin(parts: string[]) {
    return callBackend<string>('RuntimePathJoin', [parts], () =>
      parts
        .filter(Boolean)
        .join('/')
        .replace(/\\/g, '/')
        .replace(/([^:])\/+/g, '$1/')
    );
  },
  runtimePathNormalize(path: string) {
    return callBackend<string>('RuntimePathNormalize', [path], () => path.replace(/\\/g, '/'));
  },
  runtimePathToSlash(path: string) {
    return callBackend<string>('RuntimePathToSlash', [path], () => path.replace(/\\/g, '/'));
  },
  runtimePathFromSlash(path: string) {
    return callBackend<string>('RuntimePathFromSlash', [path], () => path);
  },
  runtimePathBaseName(path: string) {
    return callBackend<string>('RuntimePathBaseName', [path], () => path.split(/[\\/]/).pop() || '');
  },
  runtimePathDirName(path: string) {
    return callBackend<string>('RuntimePathDirName', [path], () => path.split(/[\\/]/).slice(0, -1).join('/') || '.');
  },
  runtimePathExtName(path: string) {
    return callBackend<string>('RuntimePathExtName', [path], () => {
      const base = path.split(/[\\/]/).pop() || '';
      const index = base.lastIndexOf('.');
      return index >= 0 ? base.slice(index) : '';
    });
  },
  runtimePathIsAbs(path: string) {
    return callBackend<boolean>('RuntimePathIsAbs', [path], () => /^([a-zA-Z]:[\\/]|\/)/.test(path));
  },
  runtimeCryptoHashFile(path: string, algorithm = 'sha256') {
    return callBackend<string>('RuntimeCryptoHashFile', [path, algorithm], () => {
      throw new Error('Native desktop runtime is not available.');
    });
  },
  runtimeDialogsOpenFile(options: RuntimeOpenFileDialogOptions = {}) {
    return callBackend<string>('RuntimeDialogsOpenFile', [options], () => '');
  },
  runtimeDialogsSaveFile(options: RuntimeSaveFileDialogOptions = {}) {
    return callBackend<string>('RuntimeDialogsSaveFile', [options], () => '');
  },
  runtimeLogsWrite(toolID: string, message: string) {
    return callBackend<LogEntry>('RuntimeLogsWrite', [toolID, message], () => {
      const entry = {
        id: mockLogs.length + 1,
        timestamp: new Date().toISOString(),
        level: 'info',
        message: `${toolID}: ${message}`
      };
      mockLogs = [...mockLogs, entry];
      return entry;
    });
  }
};
