<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import ToolStoreCard from './ToolStoreCard.svelte';
  import type { ToolStorePackage } from '../lib/types';

  export let tools: ToolStorePackage[] = [];
  export let busy = false;
  export let emptyTitle = 'No store tools found';
  export let emptyDescription = 'Refresh the Tool Store or adjust your search query.';

  const dispatch = createEventDispatcher<{
    install: ToolStorePackage;
    update: ToolStorePackage;
    details: ToolStorePackage;
  }>();
</script>

{#if tools.length === 0}
  <section class="empty-state">
    <strong>{emptyTitle}</strong>
    <span>{emptyDescription}</span>
  </section>
{:else}
  <section class="store-grid" aria-label="Tool Store">
    {#each tools as tool (tool.registry_key || tool.url)}
      <ToolStoreCard
        {tool}
        {busy}
        on:install={(event) => dispatch('install', event.detail)}
        on:update={(event) => dispatch('update', event.detail)}
        on:details={(event) => dispatch('details', event.detail)}
      />
    {/each}
  </section>
{/if}

<style>
  .store-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 22px;
  }

  .empty-state {
    min-height: 280px;
    display: grid;
    place-items: center;
    align-content: center;
    gap: 8px;
    border: 1px dashed #4d4d4d;
    border-radius: 10px;
    color: #a9a9a9;
    background: #333333;
    text-align: center;
  }

  .empty-state strong {
    color: #ffffff;
    font-size: 1rem;
  }

  .empty-state span {
    max-width: 360px;
    font-size: 0.9rem;
    line-height: 1.45;
  }
</style>
