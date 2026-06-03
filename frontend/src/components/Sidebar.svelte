<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import brandLogo from '../assets/grtbox-brand.png';
  import type { ViewName } from '../lib/types';

  export let activeView: ViewName;
  export let toolCount = 0;

  const dispatch = createEventDispatcher<{ navigate: ViewName }>();

  const navItems: Array<{ id: ViewName; label: string }> = [
    { id: 'home', label: 'Home' },
    { id: 'tools', label: 'Tools' },
    { id: 'logs', label: 'Logs' },
    { id: 'settings', label: 'Settings' },
    { id: 'about', label: 'About' }
  ];
</script>

<aside class="sidebar">
  <div class="brand">
    <div class="brand-mark">
      <img src={brandLogo} alt="GRTBox" />
    </div>
    <div>
      <strong>GRTBox</strong>
      <span>Toolbox Runtime</span>
    </div>
  </div>

  <nav aria-label="Main navigation">
    {#each navItems as item}
      <button
        type="button"
        class:active={activeView === item.id}
        title={item.label}
        on:click={() => dispatch('navigate', item.id)}
      >
        {item.label}
      </button>
    {/each}
  </nav>

  <div class="sidebar-footer">
    <span>Installed Tools</span>
    <strong>{toolCount}</strong>
  </div>
</aside>

<style>
  .sidebar {
    position: fixed;
    inset: 0 auto 0 0;
    z-index: 10;
    width: 232px;
    display: grid;
    grid-template-rows: auto minmax(0, 1fr) auto;
    padding: 20px 16px;
    color: #f0f0f0;
    border-right: 1px solid #444444;
    background: #252525;
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;
    padding: 0 0 26px;
  }

  .brand-mark {
    width: 42px;
    height: 42px;
    flex: 0 0 auto;
    overflow: hidden;
    border: 1px solid #4d4d4d;
    border-radius: 8px;
    background: #ffffff;
  }

  .brand-mark img {
    width: 100%;
    height: 100%;
    display: block;
    object-fit: contain;
  }

  .brand strong,
  .brand span {
    display: block;
    min-width: 0;
  }

  .brand strong {
    color: #ffffff;
    font-size: 1.05rem;
    letter-spacing: 0;
  }

  .brand span {
    margin-top: 2px;
    color: #a9a9a9;
    font-size: 0.78rem;
    font-weight: 700;
  }

  nav {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  nav button {
    width: 100%;
    min-height: 40px;
    padding: 0 12px;
    border: 1px solid transparent;
    border-radius: 8px;
    color: #d0d0d0;
    background: transparent;
    cursor: pointer;
    font-weight: 700;
    text-align: left;
  }

  nav button:hover,
  nav button.active {
    color: #ffffff;
    border-color: #444444;
    background: #333333;
  }

  .sidebar-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 12px;
    border: 1px solid #444444;
    border-radius: 8px;
    color: #d0d0d0;
    background: #2d2d2d;
  }

  .sidebar-footer span {
    font-size: 0.82rem;
    font-weight: 700;
  }

  .sidebar-footer strong {
    color: #ffffff;
    font-size: 1.05rem;
  }

  @media (max-width: 820px) {
    .sidebar {
      inset: 0 0 auto 0;
      width: 100%;
      min-height: 62px;
      grid-template-columns: auto minmax(0, 1fr) auto;
      grid-template-rows: auto;
      align-items: center;
      gap: 12px;
      padding: 10px 12px;
      border-right: 0;
      border-bottom: 1px solid #444444;
    }

    .brand {
      padding: 0;
    }

    .brand span,
    .sidebar-footer span {
      display: none;
    }

    nav {
      flex-direction: row;
      justify-content: center;
      overflow-x: auto;
    }

    nav button {
      width: auto;
      flex: 0 0 auto;
    }

    .sidebar-footer {
      min-width: 44px;
      justify-content: center;
      padding: 10px;
    }
  }
</style>
