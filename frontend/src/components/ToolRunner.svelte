<script lang="ts">
  import { createEventDispatcher, onDestroy } from 'svelte';
  import TCRenderer from './TCRenderer.svelte';
  import { api } from '../lib/api';
  import type { LogEntry, TCNode, ToolPackage } from '../lib/types';

  export let tool: ToolPackage | null = null;

  const dispatch = createEventDispatcher<{ close: void }>();

  let runningToolKey = '';
  let status: 'idle' | 'loading' | 'running' | 'crashed' | 'stopped' = 'idle';
  let errorMessage = '';
  let rootNode: TCNode | string | null = null;
  let toolLogs: LogEntry[] = [];
  let moduleURLs: string[] = [];
  let moduleURLCache = new Map<string, string>();
  let timers: Array<ReturnType<typeof setInterval> | ReturnType<typeof setTimeout>> = [];
  let events = new EventTarget();

  $: if (tool && toolKey(tool) !== runningToolKey) {
    startTool(tool);
  }

  onDestroy(() => {
    stopTool();
  });

  async function startTool(currentTool: ToolPackage) {
    stopTool();
    runningToolKey = toolKey(currentTool);
    status = 'loading';
    errorMessage = '';
    rootNode = null;
    toolLogs = [];
    moduleURLCache = new Map();
    moduleURLs = [];
    events = new EventTarget();

    try {
      const entry = currentTool.entry || currentTool.manifest?.entry || 'main.tc';
      const entryURL = await loadModuleURL(currentTool.id, entry);
      const module = await import(/* @vite-ignore */ entryURL);
      const main = module.default;
      if (typeof main !== 'function') {
        throw new Error('main.tc must export a default async function.');
      }
      status = 'running';
      await main(createRuntime(currentTool));
    } catch (error) {
      status = 'crashed';
      errorMessage = error instanceof Error ? error.message : String(error);
      await writeToolLog(currentTool.id, `Tool crashed: ${errorMessage}`);
    }
  }

  function stopTool() {
    for (const timer of timers) {
      clearInterval(timer);
      clearTimeout(timer);
    }
    timers = [];
    for (const url of moduleURLs) {
      URL.revokeObjectURL(url);
    }
    moduleURLs = [];
    moduleURLCache.clear();
    if (status === 'running' || status === 'loading') {
      status = 'stopped';
    }
  }

  async function reloadTool() {
    if (tool) await startTool(tool);
  }

  async function loadModuleURL(toolID: string, relativePath: string): Promise<string> {
    const normalized = normalizeModulePath(relativePath);
    const cached = moduleURLCache.get(normalized);
    if (cached) return cached;

    let source = await api.readToolModule(toolID, normalized);
    const imports = findRelativeImports(source);
    for (const importPath of imports) {
      const resolved = resolveModulePath(normalized, importPath);
      const dependencyURL = await loadModuleURL(toolID, resolved);
      source = source.split(importPath).join(dependencyURL);
    }

    const blob = new Blob([source], { type: 'text/javascript' });
    const url = URL.createObjectURL(blob);
    moduleURLCache.set(normalized, url);
    moduleURLs = [...moduleURLs, url];
    return url;
  }

  function createRuntime(currentTool: ToolPackage) {
    return {
      tool: {
        id: currentTool.id,
        name: currentTool.name,
        version: currentTool.version,
        extractedPath: currentTool.extracted_path,
        manifest: currentTool.manifest
      },
      ui: {
        createElement(type: string, props: Record<string, unknown> = {}, children: unknown = []) {
          return { type, props, children } as TCNode;
        },
        setRoot(node: TCNode | string) {
          rootNode = node;
        },
        render(node: TCNode | string) {
          rootNode = node;
        },
        update(node: TCNode | string) {
          rootNode = node;
        },
        clear() {
          rootNode = null;
        }
      },
      os: {
        platform: () => api.runtimeOSPlatform(),
        info: () => api.runtimeOSInfo(),
        isAdmin: () => api.runtimeOSIsAdmin(),
        isWindows: async () => (await api.runtimeOSPlatform()) === 'windows',
        isLinux: async () => (await api.runtimeOSPlatform()) === 'linux',
        isMacOS: async () => (await api.runtimeOSPlatform()) === 'darwin'
      },
      process: {
        exec: (command: string, args: string[] = [], options = {}) => api.runtimeProcessExec(command, args, options),
        which: (command: string) => api.runtimeProcessWhich(command)
      },
      shell: {
        exec: (command: string, options = {}) => api.runtimeShellExec(command, options)
      },
      powershell: {
        exec: (script: string, options = {}) => api.runtimePowerShellExec(script, options)
      },
      filesystem: {
        readFile: (path: string) => api.runtimeFilesystemReadFile(path),
        readFileBase64: (path: string) => api.runtimeFilesystemReadFileBase64(path),
        writeFile: (path: string, content: string) => api.runtimeFilesystemWriteFile(path, content),
        stat: (path: string) => api.runtimeFilesystemStat(path),
        listDir: (path: string) => api.runtimeFilesystemListDir(path),
        mkdirAll: (path: string) => api.runtimeFilesystemMkdirAll(path),
        removeFile: (path: string) => api.runtimeFilesystemRemoveFile(path),
        removeDir: (path: string) => api.runtimeFilesystemRemoveDir(path),
        exists: (path: string) => api.runtimeFilesystemExists(path)
      },
      path: {
        join: (...parts: string[]) =>
          parts
            .filter(Boolean)
            .join('/')
            .replace(/\\/g, '/')
            .replace(/([^:])\/+/g, '$1/'),
        basename: (value: string) => value.split(/[\\/]/).pop() || '',
        dirname: (value: string) => value.split(/[\\/]/).slice(0, -1).join('/') || '.',
        nativeJoin: (...parts: string[]) => api.runtimePathJoin(parts),
        normalize: (path: string) => api.runtimePathNormalize(path),
        toSlash: (path: string) => api.runtimePathToSlash(path),
        fromSlash: (path: string) => api.runtimePathFromSlash(path),
        nativeBasename: (path: string) => api.runtimePathBaseName(path),
        nativeDirname: (path: string) => api.runtimePathDirName(path),
        extname: (path: string) => api.runtimePathExtName(path),
        isAbs: (path: string) => api.runtimePathIsAbs(path)
      },
      storage: {
        paths: () => api.runtimeStoragePaths(currentTool.id),
        ensure: () => api.runtimeStorageEnsure(currentTool.id)
      },
      crypto: {
        hashFile: (path: string, algorithm = 'sha256') => api.runtimeCryptoHashFile(path, algorithm)
      },
      env: {
        get: (name: string) => api.runtimeEnvGet(name)
      },
      dialogs: {
        openFile: (options = {}) => api.runtimeDialogsOpenFile(options),
        saveFile: (options = {}) => api.runtimeDialogsSaveFile(options),
        confirm: (message: string) => Promise.resolve(window.confirm(message))
      },
      logs: {
        write: (message: string) => writeToolLog(currentTool.id, message)
      },
      timers: {
        setInterval: (callback: () => void, delay: number) => {
          const timer = setInterval(callback, delay);
          timers = [...timers, timer];
          return timer;
        },
        setTimeout: (callback: () => void, delay: number) => {
          const timer = setTimeout(callback, delay);
          timers = [...timers, timer];
          return timer;
        },
        clearInterval: (timer: ReturnType<typeof setInterval>) => {
          clearInterval(timer);
          timers = timers.filter((item) => item !== timer);
        },
        clearTimeout: (timer: ReturnType<typeof setTimeout>) => {
          clearTimeout(timer);
          timers = timers.filter((item) => item !== timer);
        }
      },
      http: {
        get: async (url: string, options = {}) => {
          const response = await fetch(url, options);
          return readTextResponse(response);
        },
        post: async (url: string, body: unknown, options: RequestInit = {}) => {
          const response = await fetch(url, { ...options, method: 'POST', body: typeof body === 'string' ? body : JSON.stringify(body) });
          return readTextResponse(response);
        }
      },
      json: JSON,
      events: {
        on: (name: string, callback: EventListener) => events.addEventListener(name, callback),
        off: (name: string, callback: EventListener) => events.removeEventListener(name, callback),
        emit: (name: string, detail?: unknown) => events.dispatchEvent(new CustomEvent(name, { detail }))
      }
    };
  }

  async function writeToolLog(toolID: string, message: string) {
    const entry = await api.runtimeLogsWrite(toolID, message);
    toolLogs = [...toolLogs, entry];
    return entry;
  }

  async function readTextResponse(response: Response) {
    const text = await response.text();
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${text.slice(0, 240) || response.statusText}`);
    }
    return text;
  }

  function findRelativeImports(source: string) {
    const matches = new Set<string>();
    const pattern = /(?:import\s+(?:[^'"]+\s+from\s+)?|export\s+[^'"]+\s+from\s+)["'](\.{1,2}\/[^"']+\.tc)["']/g;
    for (const match of source.matchAll(pattern)) {
      matches.add(match[1]);
    }
    return [...matches];
  }

  function normalizeModulePath(path: string) {
    return path.replace(/\\/g, '/').replace(/^\.\//, '');
  }

  function resolveModulePath(from: string, target: string) {
    const stack = from.split('/').slice(0, -1);
    for (const part of target.split('/')) {
      if (part === '.' || part === '') continue;
      if (part === '..') stack.pop();
      else stack.push(part);
    }
    return normalizeModulePath(stack.join('/'));
  }

  function toolKey(currentTool: ToolPackage) {
    return currentTool.registry_key || currentTool.id;
  }
</script>

{#if !tool}
  <section class="runner-empty">
    <strong>No tool selected</strong>
  </section>
{:else}
  <section class="runner">
    <header>
      <div>
        <span>TC Runtime</span>
        <h2>{tool.name}</h2>
      </div>
      <div class="runner-actions">
        <strong>{status}</strong>
        <button type="button" on:click={reloadTool}>Reload Tool</button>
        <button type="button" on:click={() => { stopTool(); dispatch('close'); }}>Back to Tools</button>
      </div>
    </header>

    {#if status === 'crashed'}
      <section class="crash-panel">
        <h3>Tool crashed</h3>
        <p>{errorMessage}</p>
        <div>
          <button type="button" on:click={reloadTool}>Reload Tool</button>
          <button type="button" on:click={() => dispatch('close')}>Back to Tools</button>
        </div>
      </section>
    {:else}
      <div class="tc-root">
        {#if rootNode}
          <TCRenderer node={rootNode} />
        {:else if status === 'loading'}
          <p class="runtime-note">Loading TC application...</p>
        {:else}
          <p class="runtime-note">The tool is running but has not rendered UI yet.</p>
        {/if}
      </div>
    {/if}

    <section class="tool-log">
      <h3>Tool Logs</h3>
      {#if toolLogs.length === 0}
        <p>No runtime messages yet.</p>
      {:else}
        <ol>
          {#each toolLogs.slice().reverse() as entry (entry.id)}
            <li><time>{new Date(entry.timestamp).toLocaleTimeString()}</time><span>{entry.message}</span></li>
          {/each}
        </ol>
      {/if}
    </section>
  </section>
{/if}

<style>
  .runner,
  .runner-empty {
    max-width: 1100px;
    border: 1px solid #444444;
    border-radius: 10px;
    background: #333333;
  }

  .runner-empty {
    min-height: 220px;
    display: grid;
    place-items: center;
    color: #b9b9b9;
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 18px;
    border-bottom: 1px solid #444444;
  }

  header span {
    display: block;
    margin-bottom: 4px;
    color: #b9b9b9;
    font-size: 0.75rem;
    font-weight: 800;
    text-transform: uppercase;
  }

  h2,
  h3 {
    margin: 0;
    color: #ffffff;
    letter-spacing: 0;
  }

  h2 {
    font-size: 1.3rem;
  }

  h3 {
    font-size: 1rem;
  }

  .runner-actions {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 8px;
  }

  strong {
    padding: 6px 9px;
    border: 1px solid #505050;
    border-radius: 8px;
    color: #eeeeee;
    background: #2d2d2d;
    font-size: 0.82rem;
    text-transform: capitalize;
  }

  button {
    min-height: 36px;
    padding: 0 12px;
    border: 1px solid #505050;
    border-radius: 8px;
    color: #eeeeee;
    background: #2b2b2b;
    cursor: pointer;
    font-weight: 800;
  }

  button:hover {
    border-color: #666666;
    background: #414141;
  }

  .tc-root,
  .crash-panel,
  .tool-log {
    display: grid;
    gap: 14px;
    padding: 18px;
  }

  .crash-panel {
    border-bottom: 1px solid #444444;
    background: #3a2e2e;
  }

  .crash-panel p,
  .runtime-note,
  .tool-log p {
    margin: 0;
    color: #d0d0d0;
    line-height: 1.5;
  }

  .crash-panel div {
    display: flex;
    gap: 8px;
  }

  .tool-log {
    border-top: 1px solid #444444;
  }

  ol {
    display: grid;
    gap: 8px;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  li {
    display: grid;
    grid-template-columns: 90px minmax(0, 1fr);
    gap: 10px;
    padding: 9px 10px;
    border: 1px solid #444444;
    border-radius: 8px;
    background: #252525;
  }

  time {
    color: #a9a9a9;
    font-size: 0.8rem;
  }

  li span {
    color: #eeeeee;
    overflow-wrap: anywhere;
  }

  @media (max-width: 720px) {
    header {
      align-items: flex-start;
      flex-direction: column;
    }

    .runner-actions {
      justify-content: flex-start;
    }
  }
</style>
