<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { ExternalLink, Trash2 } from '@lucide/svelte';
  import type { ToolPackage } from '../lib/types';

  export let tool: ToolPackage | null = null;

  const dispatch = createEventDispatcher<{
    open: ToolPackage;
    remove: ToolPackage;
  }>();

  $: permissions = tool?.manifest?.permissions || tool?.metadata?.permissions || [];
  $: platforms = tool?.manifest?.target_platforms || tool?.metadata?.target_platforms || [];
</script>

{#if !tool}
  <section class="empty-details">
    <strong>No tool selected</strong>
  </section>
{:else}
  <section class="details-shell">
    <header>
      <img src={tool.icon_data} alt={tool.icon_name || 'Default Tool Icon'} />
      <div>
        <h2>{tool.name || tool.id}</h2>
        <p>{tool.validation.valid ? 'Package Valid' : 'Package Invalid'}</p>
      </div>
      <div class="header-actions">
        <button type="button" class="open" title="Open" disabled={!tool.validation.valid} on:click={() => dispatch('open', tool)}>
          <ExternalLink size={16} />
          <span>Open</span>
        </button>
        <button type="button" class="danger" title="Remove" on:click={() => dispatch('remove', tool)}>
          <Trash2 size={16} />
          <span>Remove</span>
        </button>
      </div>
    </header>

    <div class="detail-list">
      <div><span>Name</span><strong>{tool.name || 'Unknown'}</strong></div>
      <div><span>ID</span><strong>{tool.id || 'Unknown'}</strong></div>
      <div><span>Version</span><strong>{tool.version || 'Unknown'}</strong></div>
      <div><span>Runtime</span><strong>{tool.runtime || tool.manifest?.runtime || 'Unknown'}</strong></div>
      <div><span>Package Format</span><strong>{tool.manifest?.package_format_version || tool.metadata?.package_format_version || '1.0.0'}</strong></div>
      <div><span>Requires Admin</span><strong>{tool.manifest?.requires_admin || tool.metadata?.requires_admin ? 'Yes' : 'No'}</strong></div>
      <div><span>Author</span><strong>{tool.author || 'Unknown'}</strong></div>
      <div><span>Description</span><strong>{tool.description || 'No description provided.'}</strong></div>
      <div><span>Entry File</span><strong>{tool.entry || 'main.tc'}</strong></div>
      <div>
        <span>Permissions</span>
        <strong class="pill-row">
          {#if permissions.length}
            {#each permissions as permission}
              <em>{permission}</em>
            {/each}
          {:else}
            None
          {/if}
        </strong>
      </div>
      <div>
        <span>Supported Platforms</span>
        <strong class="pill-row">
          {#if platforms.length}
            {#each platforms as platform}
              <em>{platform}</em>
            {/each}
          {:else}
            Any
          {/if}
        </strong>
      </div>
      <div><span>Package Location</span><strong>{tool.location}</strong></div>
      <div><span>Icon</span><strong>{tool.icon_name || 'Default Tool Icon'}</strong></div>
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
  </section>
{/if}

<style>
  .details-shell,
  .empty-details {
    max-width: 940px;
    border: 1px solid #444444;
    border-radius: 10px;
    background: #333333;
  }

  .empty-details {
    min-height: 220px;
    display: grid;
    place-items: center;
    color: #b9b9b9;
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
    border-radius: 8px;
    border: 1px solid #4d4d4d;
    background: #252525;
  }

  h2 {
    margin: 0;
    color: #ffffff;
    font-size: 1.3rem;
    letter-spacing: 0;
    overflow-wrap: anywhere;
  }

  header p {
    margin: 4px 0 0;
    color: #b9b9b9;
    font-weight: 800;
  }

  .header-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    justify-content: flex-end;
  }

  button {
    min-height: 38px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 7px;
    padding: 0 12px;
    border-radius: 8px;
    border: 1px solid #505050;
    cursor: pointer;
    font-weight: 800;
  }

  button.open {
    color: #ffffff;
    background: #3f3f3f;
  }

  button.danger {
    color: #ffd6d6;
    background: #2b2b2b;
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

  .detail-list > div:last-child {
    border-bottom: 0;
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

  .pill-row {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  em {
    padding: 4px 8px;
    border: 1px solid #505050;
    border-radius: 8px;
    color: #eeeeee;
    background: #2d2d2d;
    font-style: normal;
    font-size: 0.8rem;
  }

  .validation-panel {
    padding: 16px 18px 18px;
    border-top: 1px solid #444444;
  }

  .validation-panel p {
    margin: 0 0 8px;
    padding: 10px 12px;
    border-radius: 8px;
    font-weight: 700;
  }

  .validation-panel p:last-child {
    margin-bottom: 0;
  }

  .error {
    color: #ffd6d6;
    background: #3a2e2e;
  }

  .warning {
    color: #ffe8b0;
    background: #3a3528;
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
