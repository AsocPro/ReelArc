<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import TabContainer from './TabContainer.svelte';
  import Timeline from './Timeline.svelte';
  import TableView from './TableView.svelte';
  import type { MediaItem, TabConfig } from '../lib/types';

  export let data: MediaItem[] = [];
  export let loading = false;
  export let error = '';

  const dispatch = createEventDispatcher<{
    'item-select': MediaItem;
    'center-playhead': void;
  }>();

  let activeTab = 'timeline';
  let timelineComponent: any;

  const tabs: TabConfig[] = [
    { id: 'timeline', label: 'Timeline View' },
    { id: 'table', label: 'Table View' }
  ];

  function handleTabChange(event: CustomEvent<string>) {
    activeTab = event.detail;
  }

  function handleItemSelect(event: CustomEvent<MediaItem>) {
    dispatch('item-select', event.detail);
  }

  function handleCenterPlayhead() {
    dispatch('center-playhead');
  }

  export function centerOnPlayhead() {
    if (timelineComponent && typeof timelineComponent.centerOnPlayhead === 'function') {
      timelineComponent.centerOnPlayhead();
    }
  }
  
  export function refreshTimeline() {
    if (timelineComponent && typeof timelineComponent.refreshTimeline === 'function') {
      timelineComponent.refreshTimeline();
    }
  }
</script>

<div class="timeline-viewer">
  <TabContainer {tabs} {activeTab} on:tab-change={handleTabChange}>
    <div slot="default" let:activeTab class="view-container">
      {#if activeTab === 'timeline'}
        <Timeline 
          {data} 
          {loading} 
          {error}
          on:item-select={handleItemSelect}
          on:center-playhead={handleCenterPlayhead}
          bind:this={timelineComponent}
        />
      {:else if activeTab === 'table'}
        <TableView 
          {data} 
          {loading} 
          {error}
          on:item-select={handleItemSelect}
        />
      {/if}
    </div>
  </TabContainer>
</div>

<style>
  .timeline-viewer {
    width: 100%;
    height: 100%;
    min-height: 400px;
    border: 1px solid #e0e0e0;
    border-radius: 4px;
    overflow: hidden;
    background-color: #fff;
  }

  .view-container {
    width: 100%;
    height: 100%;
    overflow: hidden;
  }
</style>