<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { TabConfig } from '../lib/types';

  export let tabs: TabConfig[] = [];
  export let activeTab: string = '';

  const dispatch = createEventDispatcher<{
    'tab-change': string;
  }>();

  function handleTabClick(tabId: string) {
    if (tabs.find(tab => tab.id === tabId && tab.disabled)) {
      return;
    }
    activeTab = tabId;
    dispatch('tab-change', tabId);
  }

  $: if (!activeTab && tabs.length > 0) {
    activeTab = tabs[0]?.id || '';
  }
</script>

<div class="tab-container">
  <div class="tab-header">
    {#each tabs as tab}
      <button
        class="tab-button"
        class:active={activeTab === tab.id}
        class:disabled={tab.disabled}
        disabled={tab.disabled}
        on:click={() => handleTabClick(tab.id)}
      >
        {tab.label}
      </button>
    {/each}
  </div>
  
  <div class="tab-content">
    <slot {activeTab} />
  </div>
</div>

<style>
  .tab-container {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .tab-header {
    display: flex;
    border-bottom: 1px solid #e0e0e0;
    background-color: #fafafa;
  }

  .tab-button {
    padding: 0.75rem 1.5rem;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    cursor: pointer;
    font-size: 0.9rem;
    color: #666;
    transition: all 0.2s ease;
    white-space: nowrap;
  }

  .tab-button:hover:not(:disabled) {
    color: #2196f3;
    background-color: #f5f5f5;
  }

  .tab-button.active {
    color: #2196f3;
    border-bottom-color: #2196f3;
    background-color: #fff;
  }

  .tab-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .tab-content {
    flex: 1;
    overflow: hidden;
    background-color: #fff;
  }
</style>