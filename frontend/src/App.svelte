<script lang="ts">
  import { onMount } from 'svelte';
  import { Download, ExternalLink, RefreshCw, Search, ShoppingBag } from '@lucide/svelte';
  import Sidebar from './components/Sidebar.svelte';
  import ToolGrid from './components/ToolGrid.svelte';
  import ToolDetails from './components/ToolDetails.svelte';
  import ToolStoreDetails from './components/ToolStoreDetails.svelte';
  import ToolStoreGrid from './components/ToolStoreGrid.svelte';
  import ToolRunner from './components/ToolRunner.svelte';
  import LogPanel from './components/LogPanel.svelte';
  import { api } from './lib/api';
  import type { LogEntry, ToolPackage, ToolStorePackage, ToolValidationResult, ViewName } from './lib/types';

  const appVersion = '0.1.0';
  const projectSiteURL = 'https://grtbox.unaux.com/';
  const settingsStorageKey = 'grtbox.settings.v1';

  type AppSettings = {
    confirmToolUpdates: boolean;
    showInvalidTools: boolean;
    compactToolCards: boolean;
    reduceMotion: boolean;
  };

  let activeView: ViewName = 'tools';
  let tools: ToolPackage[] = [];
  let storeTools: ToolStorePackage[] = [];
  let logs: LogEntry[] = [];
  let selectedTool: ToolPackage | null = null;
  let selectedStoreTool: ToolStorePackage | null = null;
  let toolsDirectory = '';
  let loading = false;
  let storeLoading = false;
  let storeBusyURL = '';
  let notice = '';
  let noticeTone: 'info' | 'error' | 'success' = 'info';
  let searchQuery = '';
  let storeSearchQuery = '';
  let appSettings: AppSettings = defaultSettings();

  $: validTools = tools.filter((tool) => tool.validation?.valid).length;
  $: invalidTools = tools.length - validTools;
  $: visibleTools = appSettings.showInvalidTools ? tools : tools.filter((tool) => tool.validation?.valid);
  $: normalizedSearch = searchQuery.trim().toLowerCase();
  $: filteredTools = normalizedSearch
    ? visibleTools.filter((tool) => matchesToolSearch(tool, normalizedSearch))
    : visibleTools;
  $: normalizedStoreSearch = storeSearchQuery.trim().toLowerCase();
  $: filteredStoreTools = normalizedStoreSearch
    ? storeTools.filter((tool) => matchesStoreToolSearch(tool, normalizedStoreSearch))
    : storeTools;
  $: sidebarView =
    activeView === 'details' || activeView === 'runtime' || activeView === 'store'
      ? ('tools' as ViewName)
      : activeView;
  $: title =
    activeView === 'tools'
      ? 'Tools'
      : activeView === 'store'
        ? 'Tool Store'
      : activeView === 'details'
        ? 'Tool Details'
        : activeView === 'runtime'
          ? selectedTool?.name || 'Tool Runtime'
          : activeView === 'settings'
            ? 'Settings'
            : activeView === 'logs'
              ? 'Logs'
              : activeView === 'about'
                ? 'About'
                : 'Home';

  onMount(() => {
    appSettings = loadSettings();
    loadInitialData();
  });

  async function loadInitialData() {
    loading = true;
    try {
      const [toolList, logList, directory] = await Promise.all([
        api.listTools(),
        api.getLogs(),
        api.getToolsDirectory()
      ]);
      tools = toolList;
      logs = logList;
      toolsDirectory = directory;
    } finally {
      loading = false;
    }
  }

  async function refreshTools() {
    loading = true;
    try {
      tools = await api.refreshTools();
      if (activeView === 'store') {
        storeTools = await api.listToolStore();
        selectedStoreTool = selectedStoreTool ? storeTools.find((tool) => tool.url === selectedStoreTool?.url) || selectedStoreTool : null;
      }
      logs = await api.getLogs();
      showNotice('Tool list refreshed', 'success');
    } catch (error) {
      showNotice(errorMessage(error), 'error');
    } finally {
      loading = false;
    }
  }

  async function installTool() {
    try {
      const filePath = await api.selectToolPackage();
      if (!filePath) {
        showNotice('No package selected', 'info');
        return;
      }

      const preview = await api.previewToolPackage(filePath);
      tools = await api.refreshTools();
      const existing = tools.find((tool) => tool.id === preview.id);
      let result: ToolValidationResult;
      if (existing) {
        if (appSettings.confirmToolUpdates) {
          const confirmed = window.confirm(
            `${preview.name || preview.id} is already in your library. Do you want to update this tool?`
          );
          if (!confirmed) {
            showNotice('Update cancelled', 'info');
            return;
          }
        }
        result = await api.updateTool(filePath);
      } else {
        result = await api.installTool(filePath);
      }

      if (!result.valid) {
        showNotice(result.message || 'Validation Failed', 'error');
      } else {
        showNotice(result.message || (existing ? 'Tool Updated Successfully' : 'Tool Installed Successfully'), 'success');
      }
      activeView = 'tools';
      await refreshAfterMutation();
    } catch (error) {
      showNotice(errorMessage(error), 'error');
      await refreshAfterMutation();
    }
  }

  async function openTool(tool: ToolPackage) {
    try {
      selectedTool = await api.openTool(toolKey(tool));
      activeView = 'runtime';
      logs = await api.getLogs();
    } catch (error) {
      selectedTool = tool;
      activeView = 'details';
      showNotice(errorMessage(error) || 'Package Invalid', 'error');
    }
  }

  async function showDetails(tool: ToolPackage) {
    try {
      selectedTool = await api.getToolDetails(toolKey(tool));
    } catch {
      selectedTool = tool;
    }
    activeView = 'details';
  }

  async function removeTool(tool: ToolPackage) {
    const confirmed = window.confirm(`Remove ${tool.name || tool.id}?`);
    if (!confirmed) return;

    try {
      await api.removeTool(toolKey(tool));
      if (selectedTool && toolKey(selectedTool) === toolKey(tool)) {
        selectedTool = null;
        activeView = 'tools';
      }
      showNotice('Tool Removed Successfully', 'success');
      await refreshAfterMutation();
    } catch (error) {
      showNotice(errorMessage(error), 'error');
    }
  }

  async function openToolStore() {
    activeView = 'store';
    if (storeTools.length === 0) {
      await loadToolStore();
    }
  }

  async function loadToolStore() {
    storeLoading = true;
    try {
      storeTools = await api.listToolStore();
      selectedStoreTool = selectedStoreTool ? storeTools.find((tool) => tool.url === selectedStoreTool?.url) || null : null;
      showNotice('Tool Store refreshed', 'success');
    } catch (error) {
      showNotice(errorMessage(error), 'error');
    } finally {
      storeLoading = false;
    }
  }

  async function installStoreTool(tool: ToolStorePackage) {
    storeBusyURL = tool.url;
    try {
      const result = await api.installStoreTool(tool.url);
      if (!result.valid) {
        showNotice(result.message || 'Validation Failed', 'error');
      } else {
        showNotice(result.message || 'Tool Installed Successfully', 'success');
      }
      await refreshAfterMutation();
      await loadToolStore();
    } catch (error) {
      showNotice(errorMessage(error), 'error');
    } finally {
      storeBusyURL = '';
    }
  }

  async function updateStoreTool(tool: ToolStorePackage) {
    storeBusyURL = tool.url;
    try {
      const result = await api.updateStoreTool(tool.url);
      if (!result.valid) {
        showNotice(result.message || 'Validation Failed', 'error');
      } else {
        showNotice(result.message || 'Tool Updated Successfully', 'success');
      }
      await refreshAfterMutation();
      await loadToolStore();
    } catch (error) {
      showNotice(errorMessage(error), 'error');
    } finally {
      storeBusyURL = '';
    }
  }

  function showStoreDetails(tool: ToolStorePackage) {
    selectedStoreTool = tool;
  }

  async function runToolAction(action: string) {
    if (!selectedTool) return;
    try {
      await api.runToolAction(toolKey(selectedTool), action);
      logs = await api.getLogs();
      showNotice('Action logged', 'success');
    } catch (error) {
      showNotice(errorMessage(error), 'error');
    }
  }

  async function refreshLogs() {
    logs = await api.getLogs();
  }

  async function refreshAfterMutation() {
    tools = await api.refreshTools();
    logs = await api.getLogs();
  }

  function navigate(view: ViewName) {
    activeView = view;
  }

  async function visitSite() {
    try {
      await api.openExternalURL(projectSiteURL);
    } catch (error) {
      showNotice(errorMessage(error), 'error');
    }
  }

  function defaultSettings(): AppSettings {
    return {
      confirmToolUpdates: true,
      showInvalidTools: true,
      compactToolCards: false,
      reduceMotion: false
    };
  }

  function loadSettings(): AppSettings {
    try {
      const raw = window.localStorage.getItem(settingsStorageKey);
      if (!raw) return defaultSettings();
      return { ...defaultSettings(), ...JSON.parse(raw) };
    } catch {
      return defaultSettings();
    }
  }

  function saveSettings(next: AppSettings) {
    appSettings = next;
    window.localStorage.setItem(settingsStorageKey, JSON.stringify(next));
  }

  function updateSetting(key: keyof AppSettings, value: boolean) {
    saveSettings({ ...appSettings, [key]: value });
  }

  function resetSettings() {
    saveSettings(defaultSettings());
    showNotice('Settings reset', 'success');
  }

  function toolKey(tool: ToolPackage) {
    return tool.registry_key || tool.id;
  }

  function matchesToolSearch(tool: ToolPackage, query: string) {
    return [
      tool.name,
      tool.id,
      tool.author,
      tool.description,
      tool.manifest?.author,
      tool.manifest?.description
    ]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(query));
  }

  function matchesStoreToolSearch(tool: ToolStorePackage, query: string) {
    return [tool.name, tool.id, tool.author, tool.description]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(query));
  }

  function showNotice(message: string, tone: 'info' | 'error' | 'success' = 'info') {
    notice = message;
    noticeTone = tone;
    window.setTimeout(() => {
      if (notice === message) notice = '';
    }, 3200);
  }

  function errorMessage(error: unknown) {
    if (error instanceof Error) return error.message;
    if (typeof error === 'string') return error;
    return 'Operation failed';
  }
