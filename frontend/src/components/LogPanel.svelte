<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { RefreshCw } from '@lucide/svelte';
  import type { LogEntry } from '../lib/types';

  export let logs: LogEntry[] = [];

  const dispatch = createEventDispatcher<{ refresh: void }>();
</script>

<section class="logs-shell">
  <header>
    <h2>Logs</h2>
    <button type="button" title="Refresh Logs" on:click={() => dispatch('refresh')}>
      <RefreshCw size={16} />
      <span>Refresh</span>
    </button>
  </header>

  {#if logs.length === 0}
    <div class="empty">No log entries</div>
  {:else}
    <ol>
      {#each logs.slice().reverse() as entry (entry.id)}
        <li>
          <span class:warn={entry.level === 'warn'} class:error={entry.level === 'error'}>{entry.level}</span>
          <time>{new Date(entry.timestamp).toLocaleString()}</time>
          <p>{entry.message}</p>
        </li>
      {/each}
    </ol>
  {/if}
</section>

<style>
  .logs-shell {
    max-width: 940px;
    border: 1px solid #444444;
    border-radius: 10px;
    background: #333333;
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 18px;
    border-bottom: 1px solid #444444;
  }

  h2 {
    margin: 0;
    color: #ffffff;
    font-size: 1.25rem;
    letter-spacing: 0;
  }

  button {
    min-height: 36px;
    display: inline-flex;
    align-items: center;
    gap: 7px;
    padding: 0 12px;
    border-radius: 8px;
    border: 1px solid #505050;
    color: #eeeeee;
    background: #2b2b2b;
    cursor: pointer;
    font-weight: 800;
  }

  .empty {
    min-height: 180px;
    display: grid;
    place-items: center;
    color: #b9b9b9;
    font-weight: 800;
  }

  ol {
    display: grid;
    gap: 0;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  li {
    display: grid;
    grid-template-columns: 70px 180px minmax(0, 1fr);
    gap: 12px;
    align-items: start;
    padding: 13px 18px;
    border-bottom: 1px solid #444444;
  }

  li:last-child {
    border-bottom: 0;
  }

  li > span {
    padding: 4px 8px;
    border-radius: 8px;
    border: 1px solid #4d6f61;
    color: #d8f4e5;
    background: #2f3a35;
    font-size: 0.72rem;
    font-weight: 900;
    text-transform: uppercase;
    text-align: center;
  }

  li > span.warn {
    border-color: #756235;
    color: #ffe8b0;
    background: #3a3528;
  }

  li > span.error {
    border-color: #7b4a4a;
    color: #ffd6d6;
    background: #3a2e2e;
  }

  time {
    color: #b9b9b9;
    font-size: 0.82rem;
    font-weight: 700;
  }

  p {
    min-width: 0;
    margin: 0;
    overflow-wrap: anywhere;
    color: #eeeeee;
  }

  @media (max-width: 720px) {
    li {
      grid-template-columns: 1fr;
    }

    li > span {
      width: fit-content;
    }
  }
</style>
