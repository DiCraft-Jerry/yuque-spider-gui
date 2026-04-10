<script>
  import { onDestroy, tick } from 'svelte';

  /** @type {string | boolean} */
  export let value;

  /** @type {{ value: string | boolean, label: string }[]} */
  export let options = [];

  let open = false;
  let rootEl;
  let triggerEl;
  let menuTop = 0;
  let menuLeft = 0;
  let menuWidth = 0;

  /** @type {(() => void) | null} */
  let detachSidebarScroll = null;

  $: currentLabel = options.find((o) => o.value === value)?.label ?? '—';

  function close() {
    open = false;
    if (detachSidebarScroll) {
      detachSidebarScroll();
      detachSidebarScroll = null;
    }
  }

  function attachSidebarScrollClose() {
    if (detachSidebarScroll) {
      detachSidebarScroll();
      detachSidebarScroll = null;
    }
    const el = document.querySelector('.app-sidebar');
    if (!el) return;
    const fn = () => close();
    el.addEventListener('scroll', fn, { passive: true });
    detachSidebarScroll = () => el.removeEventListener('scroll', fn);
  }

  async function toggle(e) {
    e?.stopPropagation?.();
    open = !open;
    if (open) {
      await tick();
      placeMenu();
      attachSidebarScrollClose();
    } else if (detachSidebarScroll) {
      detachSidebarScroll();
      detachSidebarScroll = null;
    }
  }

  function placeMenu() {
    if (!triggerEl) return;
    const r = triggerEl.getBoundingClientRect();
    menuTop = r.bottom + 4;
    menuLeft = r.left;
    menuWidth = r.width;
  }

  function pick(v, e) {
    e?.stopPropagation?.();
    value = v;
    close();
  }

  function onDocClick(e) {
    if (!open || !rootEl) return;
    if (!rootEl.contains(/** @type {Node} */ (e.target))) close();
  }

  function onWinResize() {
    if (open) placeMenu();
  }

  onDestroy(() => {
    if (detachSidebarScroll) detachSidebarScroll();
  });
</script>

<svelte:window on:click={onDocClick} on:resize={onWinResize} />

<div class="config-dropdown" bind:this={rootEl}>
  <button
    type="button"
    class="config-dropdown-trigger"
    bind:this={triggerEl}
    aria-haspopup="listbox"
    aria-expanded={open}
    on:click|stopPropagation={toggle}
  >
    <span class="config-dropdown-value">{currentLabel}</span>
    <span class="config-dropdown-chevron" class:open>
      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" aria-hidden="true">
        <path
          d="M6 9l6 6 6-6"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
      </svg>
    </span>
  </button>

  {#if open}
    <ul
      class="config-dropdown-menu"
      role="listbox"
      style="top:{menuTop}px;left:{menuLeft}px;width:{menuWidth}px;"
    >
      {#each options as opt (String(opt.value))}
        <li role="presentation">
          <button
            type="button"
            role="option"
            class="config-dropdown-option"
            class:selected={opt.value === value}
            aria-selected={opt.value === value}
            on:click|stopPropagation={(e) => pick(opt.value, e)}
          >
            {opt.label}
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .config-dropdown {
    position: relative;
    width: 100%;
  }

  .config-dropdown-trigger {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    width: 100%;
    padding: 8px 10px;
    border-radius: 8px;
    border: 1px solid var(--sidebar-input-border);
    background: var(--sidebar-input-bg);
    color: var(--sidebar-text-main);
    font-size: 0.85rem;
    box-sizing: border-box;
    cursor: pointer;
    text-align: left;
    transition:
      border-color 0.15s ease,
      box-shadow 0.15s ease;
  }

  .config-dropdown-trigger:hover {
    border-color: rgba(99, 102, 241, 0.45);
  }

  .config-dropdown-trigger:focus {
    outline: none;
    border-color: #6366f1;
    box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.28);
  }

  .config-dropdown-value {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .config-dropdown-chevron {
    flex-shrink: 0;
    display: flex;
    color: var(--sidebar-text-sub);
    transition: transform 0.2s ease;
  }

  .config-dropdown-chevron.open {
    transform: rotate(180deg);
  }

  .config-dropdown-menu {
    position: fixed;
    margin: 0;
    padding: 4px;
    list-style: none;
    z-index: 4000;
    border-radius: 10px;
    border: 1px solid var(--sidebar-border);
    background: var(--bg-sidebar-card);
    backdrop-filter: blur(8px);
    box-shadow:
      0 10px 40px rgba(15, 23, 42, 0.12),
      0 2px 8px rgba(15, 23, 42, 0.06);
    box-sizing: border-box;
    max-height: min(280px, 70vh);
    overflow-y: auto;
  }

  :global(.theme-dark) .config-dropdown-menu {
    box-shadow:
      0 12px 48px rgba(0, 0, 0, 0.45),
      0 0 0 1px rgba(255, 255, 255, 0.06);
  }

  .config-dropdown-option {
    display: block;
    width: 100%;
    padding: 8px 10px;
    margin: 0;
    border: none;
    border-radius: 6px;
    background: transparent;
    color: var(--sidebar-text-main);
    font-size: 0.82rem;
    text-align: left;
    cursor: pointer;
    transition: background 0.12s ease;
  }

  .config-dropdown-option:hover,
  .config-dropdown-option:focus {
    outline: none;
    background: rgba(99, 102, 241, 0.12);
  }

  .config-dropdown-option.selected {
    background: rgba(99, 102, 241, 0.18);
    font-weight: 500;
  }

  :global(.theme-dark) .config-dropdown-option.selected {
    background: rgba(99, 102, 241, 0.22);
  }
</style>