</script>

<main class="app-shell" class:compact-tools={appSettings.compactToolCards} class:reduce-motion={appSettings.reduceMotion}>
  <Sidebar activeView={sidebarView} toolCount={tools.length} on:navigate={(event) => navigate(event.detail)} />

  <section class="workspace">
    <header class="topbar">
      <div>
        <h1>{title}</h1>
      </div>

      <div class="actions">
        <button class="secondary" type="button" on:click={visitSite} title="Visit Site">
          <ExternalLink size={16} />
          <span>Visit Site</span>
        </button>
        <button class="secondary" type="button" on:click={refreshTools} title="Refresh">
          <span class:spin={loading || storeLoading} class="refresh-icon">
            <RefreshCw size={16} />
          </span>
          <span>Refresh</span>
        </button>
        <button class="secondary" type="button" on:click={openToolStore} title="Tool Store">
          <ShoppingBag size={16} />
          <span>Tool Store</span>
        </button>
        <button class="primary" type="button" on:click={installTool} title="Install Tool">
          <Download size={16} />
          <span>Install Tool</span>
        </button>
      </div>
    </header>

    {#if notice}
      <div class:toast-error={noticeTone === 'error'} class:toast-success={noticeTone === 'success'} class="toast">
        {notice}
      </div>
    {/if}

    <section class="content">
      {#if activeView === 'tools'}
        <section class="tools-view">
          <div class="tools-bar">
            <label class="search-box" aria-label="Search tools">
              <Search size={18} />
              <input bind:value={searchQuery} type="search" placeholder="Search tools..." />
            </label>
            <span class="tool-count">{filteredTools.length} of {visibleTools.length} tools</span>
          </div>

          <ToolGrid
            tools={filteredTools}
            emptyTitle={tools.length === 0 ? 'No tools installed' : visibleTools.length === 0 ? 'Invalid tools hidden' : 'No matching tools'}
            emptyDescription={tools.length === 0 ? 'Install a .tl package to populate this launcher.' : visibleTools.length === 0 ? 'Enable "Show invalid packages" in Settings to review invalid tools.' : 'Adjust the search query to show more tools.'}
            on:open={(event) => openTool(event.detail)}
            on:details={(event) => showDetails(event.detail)}
            on:remove={(event) => removeTool(event.detail)}
          />
        </section>
      {:else if activeView === 'store'}
        <section class="store-view">
          <div class="tools-bar">
            <label class="search-box" aria-label="Search store tools">
              <Search size={18} />
              <input bind:value={storeSearchQuery} type="search" placeholder="Search tools..." />
            </label>
            <div class="store-tools-summary">
              {#if storeLoading}
                <span>Loading Tool Store...</span>
              {:else}
                <span>{filteredStoreTools.length} of {storeTools.length} store tools</span>
              {/if}
              <button type="button" on:click={loadToolStore} disabled={storeLoading}>Refresh Store</button>
            </div>
          </div>

          <ToolStoreGrid
            tools={filteredStoreTools}
            busy={Boolean(storeBusyURL)}
            emptyTitle={storeTools.length === 0 ? 'No store tools loaded' : 'No matching store tools'}
            emptyDescription={storeTools.length === 0 ? 'Refresh the Tool Store to load packages from the remote tools.json file.' : 'Adjust the search query to show more tools.'}
            on:install={(event) => installStoreTool(event.detail)}
            on:update={(event) => updateStoreTool(event.detail)}
            on:details={(event) => showStoreDetails(event.detail)}
          />

          {#if selectedStoreTool}
            <ToolStoreDetails
              tool={selectedStoreTool}
              busy={storeBusyURL === selectedStoreTool.url}
              on:close={() => { selectedStoreTool = null; }}
              on:install={(event) => installStoreTool(event.detail)}
              on:update={(event) => updateStoreTool(event.detail)}
            />
          {/if}
        </section>
      {:else if activeView === 'home'}
        <section class="home-view">
          <div class="panel intro-panel">
            <h2>GRTBox</h2>
            <p>Desktop toolbox launcher and modular runtime for open `.tl` packages.</p>
          </div>
          <div class="metric-grid">
            <div class="panel metric">
              <span>{tools.length}</span>
              <p>Total Tools</p>
            </div>
            <div class="panel metric">
              <span>{validTools}</span>
              <p>Valid</p>
            </div>
            <div class="panel metric">
              <span>{invalidTools}</span>
              <p>Invalid</p>
            </div>
          </div>
        </section>
      {:else if activeView === 'details'}
        <ToolDetails tool={selectedTool} on:open={(event) => openTool(event.detail)} on:remove={(event) => removeTool(event.detail)} />
      {:else if activeView === 'runtime'}
        <ToolRunner tool={selectedTool} on:close={() => { activeView = 'tools'; selectedTool = null; }} />
      {:else if activeView === 'settings'}
        <section class="settings-view">
          <header class="settings-header">
            <h2>Settings</h2>
            <p>Adjust how the launcher behaves on this computer.</p>
          </header>

          <label class="setting-toggle">
            <input
              type="checkbox"
              checked={appSettings.confirmToolUpdates}
              on:change={(event) => updateSetting('confirmToolUpdates', event.currentTarget.checked)}
            />
            <span>
              <strong>Confirm tool updates</strong>
              <small>Ask before replacing an installed `.tl` package with a package that has the same tool ID.</small>
            </span>
          </label>

          <label class="setting-toggle">
            <input
              type="checkbox"
              checked={appSettings.showInvalidTools}
              on:change={(event) => updateSetting('showInvalidTools', event.currentTarget.checked)}
            />
            <span>
              <strong>Show invalid packages</strong>
              <small>Keep broken or incompatible packages visible in the Tools grid for inspection and removal.</small>
            </span>
          </label>

          <label class="setting-toggle">
            <input
              type="checkbox"
              checked={appSettings.compactToolCards}
              on:change={(event) => updateSetting('compactToolCards', event.currentTarget.checked)}
            />
            <span>
              <strong>Compact launcher cards</strong>
              <small>Use tighter tool cards when you want to scan more installed tools at once.</small>
            </span>
          </label>

          <label class="setting-toggle">
            <input
              type="checkbox"
              checked={appSettings.reduceMotion}
              on:change={(event) => updateSetting('reduceMotion', event.currentTarget.checked)}
            />
            <span>
              <strong>Reduce motion</strong>
              <small>Disable hover movement and spinner animation in the launcher UI.</small>
            </span>
          </label>

          <div class="settings-actions">
            <button type="button" on:click={resetSettings}>Reset Settings</button>
          </div>
        </section>
      {:else if activeView === 'about'}
        <section class="about-view panel">
          <h2>About GRTBox</h2>
          <p>GRTBox is a Windows-focused desktop application built with Go, Wails and Svelte. It installs `.tl` packages and runs TC modules through generic desktop primitives.</p>
          <dl>
            <div>
              <dt>Current Version</dt>
              <dd>{appVersion}</dd>
            </div>
            <div>
              <dt>Package Type</dt>
              <dd>ZIP archive renamed to `.tl`</dd>
            </div>
            <div>
              <dt>Default Icon</dt>
              <dd>Default Tool Icon</dd>
            </div>
            <div>
              <dt>Tools Directory</dt>
              <dd>{toolsDirectory}</dd>
            </div>
            <div>
              <dt>Runtime Mode</dt>
              <dd>TC executable modules</dd>
            </div>
            <div>
              <dt>Administrator Mode</dt>
              <dd>Required on Windows startup</dd>
            </div>
            <div>
              <dt>Project Site</dt>
              <dd>{projectSiteURL}</dd>
            </div>
          </dl>
        </section>
      {:else}
        <LogPanel {logs} on:refresh={refreshLogs} />
      {/if}
    </section>
  </section>
</main>

<style>
  .app-shell {
    width: 100%;
    height: 100%;
    color: #f2f2f2;
    background: #2d2d2d;
  }

  .workspace {
    min-width: 0;
    display: grid;
    grid-template-rows: auto auto minmax(0, 1fr);
    height: 100%;
    margin-left: 232px;
    overflow: hidden;
    background: #2d2d2d;
  }

  .topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 20px;
    min-height: 76px;
    padding: 18px 28px;
    border-bottom: 1px solid #444444;
    background: #2d2d2d;
  }

  h1 {
    margin: 0;
    color: #ffffff;
    font-size: 1.45rem;
    line-height: 1.15;
    letter-spacing: 0;
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 10px;
  }

  .actions button {
    min-height: 38px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 0 14px;
    border: 1px solid #505050;
    border-radius: 8px;
    cursor: pointer;
    font-weight: 700;
    transition:
      background 120ms ease,
      border-color 120ms ease;
  }

  .primary {
    color: #ffffff;
    background: #3f3f3f;
  }

  .primary:hover,
  .secondary:hover {
    border-color: #666666;
    background: #464646;
  }

  .secondary {
    color: #eeeeee;
    background: #343434;
  }

  .refresh-icon {
    width: 16px;
    height: 16px;
    display: inline-grid;
    place-items: center;
  }

  .refresh-icon.spin {
    animation: spin 900ms linear infinite;
  }

  .toast {
    align-self: start;
    justify-self: end;
    margin: 14px 28px 0;
    max-width: min(520px, calc(100vw - 320px));
    padding: 10px 12px;
    border: 1px solid #555555;
    border-radius: 8px;
    color: #e7e7e7;
    background: #383838;
    font-weight: 700;
  }

  .toast-error {
    border-color: #7b4a4a;
    color: #ffd6d6;
    background: #3a2e2e;
  }

  .toast-success {
    border-color: #4d6f61;
    color: #d8f4e5;
    background: #2f3a35;
  }

  .content {
    min-height: 0;
    overflow: auto;
    padding: 24px 28px 30px;
    background: #2d2d2d;
  }

  .tools-view,
  .store-view,
  .home-view {
    display: grid;
    gap: 22px;
  }

  .tools-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }

  .search-box {
    width: min(620px, 100%);
    min-height: 46px;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 0 14px;
    border: 1px solid #444444;
    border-radius: 8px;
    color: #b8b8b8;
    background: #252525;
  }

  .search-box:focus-within {
    border-color: #666666;
    background: #282828;
  }

  input {
    width: 100%;
    min-width: 0;
    border: 0;
    outline: 0;
    color: #ffffff;
    background: transparent;
  }

  input::placeholder {
    color: #8f8f8f;
  }

  .tool-count {
    flex: 0 0 auto;
    color: #b9b9b9;
    font-size: 0.88rem;
    font-weight: 700;
  }

  .store-tools-summary {
    flex: 0 0 auto;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: flex-end;
    gap: 10px;
    color: #b9b9b9;
    font-size: 0.88rem;
    font-weight: 700;
  }

  .store-tools-summary button {
    min-height: 36px;
    padding: 0 13px;
    border: 1px solid #505050;
    border-radius: 8px;
    color: #eeeeee;
    background: #343434;
    cursor: pointer;
    font-weight: 800;
  }

  .store-tools-summary button:hover:not(:disabled) {
    border-color: #666666;
    background: #414141;
  }

  .store-tools-summary button:disabled {
    cursor: not-allowed;
    opacity: 0.52;
  }

  .panel,
  .settings-view {
    max-width: 920px;
    border: 1px solid #444444;
    border-radius: 10px;
    background: #333333;
  }

  .settings-view {
    overflow: hidden;
  }

  .settings-header {
    padding: 20px;
    border-bottom: 1px solid #444444;
  }

  .settings-header h2 {
    margin: 0 0 8px;
    color: #ffffff;
    font-size: 1.15rem;
    letter-spacing: 0;
  }

  .settings-header p {
    margin: 0;
    color: #c9c9c9;
    line-height: 1.5;
  }

  .setting-toggle {
    display: grid;
    grid-template-columns: 22px minmax(0, 1fr);
    gap: 14px;
    padding: 16px 20px;
    border-bottom: 1px solid #444444;
    cursor: pointer;
  }

  .setting-toggle:hover {
    background: #363636;
  }

  .setting-toggle input {
    width: 18px;
    height: 18px;
    margin: 2px 0 0;
    accent-color: #d8d8d8;
    cursor: pointer;
  }

  .setting-toggle span,
  .setting-toggle strong,
  .setting-toggle small {
    display: block;
    min-width: 0;
  }

  .setting-toggle strong {
    color: #ffffff;
    font-size: 0.95rem;
  }

  .setting-toggle small {
    margin-top: 5px;
    color: #b9b9b9;
    line-height: 1.45;
  }

  .settings-actions {
    padding: 16px 20px;
  }

  .settings-actions button {
    min-height: 36px;
    padding: 0 13px;
    border: 1px solid #505050;
    border-radius: 8px;
    color: #eeeeee;
    background: #2b2b2b;
    cursor: pointer;
    font-weight: 800;
  }

  .settings-actions button:hover {
    border-color: #666666;
    background: #414141;
  }

  .intro-panel,
  .about-view {
    padding: 20px;
  }

  .intro-panel h2,
  .about-view h2 {
    margin: 0 0 8px;
    color: #ffffff;
    font-size: 1.15rem;
    letter-spacing: 0;
  }

  .intro-panel p,
  .about-view p {
    max-width: 760px;
    margin: 0;
    color: #c9c9c9;
    line-height: 1.55;
  }

  .metric-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 14px;
    max-width: 920px;
  }

  .metric {
    min-height: 96px;
    padding: 18px;
  }

  .metric span {
    display: block;
    color: #ffffff;
    font-size: 1.65rem;
    font-weight: 800;
  }

  .metric p {
    margin: 6px 0 0;
    color: #b9b9b9;
    font-size: 0.88rem;
    font-weight: 700;
  }

  dl {
    display: grid;
    margin: 18px 0 0;
    border-top: 1px solid #444444;
  }

  dl > div {
    display: grid;
    grid-template-columns: 180px minmax(0, 1fr);
    gap: 16px;
    padding: 13px 0;
    border-bottom: 1px solid #444444;
  }

  dt {
    color: #b9b9b9;
    font-weight: 800;
  }

  dd {
    min-width: 0;
    margin: 0;
    color: #f0f0f0;
    overflow-wrap: anywhere;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }

  :global(.compact-tools .tool-grid) {
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 14px;
  }

  :global(.compact-tools .tool-card) {
    min-height: 250px;
    gap: 12px;
    padding: 16px;
  }

  :global(.compact-tools .tool-card img) {
    width: 64px;
    height: 64px;
  }

  :global(.compact-tools .tool-card .description) {
    min-height: 48px;
    margin-top: 10px;
    -webkit-line-clamp: 2;
    line-clamp: 2;
  }

  :global(.reduce-motion *),
  :global(.reduce-motion *::before),
  :global(.reduce-motion *::after) {
    animation: none !important;
    transition: none !important;
  }

  @media (max-width: 820px) {
    .workspace {
      margin-left: 0;
      padding-top: 62px;
    }

    .topbar {
      align-items: flex-start;
      flex-direction: column;
      padding: 18px;
    }

    .actions {
      width: 100%;
      justify-content: flex-start;
    }

    .content {
      padding: 18px;
    }

    .tools-bar {
      align-items: stretch;
      flex-direction: column;
    }

    .tool-count {
      align-self: flex-start;
    }

    .store-tools-summary {
      align-self: flex-start;
      justify-content: flex-start;
    }

    .metric-grid,
    dl > div {
      grid-template-columns: 1fr;
    }

    .toast {
      justify-self: stretch;
      max-width: none;
      margin: 12px 18px 0;
    }
  }
</style>
