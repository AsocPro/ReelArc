<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import type { TranscriptionStatus } from '../lib/types';
  import { fetchTranscriptionStatus } from '../lib/api';
  
  let statuses: TranscriptionStatus[] = [];
  let loading = true;
  let error = '';
  let refreshInterval: number;
  
  onMount(() => {
    loadTranscriptionStatus();
    
    // Refresh status every 10 seconds
    refreshInterval = setInterval(loadTranscriptionStatus, 10000) as unknown as number;
  });
  
  onDestroy(() => {
    if (refreshInterval) {
      clearInterval(refreshInterval);
    }
  });
  
  async function loadTranscriptionStatus() {
    try {
      loading = true;
      const result = await fetchTranscriptionStatus();
      statuses = result || [];
      loading = false;
    } catch (err) {
      loading = false;
      error = 'Failed to load transcription status';
      console.error(error, err);
    }
  }
  
  function getStatusClass(status: string): string {
    switch (status) {
      case 'completed':
        return 'status-completed';
      case 'processing':
        return 'status-processing';
      case 'queued':
        return 'status-queued';
      case 'failed':
        return 'status-failed';
      default:
        return '';
    }
  }
  
  function formatTimestamp(timestamp: string): string {
    return new Date(timestamp).toLocaleString();
  }
</script>

<div class="transcription-status">
  <h2>Transcription Status</h2>
  
  <button class="refresh-btn" on:click={loadTranscriptionStatus} disabled={loading}>
    {loading ? 'Refreshing...' : 'Refresh'}
  </button>
  
  {#if loading && (!statuses || statuses.length === 0)}
    <div class="loading">Loading transcription status...</div>
  {:else if error}
    <div class="error">{error}</div>
  {:else if !statuses || statuses.length === 0}
    <div class="empty">No transcription jobs found</div>
  {:else}
    <div class="status-list">
      <div class="status-header">
        <div class="filename">Filename</div>
        <div class="status">Status</div>
        <div class="timestamp">Timestamp</div>
      </div>
      
      {#each statuses as status}
        <div class="status-item">
          <div class="filename">{status.filename}</div>
          <div class="status">
            <span class={getStatusClass(status.status)}>{status.status}</span>
            {#if status.error}
              <div class="error-message">{status.error}</div>
            {/if}
          </div>
          <div class="timestamp">{formatTimestamp(status.timestamp)}</div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .transcription-status {
    padding: 1rem;
    background-color: white;
    height: 100%;
    overflow-y: auto;
  }
  
  h2 {
    margin-top: 0;
    margin-bottom: 1rem;
    color: #333;
  }
  
  .refresh-btn {
    background-color: #2196f3;
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.875rem;
    margin-bottom: 1rem;
  }
  
  .refresh-btn:disabled {
    background-color: #bdbdbd;
    cursor: not-allowed;
  }
  
  .loading, .error, .empty {
    padding: 2rem;
    text-align: center;
    background-color: #f5f5f5;
    border-radius: 4px;
    color: #666;
    margin-top: 1rem;
  }
  
  .error {
    background-color: #ffebee;
    color: #d32f2f;
  }
  
  .empty {
    color: #999;
    font-style: italic;
  }
  
  .status-list {
    margin-top: 1rem;
    border: 1px solid #eee;
    border-radius: 4px;
    overflow: hidden;
  }
  
  .status-header {
    display: grid;
    grid-template-columns: 2fr 1fr 1fr;
    gap: 1rem;
    padding: 0.75rem 1rem;
    background-color: #f5f5f5;
    font-weight: 600;
    color: #555;
    border-bottom: 1px solid #eee;
  }
  
  .status-item {
    display: grid;
    grid-template-columns: 2fr 1fr 1fr;
    gap: 1rem;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid #eee;
  }
  
  .status-item:last-child {
    border-bottom: none;
  }
  
  .status-completed {
    color: #4caf50;
    font-weight: 600;
  }
  
  .status-processing {
    color: #2196f3;
    font-weight: 600;
  }
  
  .status-queued {
    color: #ff9800;
    font-weight: 600;
  }
  
  .status-failed {
    color: #f44336;
    font-weight: 600;
  }
  
  .error-message {
    font-size: 0.75rem;
    color: #f44336;
    margin-top: 0.25rem;
    white-space: pre-wrap;
    word-break: break-word;
  }
</style>