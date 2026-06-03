<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { ToolPackage } from '../lib/types';

  export let tool: ToolPackage;

  const dispatch = createEventDispatcher<{
    open: ToolPackage;
    details: ToolPackage;
    remove: ToolPackage;
  }>();

  function activateCard() {
    if (tool.validation?.valid) {
      dispatch('open', tool);
      return;
    }
    dispatch('details', tool);
  }

</script>

<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<article
  class="tool-card"
  class:invalid={!tool.validation?.valid}
  title={tool.validation?.valid ? 'Open tool' : 'Package Invalid'}
  on:dblclick={activateCard}
>
  <div class="tile-main">
    <img src={tool.icon_data} alt={tool.icon_name || 'Default Tool Icon'} />
    <div class="tool-copy">
      <h2>{tool.name || tool.id || 'Unnamed Tool'}</h2>
      <p class="version">Version {tool.version || 'Unknown'}</p>
      <p class="description">{tool.description || 'No description provided.'}</p>
    </div>
  </div>

  <div class="tile-footer">
    <span class:status-ok={tool.validation?.valid} class:status-bad={!tool.validation?.valid} class="status">
      {tool.validation?.valid ? 'Valid' : 'Validation Failed'}
    </span>
    <div class="card-actions">
      <button type="button" title="Open" disabled={!tool.validation?.valid} on:click|stopPropagation={() => dispatch('open', tool)}>
        Open
      </button>
      <button type="button" title="Details" on:click|stopPropagation={() => dispatch('details', tool)}>Details</button>
      <button type="button" class="danger" title="Remove" on:click|stopPropagation={() => dispatch('remove', tool)}>Remove</button>
    </div>
  </div>
</article>

<style>
  .tool-card {
    min-height: 320px;
    display: grid;
    grid-template-rows: minmax(0, 1fr) auto;
    gap: 18px;
    padding: 22px;
    border: 1px solid #444444;
    border-radius: 12px;
    background: #333333;
    cursor: pointer;
    transition:
      border-color 120ms ease,
      background 120ms ease,
      transform 120ms ease;
  }

  .tool-card:hover {
    border-color: #666666;
    background: #383838;
    transform: translateY(-2px);
  }

  .tool-card.invalid {
    border-color: #5d4545;
  }

  .tile-main {
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

  .tool-copy {
    min-width: 0;
  }

  h2 {
    margin: 0;
    color: #ffffff;
    font-size: 1.28rem;
    line-height: 1.25;
    letter-spacing: 0;
    overflow-wrap: anywhere;
  }

  .version {
    margin: 8px 0 0;
    color: #b9b9b9;
    font-size: 0.86rem;
    font-weight: 800;
  }

  .description {
    display: -webkit-box;
    min-height: 74px;
    margin: 16px 0 0;
    overflow: hidden;
    color: #d0d0d0;
    font-size: 0.96rem;
    line-height: 1.55;
    line-clamp: 3;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 3;
  }

  .tile-footer {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 14px;
  }

  .status {
    max-width: 150px;
    padding: 6px 9px;
    border: 1px solid #505050;
    border-radius: 999px;
    color: #dcdcdc;
    background: #2d2d2d;
    font-size: 0.72rem;
    font-weight: 800;
    overflow-wrap: anywhere;
  }

  .status-ok {
    border-color: #4d6f61;
    color: #d8f4e5;
  }

  .status-bad {
    border-color: #7b4a4a;
    color: #ffd6d6;
  }

  .card-actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 8px;
  }

  button {
    min-width: 68px;
    min-height: 34px;
    padding: 0 10px;
    border: 1px solid #505050;
    border-radius: 8px;
    color: #eeeeee;
    background: #2b2b2b;
    cursor: pointer;
    font-size: 0.78rem;
    font-weight: 800;
  }

  button:hover {
    border-color: #666666;
    background: #414141;
  }

  button.danger {
    color: #ffd6d6;
  }

  button:disabled {
    cursor: not-allowed;
    opacity: 0.48;
  }

  @media (max-width: 460px) {
    .tile-footer {
      align-items: stretch;
      flex-direction: column;
    }

    .card-actions {
      justify-content: flex-start;
    }
  }
</style>
