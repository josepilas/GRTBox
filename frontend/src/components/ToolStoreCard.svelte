<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { ToolStorePackage } from '../lib/types';

  export let tool: ToolStorePackage;
  export let busy = false;

  const dispatch = createEventDispatcher<{
    install: ToolStorePackage;
    update: ToolStorePackage;
    details: ToolStorePackage;
  }>();

  $: valid = tool.validation?.valid;
  $: canInstall = valid && !tool.installed;
  $: canUpdate = valid && tool.installed && tool.update_available;
</script>

<article class="store-card" class:invalid={!valid} class:installed={tool.installed}>
  <div class="store-main">
    <img src={tool.icon_data} alt={tool.icon_name || 'Default Tool Icon'} />
    <div class="store-copy">
      <div class="store-heading">
        <h2>{tool.name || tool.id || 'Unnamed Tool'}</h2>
        <span>Version {tool.version || 'Unknown'}</span>
      </div>
      <p>{tool.description || 'No description provided.'}</p>
      <dl>
        <div>
          <dt>ID</dt>
          <dd>{tool.id || 'Unknown'}</dd>
        </div>
        <div>
          <dt>Author</dt>
          <dd>{tool.author || 'Unknown'}</dd>
        </div>
      </dl>
    </div>
  </div>

  <footer>
    <div class="badges">
      {#if tool.installed}
        <span class="badge">Installed{tool.installed_version ? ` ${tool.installed_version}` : ''}</span>
      {/if}
      {#if tool.update_available}
        <span class="badge update">Update Available</span>
      {/if}
      {#if !valid}
        <span class="badge invalid-badge">Invalid Package</span>
      {/if}
    </div>
    <div class="actions">
      {#if canUpdate}
        <button type="button" disabled={busy} on:click={() => dispatch('update', tool)}>Update</button>
      {:else if canInstall}
        <button type="button" disabled={busy} on:click={() => dispatch('install', tool)}>Install</button>
      {:else if tool.installed}
        <button type="button" disabled>Installed</button>
      {:else}
        <button type="button" disabled>Install</button>
      {/if}
      <button type="button" class="secondary" on:click={() => dispatch('details', tool)}>Details</button>
    </div>
  </footer>
</article>

<style>
  .store-card {
    min-height: 330px;
    display: grid;
    grid-template-rows: minmax(0, 1fr) auto;
    gap: 18px;
    padding: 22px;
    border: 1px solid #444444;
    border-radius: 12px;
    background: #333333;
    transition:
      border-color 120ms ease,
      background 120ms ease,
      transform 120ms ease;
  }

  .store-card:hover {
    border-color: #666666;
    background: #383838;
    transform: translateY(-2px);
  }

  .store-card.invalid {
    border-color: #674b4b;
  }

  .store-card.installed {
    border-color: #4c5c55;
  }

  .store-main {
    display: grid;
    align-content: start;
    gap: 18px;
    min-width: 0;
  }

  img {
    width: 86px;
    height: 86px;
    object-fit: cover;
    border: 1px solid #4d4d4d;
    border-radius: 12px;
    background: #252525;
  }

  .store-copy,
  .store-heading,
  dl,
  dd {
    min-width: 0;
  }

  h2 {
    margin: 0;
    color: #ffffff;
    font-size: 1.26rem;
    line-height: 1.25;
    letter-spacing: 0;
    overflow-wrap: anywhere;
  }

  .store-heading span {
    display: block;
    margin-top: 8px;
    color: #b9b9b9;
    font-size: 0.86rem;
    font-weight: 800;
  }

  p {
    display: -webkit-box;
    min-height: 72px;
    margin: 15px 0 0;
    overflow: hidden;
    color: #d0d0d0;
    font-size: 0.95rem;
    line-height: 1.55;
    line-clamp: 3;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 3;
  }

  dl {
    display: grid;
    gap: 6px;
    margin: 16px 0 0;
  }

  dl div {
    display: grid;
    grid-template-columns: 66px minmax(0, 1fr);
    gap: 10px;
  }

  dt {
    color: #a9a9a9;
    font-size: 0.78rem;
    font-weight: 800;
  }

  dd {
    margin: 0;
    color: #eeeeee;
    font-size: 0.82rem;
    font-weight: 700;
    overflow-wrap: anywhere;
  }

  footer {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 14px;
  }

  .badges,
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .badge {
    padding: 6px 9px;
    border: 1px solid #4d6f61;
    border-radius: 999px;
    color: #d8f4e5;
    background: #2f3a35;
    font-size: 0.72rem;
    font-weight: 800;
  }

  .badge.update {
    border-color: #6f654d;
    color: #ffe8b0;
    background: #3a3528;
  }

  .badge.invalid-badge {
    border-color: #7b4a4a;
    color: #ffd6d6;
    background: #3a2e2e;
  }

  button {
    min-width: 76px;
    min-height: 34px;
    padding: 0 11px;
    border: 1px solid #505050;
    border-radius: 8px;
    color: #ffffff;
    background: #3f3f3f;
    cursor: pointer;
    font-size: 0.78rem;
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

  @media (max-width: 520px) {
    footer {
      align-items: stretch;
      flex-direction: column;
    }
  }
</style>
