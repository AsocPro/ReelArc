<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { MediaItem } from '../lib/types';
  import { updateLabels } from '../lib/api';
  
  export let item: MediaItem | null = null;
  export let open = false;
  
  const dispatch = createEventDispatcher<{
    'close': void;
    'update': MediaItem;
  }>();
  
  let newLabel = '';
  let labels: string[] = [];
  let saving = false;
  
  $: if (item) {
    labels = [...item.labels];
  }
  
  function closePanel() {
    dispatch('close');
  }
  
  function addLabel() {
    if (newLabel.trim() && !labels.includes(newLabel.trim())) {
      labels = [...labels, newLabel.trim()];
      newLabel = '';
    }
  }
  
  function removeLabel(index: number) {
    labels = labels.filter((_, i) => i !== index);
  }
  
  async function saveLabels() {
    if (!item) return;
    
    saving = true;
    try {
      const updatedItem = await updateLabels(item.id, labels);
      if (updatedItem) {
        // In a real app, we'd use the server response
        // For now, we'll just update our local item
        const updated = { ...item, labels };
        dispatch('update', updated);
      }
    } catch (error) {
      console.error('Failed to save labels:', error);
    } finally {
      saving = false;
    }
  }
  
  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      addLabel();
    }
  }
</script>

<div class="side-panel" class:open>
  <div class="panel-header">
    <h2>Media Details</h2>
    <button class="close-btn" on:click={closePanel}>×</button>
  </div>
  
  {#if item}
    <div class="panel-content">
      <div class="media-preview">
        {#if item.type === 'photo'}
          <img src={`/media/${item.filename}`} alt={item.filename} />
        {:else if item.type === 'audio'}
          <audio controls>
            <source src={`/media/${item.filename}`} type="audio/mpeg">
            Your browser does not support the audio element.
          </audio>
        {:else if item.type === 'video'}
          <video controls>
            <source src={`/media/${item.filename}`} type="video/mp4">
            Your browser does not support the video element.
          </video>
        {/if}
      </div>
      
      <div class="media-info">
        <div class="info-item">
          <span class="label">Filename:</span>
          <span class="value">{item.filename}</span>
        </div>
        
        <div class="info-item">
          <span class="label">Type:</span>
          <span class="value">{item.type}</span>
        </div>
        
        <div class="info-item">
          <span class="label">Timestamp:</span>
          <span class="value">{new Date(item.timestamp).toLocaleString()}</span>
        </div>
        
        {#if item.duration}
          <div class="info-item">
            <span class="label">Duration:</span>
            <span class="value">{item.duration} seconds</span>
          </div>
        {/if}
        
        {#if item.transcription}
          <div class="info-item transcription">
            <span class="label">Transcription:</span>
            <div class="value transcription-text">{item.transcription}</div>
          </div>
        {/if}
        
        <div class="info-item labels-section">
          <span class="label">Labels:</span>
          <div class="labels-container">
            {#if labels.length === 0}
              <span class="no-labels">No labels</span>
            {:else}
              <div class="labels-list">
                {#each labels as label, i}
                  <div class="label-item">
                    <span>{label}</span>
                    <button class="remove-label" on:click={() => removeLabel(i)}>×</button>
                  </div>
                {/each}
              </div>
            {/if}
            
            <div class="add-label">
              <input 
                type="text" 
                bind:value={newLabel} 
                placeholder="Add new label" 
                on:keydown={handleKeydown}
              />
              <button on:click={addLabel}>Add</button>
            </div>
            
            <button class="save-btn" on:click={saveLabels} disabled={saving}>
              {saving ? 'Saving...' : 'Save Labels'}
            </button>
          </div>
        </div>
      </div>
    </div>
  {:else}
    <div class="panel-content empty">
      <p>Select an item from the timeline to view details</p>
    </div>
  {/if}
</div>

<style>
  .side-panel {
    position: fixed;
    top: 0;
    right: -400px;
    width: 380px;
    height: 100vh;
    background-color: white;
    box-shadow: -2px 0 10px rgba(0, 0, 0, 0.1);
    transition: right 0.3s ease;
    z-index: 1000;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  
  .side-panel.open {
    right: 0;
  }
  
  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    border-bottom: 1px solid #eee;
  }
  
  .panel-header h2 {
    margin: 0;
    font-size: 1.25rem;
  }
  
  .close-btn {
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
    color: #666;
  }
  
  .panel-content {
    flex: 1;
    overflow-y: auto;
    padding: 1rem;
  }
  
  .panel-content.empty {
    display: flex;
    justify-content: center;
    align-items: center;
    color: #999;
  }
  
  .media-preview {
    margin-bottom: 1.5rem;
    display: flex;
    justify-content: center;
    align-items: center;
    max-height: 200px;
    overflow: hidden;
    border-radius: 4px;
    background-color: #f5f5f5;
  }
  
  .media-preview img {
    max-width: 100%;
    max-height: 200px;
    object-fit: contain;
  }
  
  .media-preview audio,
  .media-preview video {
    width: 100%;
  }
  
  .media-info {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }
  
  .info-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
  
  .info-item .label {
    font-weight: 600;
    color: #555;
    font-size: 0.875rem;
  }
  
  .info-item .value {
    color: #333;
  }
  
  .transcription-text {
    background-color: #f5f5f5;
    padding: 0.75rem;
    border-radius: 4px;
    font-size: 0.875rem;
    line-height: 1.5;
    max-height: 150px;
    overflow-y: auto;
    white-space: pre-wrap;
  }
  
  .labels-container {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }
  
  .no-labels {
    color: #999;
    font-style: italic;
    font-size: 0.875rem;
  }
  
  .labels-list {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }
  
  .label-item {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    background-color: #e3f2fd;
    color: #1976d2;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.75rem;
  }
  
  .remove-label {
    background: none;
    border: none;
    color: #1976d2;
    cursor: pointer;
    font-size: 1rem;
    padding: 0;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  
  .add-label {
    display: flex;
    gap: 0.5rem;
    margin-top: 0.5rem;
  }
  
  .add-label input {
    flex: 1;
    padding: 0.5rem;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 0.875rem;
  }
  
  .add-label button {
    background-color: #e0e0e0;
    border: none;
    padding: 0.5rem 0.75rem;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.875rem;
  }
  
  .save-btn {
    margin-top: 1rem;
    background-color: #2196f3;
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.875rem;
  }
  
  .save-btn:disabled {
    background-color: #bdbdbd;
    cursor: not-allowed;
  }
</style>