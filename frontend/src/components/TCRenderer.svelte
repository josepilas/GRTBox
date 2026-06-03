<script lang="ts">
  import type { TCNode } from '../lib/types';

  export let node: TCNode | string;

  $: normalizedChildren = normalizeChildren(typeof node === 'string' ? undefined : node.children);
  $: props = typeof node === 'string' ? {} : node.props || {};

  function normalizeChildren(children: TCNode['children'] | undefined): Array<TCNode | string> {
    if (children === undefined || children === null) return [];
    return Array.isArray(children) ? children : [children];
  }

  function stringProp(name: string) {
    const value = props[name];
    return typeof value === 'string' ? value : undefined;
  }

  function styleProp() {
    const value = props.style;
    return typeof value === 'string' ? value : undefined;
  }

  function boolProp(name: string) {
    const value = props[name];
    if (typeof value === 'boolean') return value;
    if (typeof value === 'string') {
      const normalized = value.trim().toLowerCase();
      if (normalized === '' || normalized === 'false' || normalized === '0' || normalized === 'no' || normalized === 'off') {
        return false;
      }
      return true;
    }
    if (typeof value === 'number') return value !== 0;
    return Boolean(value);
  }

  function handleClick() {
    const handler = props.onClick;
    if (typeof handler === 'function') handler();
  }

  function handleChange(event: Event) {
    const handler = props.onChange;
    if (typeof handler !== 'function') return;
    const target = event.target as HTMLSelectElement | HTMLInputElement | HTMLTextAreaElement;
    if (target instanceof HTMLInputElement && target.type === 'checkbox') {
      handler(target.checked, event);
      return;
    }
    handler(target.value, event);
  }

  function handleInput(event: Event) {
    const target = event.target as HTMLInputElement | HTMLTextAreaElement;
    const value = target.value;
    const inputHandler = props.onInput;
    if (typeof inputHandler === 'function') {
      inputHandler(value, event);
      return;
    }
    const changeHandler = props.onChange;
    if (typeof changeHandler === 'function') changeHandler(value, event);
  }
</script>

