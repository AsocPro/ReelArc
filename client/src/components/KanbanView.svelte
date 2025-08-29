<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { MediaItem } from '../lib/types';

  export let data: MediaItem[] = [];
  export let loading = false;
  export let error = '';

  const dispatch = createEventDispatcher<{
    'item-select': MediaItem;
  }>();

  function handleItemClick(item: MediaItem) {
    dispatch('item-select', item);
  }

  function formatTimestamp(timestamp: string): string {
    return new Date(timestamp).toLocaleString();
  }

  function formatDuration(duration?: number): string {
    if (!duration) return '-';
    const minutes = Math.floor(duration / 60);
    const seconds = Math.floor(duration % 60);
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
  }

  function getTypeIcon(type: string): string {
    switch (type) {
      case 'photo': return '📷';
      case 'audio': return '🎵';
      case 'video': return '🎬';
      case 'note': return '📝';
      default: return '📄';
    }
  }

  // Group items by type for Kanban columns
  $: groupedItems = data.reduce((groups, item) => {
    const type = item.type;
    if (!groups[type]) {
      groups[type] = [];
    }
    groups[type].push(item);
    return groups;
  }, {} as Record<string, MediaItem[]>);

  // Define column order
  const columnOrder = ['photo', 'audio', 'video', 'note'];

  function getColumnTitle(type: string): string {
    switch (type) {
      case 'photo': return 'Photos';
      case 'audio': return 'Audio';
      case 'video': return 'Videos';
      case 'note': return 'Notes';
      default: return type.charAt(0).toUpperCase() + type.slice(1);
    }
  }

  function getColumnColor(type: string): string {
    switch (type) {
      case 'photo': return '#e8f5e8';
      case 'audio': return '#e3f2fd';
      case 'video': return '#fff3e0';
      case 'note': return '#f3e5f5';
      default: return '#f5f5f5';
    }
  }
</script>

<div class="kanban-container">
  {#if loading}
    <div class="loading">Loading kanban data...</div>
  {:else if error}
    <div class="error">{error}</div>
  {:else if data.length === 0}
    <div class="empty">No media items found</div>
  {:else}
    <div class="kanban-board">
      {#each columnOrder as type}
        {@const items = groupedItems[type] || []}
        <div class="kanban-column" style="background-color: {getColumnColor(type)}">
          <div class="column-header">
            <h3 class="column-title">
              <span class="type-icon">{getTypeIcon(type)}</span>
              {getColumnTitle(type)}
            </h3>
            <span class="item-count">{items.length}</span>
          </div>
          <div class="column-content">
            {#each items as item}
              <div
                class="kanban-card"
                class:has-transcription={item.transcription}
                on:click={() => handleItemClick(item)}
              >
                <div class="card-header">
                  <div class="filename" title={item.filename}>
                    {item.filename}
                  </div>
                  <div class="timestamp">
                    {formatTimestamp(item.timestamp)}
                  </div>
                </div>

                {#if item.duration}
                  <div class="duration">
                    Duration: {formatDuration(item.duration)}
                  </div>
                {/if}

                {#if item.labels && item.labels.length > 0}
                  <div class="labels">
                    {#each item.labels.slice(0, 2) as label}
                      <span class="label-tag">{label}</span>
                    {/each}
                    {#if item.labels.length > 2}
                      <span class="label-more">+{item.labels.length - 2}</span>
                    {/if}
                  </div>
                {/if}

                {#if item.transcription}
                  <div class="transcription-preview" title={item.transcription}>
                    {item.transcription.substring(0, 80)}{item.transcription.length > 80 ? '...' : ''}
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .kanban-container {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .kanban-board {
    display: flex;
    gap: 16px;
    padding: 16px;
    height: 100%;
    overflow-x: auto;
    overflow-y: hidden;
  }

  .kanban-column {
    flex: 1;
    min-width: 280px;
    max-width: 320px;
    border-radius: 8px;
    border: 1px solid #e0e0e0;
    display: flex;
    flex-direction: column;
    background-color: #f9f9f9;
  }

  .column-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid #e0e0e0;
    background-color: rgba(255, 255, 255, 0.8);
    border-radius: 8px 8px 0 0;
  }

  .column-title {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: #333;
  }

  .type-icon {
    font-size: 1.1rem;
  }

  .item-count {
    background-color: #666;
    color: white;
    padding: 2px 8px;
    border-radius: 12px;
    font-size: 0.75rem;
    font-weight: 500;
  }

  .column-content {
    flex: 1;
    padding: 12px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .kanban-card {
    background-color: white;
    border: 1px solid #e0e0e0;
    border-radius: 6px;
    padding: 12px;
    cursor: pointer;
    transition: all 0.2s ease;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .kanban-card:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
    transform: translateY(-1px);
  }

  .kanban-card.has-transcription {
    border-left: 3px solid #4caf50;
  }

  .card-header {
    margin-bottom: 8px;
  }

  .filename {
    font-weight: 600;
    font-size: 0.9rem;
    color: #333;
    margin-bottom: 4px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .timestamp {
    font-size: 0.75rem;
    color: #666;
  }

  .duration {
    font-size: 0.8rem;
    color: #666;
    margin-bottom: 8px;
    font-family: monospace;
  }

  .labels {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    margin-bottom: 8px;
  }

  .label-tag {
    background-color: #e3f2fd;
    color: #1976d2;
    padding: 2px 6px;
    border-radius: 10px;
    font-size: 0.7rem;
    white-space: nowrap;
  }

  .label-more {
    color: #666;
    font-size: 0.7rem;
    font-style: italic;
  }

  .transcription-preview {
    font-size: 0.8rem;
    color: #333;
    line-height: 1.4;
    padding: 6px;
    background-color: #f8f9fa;
    border-radius: 4px;
    border-left: 2px solid #4caf50;
  }

  .loading, .error, .empty {
    display: flex;
    justify-content: center;
    align-items: center;
    width: 100%;
    height: 100%;
    color: #666;
    font-size: 1rem;
  }

  .error {
    color: #d32f2f;
  }

  /* Responsive adjustments */
  @media (max-width: 768px) {
    .kanban-board {
      padding: 8px;
      gap: 8px;
    }

    .kanban-column {
      min-width: 240px;
      max-width: 280px;
    }
  }
</style>