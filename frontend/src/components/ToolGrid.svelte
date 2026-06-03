<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import ToolCard from './ToolCard.svelte';
  import type { ToolPackage } from '../lib/types';

  export let tools: ToolPackage[] = [];
  export let emptyTitle = 'No tools installed';
  export let emptyDescription = 'Install a .tl package to populate this launcher.';

  const dispatch = createEventDispatcher<{
    open: ToolPackage;
    details: ToolPackage;
    remove: ToolPackage;
  }>();
</script>

{#if tools.length === 0}
  <section class="empty-state">
    <strong>{emptyTitle}</strong>
    <span>{emptyDescription}</span>
  </section>
{:else}
  <section class="tool-grid" aria-label="Installed Tools">
    {#each tools as tool (tool.registry_key || tool.id || tool.location)}
      <ToolCard
        {tool}
        on:open={(event) => dispatch('open', event.detail)}
        on:details={(event) => dispatch('details', event.detail)}
        on:remove={(event) => dispatch('remove', event.detail)}
      />
    {/each}
  </section>
{/if}

<style>
  .tool-grid {
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
    max-width: 320px;
    font-size: 0.9rem;
    line-height: 1.45;
  }
</style>