{#if typeof node === 'string'}
  {node}
{:else if node.type === 'div'}
  <div class={stringProp('className')} style={styleProp()}>
    {#each normalizedChildren as child}
      <svelte:self node={child} />
    {/each}
  </div>
{:else if node.type === 'section'}
  <section class={stringProp('className')} style={styleProp()}>
    {#each normalizedChildren as child}
      <svelte:self node={child} />
    {/each}
  </section>
{:else if node.type === 'h1'}
  <h1 class={stringProp('className')}>{#each normalizedChildren as child}<svelte:self node={child} />{/each}</h1>
{:else if node.type === 'h2'}
  <h2 class={stringProp('className')}>{#each normalizedChildren as child}<svelte:self node={child} />{/each}</h2>
{:else if node.type === 'h3'}
  <h3 class={stringProp('className')}>{#each normalizedChildren as child}<svelte:self node={child} />{/each}</h3>
{:else if node.type === 'p'}
  <p class={stringProp('className')}>{#each normalizedChildren as child}<svelte:self node={child} />{/each}</p>
{:else if node.type === 'button'}
  <button type="button" class={stringProp('className')} style={styleProp()} disabled={boolProp('disabled')} on:click={handleClick}>
    {#each normalizedChildren as child}<svelte:self node={child} />{/each}
  </button>
{:else if node.type === 'select'}
  <select class={stringProp('className')} value={stringProp('value')} disabled={boolProp('disabled')} on:change={handleChange}>
    {#each normalizedChildren as child}
      <svelte:self node={child} />
    {/each}
  </select>
{:else if node.type === 'input'}
  <input
    class={stringProp('className')}
    type={stringProp('type') || 'text'}
    value={stringProp('value') || ''}
    placeholder={stringProp('placeholder')}
    checked={boolProp('checked')}
    disabled={boolProp('disabled')}
    on:input={handleInput}
    on:change={handleChange}
  />
{:else if node.type === 'option'}
  <option value={stringProp('value')} selected={boolProp('selected')} disabled={boolProp('disabled')}>{#each normalizedChildren as child}<svelte:self node={child} />{/each}</option>
{:else if node.type === 'textarea'}
  <textarea
    class={stringProp('className')}
    value={stringProp('value') || ''}
    placeholder={stringProp('placeholder')}
    disabled={boolProp('disabled')}
    on:input={handleInput}
    on:change={handleChange}
  ></textarea>
{:else if node.type === 'pre'}
  <pre class={stringProp('className')}>{#each normalizedChildren as child}<svelte:self node={child} />{/each}</pre>
{:else if node.type === 'span'}
  <span class={stringProp('className')} style={styleProp()}>{#each normalizedChildren as child}<svelte:self node={child} />{/each}</span>
{:else if node.type === 'label'}
  <label class={stringProp('className')} for={stringProp('for')}>{#each normalizedChildren as child}<svelte:self node={child} />{/each}</label>
{:else}
  <div class="tc-unsupported">Unsupported TC UI element: {node.type}</div>
{/if}

<style>
  :global(.tc-app),
  :global(.tc-panel),
  section,
  div {
    min-width: 0;
  }

  :global(.tc-app) {
    display: grid;
    gap: 16px;
  }

  :global(.tc-panel),
  section {
    display: grid;
    gap: 12px;
    padding: 14px;
    border: 1px solid #444444;
    border-radius: 10px;
    background: #2d2d2d;
  }

  :global(.tc-row) {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 10px;
  }

  :global(.tc-actions),
  :global(.tc-tabs) {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  :global(.tc-green-button) {
    min-height: 48px;
    border-color: #4d6f61;
    color: #06170e;
    background: #7bd68f;
    font-size: 0.98rem;
  }

  :global(.tc-green-button:hover:not(:disabled)) {
    border-color: #73a982;
    background: #91e8a3;
  }

  :global(.tc-progress-panel),
  :global(.tc-progress-stack) {
    gap: 10px;
  }

  :global(.tc-progress-card) {
    display: grid;
    gap: 8px;
    padding: 10px;
    border: 1px solid #444444;
    border-radius: 8px;
    background: #252525;
  }

  :global(.tc-progress-heading) {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    align-items: center;
  }

  :global(.tc-progress-heading span:first-child) {
    color: #ffffff;
    font-weight: 800;
  }

  :global(.tc-progress-track) {
    height: 12px;
    overflow: hidden;
    border: 1px solid #505050;
    border-radius: 999px;
    background: #1f1f1f;
  }

  :global(.tc-progress-fill) {
    min-width: 0;
    height: 100%;
    border-radius: 999px;
    background: #7bd68f;
    transition: width 0.18s ease;
  }

  :global(.tc-progress-working) {
    background: #ffe08a;
  }

  :global(.tc-tabs) {
    padding: 4px;
    border: 1px solid #444444;
    border-radius: 8px;
    background: #252525;
  }

  :global(.tc-tab-active) {
    border-color: #707070;
    background: #454545;
  }

  :global(.tc-status-grid),
  :global(.tc-diagnostics),
  :global(.tc-help-list) {
    display: grid;
    gap: 8px;
  }

  :global(.tc-status-row),
  :global(.tc-diagnostic-row),
  :global(.tc-result-row) {
    display: grid;
    grid-template-columns: minmax(150px, 0.8fr) minmax(0, 1.2fr);
    gap: 10px;
    padding: 10px;
    border: 1px solid #444444;
    border-radius: 8px;
    background: #252525;
  }

  :global(.tc-muted) {
    color: #a9a9a9;
  }

  :global(.tc-success) {
    color: #d8f4e5;
  }

  :global(.tc-warning) {
    color: #ffe8b0;
  }

  :global(.tc-error) {
    color: #ffd6d6;
  }

  :global(.tc-credit),
  :global(.tc-file) {
    color: #cfcfcf;
    font-size: 0.9rem;
  }

  :global(.tc-result) {
    padding: 12px;
    border: 1px solid #444444;
    border-radius: 8px;
    color: #eeeeee;
    background: #252525;
    font-weight: 800;
  }

  :global(.tc-result.tc-success) {
    border-color: #4d6f61;
    background: #2f3a35;
  }

  :global(.tc-result.tc-warning) {
    border-color: #756235;
    color: #ffe8b0;
    background: #3a3528;
  }

  :global(.tc-result.tc-error) {
    border-color: #7b4a4a;
    background: #3a2e2e;
  }

  :global(.tc-pre-small) {
    min-height: 180px;
    max-height: 280px;
  }

  h1,
  h2,
  h3,
  p {
    margin: 0;
  }

  h1,
  h2,
  h3 {
    color: #ffffff;
    letter-spacing: 0;
  }

  p,
  label,
  span {
    color: #d0d0d0;
    line-height: 1.5;
  }

  button,
  select,
  input,
  textarea {
    min-height: 40px;
    border: 1px solid #505050;
    border-radius: 8px;
    color: #eeeeee;
    background: #252525;
  }

  button {
    padding: 0 12px;
    cursor: pointer;
    font-weight: 800;
  }

  button:hover {
    border-color: #666666;
    background: #414141;
  }

  button:disabled {
    cursor: not-allowed;
    opacity: 0.5;
  }

  select,
  input,
  textarea {
    width: 100%;
    padding: 0 10px;
  }

  textarea {
    min-height: 96px;
    padding: 10px;
    resize: vertical;
    line-height: 1.45;
  }

  pre {
    min-height: 110px;
    margin: 0;
    padding: 12px;
    overflow: auto;
    border: 1px solid #444444;
    border-radius: 8px;
    color: #f2f2f2;
    background: #1f1f1f;
    font-family: "Cascadia Code", "SFMono-Regular", Consolas, monospace;
    font-size: 0.86rem;
    line-height: 1.45;
  }

  .tc-unsupported {
    padding: 10px;
    border-radius: 8px;
    color: #ffd6d6;
    background: #3a2e2e;
  }
</style>
