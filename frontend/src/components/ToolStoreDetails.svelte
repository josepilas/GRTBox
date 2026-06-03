<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { ToolStorePackage } from '../lib/types';

  export let tool: ToolStorePackage | null = null;
  export let busy = false;

  const dispatch = createEventDispatcher<{
    close: void;
    install: ToolStorePackage;
    update: ToolStorePackage;
  }>();

  $: manifestJSON = tool?.manifest ? JSON.stringify(tool.manifest, null, 2) : '{}';
  $: valid = Boolean(tool?.validation?.valid);
  $: canInstall = Boolean(tool && valid && !tool.installed);
  $: canUpdate = Boolean(tool && valid && tool.installed && tool.update_available);
</script>

{#if tool}
  <section class="store-details">
    <header>
      <img src={tool.icon_data} alt={tool.icon_name || 'Default Tool Icon'} />
      <div>
        <h2>{tool.name || tool.id || 'Store Tool'}</h2>
        <p>{tool.url}</p>
      </div>
      <div class="header-actions">
        {#if canUpdate}
          <button type="button" disabled={busy} on:click={() => dispatch('update', tool)}>Update</button>
        {:else if canInstall}
          <button type="button" disabled={busy} on:click={() => dispatch('install', tool)}>Install</button>
        {/if}
        <button type="button" class="secondary" on:click={() => dispatch('close')}>Close</button>
      </div>
    </header>

    <div class="detail-list">
      <div><span>Name</span><strong>{tool.name || 'Unknown'}</strong></div>
      <div><span>ID</span><strong>{tool.id || 'Unknown'}</strong></div>
      <div><span>Version</span><strong>{tool.version || 'Unknown'}</strong></div>
      <div><span>Installed Version</span><strong>{tool.installed_version || 'Not installed'}</strong></div>
      <div><span>Author</span><strong>{tool.author || 'Unknown'}</strong></div>
      <div><span>Description</span><strong>{tool.description || 'No description provided.'}</strong></div>
      <div><span>Runtime</span><strong>{tool.runtime || 'Unknown'}</strong></div>
      <div><span>Entry File</span><strong>{tool.entry || 'Unknown'}</strong></div>
      <div><span>Package URL</span><strong>{tool.url}</strong></div>
    </div>

    {#if tool.validation.errors.length || tool.validation.warnings.length}
      <section class="validation-panel">
        {#each tool.validation.errors as error}
          <p class="error">{error}</p>
        {/each}
        {#each tool.validation.warnings as warning}
          <p class="warning">{warning}</p>
        {/each}
      </section>
    {/if}

    <section class="manifest-panel">
      <h3>Full Manifest</h3>
      <pre>{manifestJSON}</pre>
    </section>
  </section>
{/if}

<style>
  .store-details {
    max-width: 980px;
    border: 1px solid #444444;
    border-radius: 10px;
    background: #333333;
    overflow: hidden;
  }

  header {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto;
    gap: 14px;
    align-items: center;
    padding: 18px;
    border-bottom: 1px solid #444444;
  }

  img {
    width: 64px;
    height: 64px;
    object-fit: cover;
    border: 1px solid #4d4d4d;
    border-radius: 8px;
    background: #252525;
  }

  h2,
  h3 {
    margin: 0;
    color: #ffffff;
    letter-spacing: 0;
  }

  h2 {
    font-size: 1.25rem;
    overflow-wrap: anywhere;
  }

  header p {
    margin: 5px 0 0;
    color: #b9b9b9;
    font-size: 0.84rem;
    font-weight: 700;
    overflow-wrap: anywhere;
  }

  .header-actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 8px;
  }

  button {
    min-height: 36px;
    padding: 0 12px;
    border: 1px solid #505050;
    border-radius: 8px;
    color: #ffffff;
    background: #3f3f3f;
    cursor: pointer;
    font-weight: 800;
  }

  button.secondary {
    color: #eeeeee;
    background: #2b2b2b;
  }

  button:hover:not(:disabled) {
    border-color: #666666;
    background: #464646;
  }

  button:disabled {
    cursor: not-allowed;
    opacity: 0.52;
  }

  .detail-list {
    display: grid;
  }

  .detail-list > div {
    display: grid;
    grid-template-columns: 180px minmax(0, 1fr);
    gap: 16px;
    padding: 14px 18px;
    border-bottom: 1px solid #444444;
  }

  .detail-list span {
    color: #b9b9b9;
    font-weight: 800;
  }

  strong {
    min-width: 0;
    color: #eeeeee;
    overflow-wrap: anywhere;
  }

  .validation-panel,
  .manifest-panel {
    padding: 16px 18px 18px;
    border-top: 1px solid #444444;
  }

  .validation-panel p {
    margin: 0 0 8px;
    padding: 10px 12px;
    border-radius: 8px;
    font-weight: 700;
  }

  .error {
    color: #ffd6d6;
    background: #3a2e2e;
  }

  .warning {
    color: #ffe8b0;
    background: #3a3528;
  }

  .manifest-panel h3 {
    margin-bottom: 10px;
    font-size: 0.95rem;
  }

  pre {
    max-height: 360px;
    margin: 0;
    overflow: auto;
    padding: 14px;
    border: 1px solid #444444;
    border-radius: 8px;
    color: #e7e7e7;
    background: #252525;
    font-size: 0.82rem;
    line-height: 1.5;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
  }

  @media (max-width: 720px) {
    header,
    .detail-list > div {
      grid-template-columns: 1fr;
    }

    .header-actions {
      justify-content: flex-start;
    }
  }
</style>
