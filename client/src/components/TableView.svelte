<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { MediaItem } from '../lib/types';

  export let data: MediaItem[] = [];
  export let loading = false;
  export let error = '';

  const dispatch = createEventDispatcher<{
    'item-select': MediaItem;
  }>();

  function handleRowClick(item: MediaItem) {
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
      default: return '📄';
    }
  }
</script>

<div class="table-container">
  {#if loading}
    <div class="loading">Loading table data...</div>
  {:else if error}
    <div class="error">{error}</div>
  {:else if data.length === 0}
    <div class="empty">No media items found</div>
  {:else}
    <div class="table-wrapper">
      <table class="media-table">
        <thead>
          <tr>
            <th>Type</th>
            <th>Filename</th>
            <th>Timestamp</th>
            <th>Duration</th>
            <th>Labels</th>
            <th>Transcription</th>
          </tr>
        </thead>
        <tbody>
          {#each data as item}
            <tr 
              class="table-row" 
              class:has-transcription={item.transcription}
              on:click={() => handleRowClick(item)}
            >
              <td class="type-cell">
                <span class="type-icon">{getTypeIcon(item.type)}</span>
                <span class="type-text">{item.type}</span>
              </td>
              <td class="filename-cell" title={item.filename}>
                {item.filename}
              </td>
              <td class="timestamp-cell">
                {formatTimestamp(item.timestamp)}
              </td>
              <td class="duration-cell">
                {formatDuration(item.duration)}
              </td>
              <td class="labels-cell">
                {#if item.labels && item.labels.length > 0}
                  <div class="labels">
                    {#each item.labels.slice(0, 3) as label}
                      <span class="label-tag">{label}</span>
                    {/each}
                    {#if item.labels.length > 3}
                      <span class="label-more">+{item.labels.length - 3}</span>
                    {/if}
                  </div>
                {:else}
                  <span class="no-labels">-</span>
                {/if}
              </td>
              <td class="transcription-cell">
                {#if item.transcription}
                  <div class="transcription-preview" title={item.transcription}>
                    {item.transcription.substring(0, 100)}{item.transcription.length > 100 ? '...' : ''}
                  </div>
                {:else}
                  <span class="no-transcription">-</span>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .table-container {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  .table-wrapper {
    flex: 1;
    overflow: auto;
  }

  .media-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.9rem;
  }

  .media-table th {
    background-color: #f5f5f5;
    padding: 0.75rem;
    text-align: left;
    font-weight: 600;
    color: #333;
    border-bottom: 2px solid #e0e0e0;
    position: sticky;
    top: 0;
    z-index: 1;
  }

  .media-table td {
    padding: 0.75rem;
    border-bottom: 1px solid #e0e0e0;
    vertical-align: top;
  }

  .table-row {
    cursor: pointer;
    transition: background-color 0.2s ease;
  }

  .table-row:hover {
    background-color: #f9f9f9;
  }

  .table-row.has-transcription {
    border-left: 3px solid #4caf50;
  }

  .type-cell {
    white-space: nowrap;
    min-width: 80px;
  }

  .type-icon {
    margin-right: 0.5rem;
    font-size: 1.1rem;
  }

  .type-text {
    text-transform: capitalize;
    color: #666;
  }

  .filename-cell {
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-family: monospace;
    font-size: 0.85rem;
  }

  .timestamp-cell {
    white-space: nowrap;
    color: #666;
    font-size: 0.85rem;
  }

  .duration-cell {
    white-space: nowrap;
    color: #666;
    font-family: monospace;
    text-align: right;
  }

  .labels-cell {
    max-width: 150px;
  }

  .labels {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
  }

  .label-tag {
    background-color: #e3f2fd;
    color: #1976d2;
    padding: 0.2rem 0.4rem;
    border-radius: 12px;
    font-size: 0.75rem;
    white-space: nowrap;
  }

  .label-more {
    color: #666;
    font-size: 0.75rem;
    font-style: italic;
  }

  .no-labels {
    color: #999;
  }

  .transcription-cell {
    max-width: 300px;
  }

  .transcription-preview {
    color: #333;
    line-height: 1.4;
    font-size: 0.85rem;
  }

  .no-transcription {
    color: #999;
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
</style>