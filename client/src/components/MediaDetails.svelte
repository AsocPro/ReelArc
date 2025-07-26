<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { MediaItem } from '../lib/types';
  import { updateLabels } from '../lib/api';
  import { mediaPlayback } from '../lib/stores';
  
  export let item: MediaItem | null = null;
  
  const dispatch = createEventDispatcher<{
    'update': MediaItem;
    'center-playhead': void;
  }>();
  
  let newLabel = '';
  let labels: string[] = [];
  let saving = false;
  let mediaElement: HTMLMediaElement | null = null;
  let playbackInterval: number | null = null;
  
  $: if (item) {
    labels = [...(item.labels || [])];
    
    // Set up media tracking when item changes
    // Use setTimeout to ensure DOM is updated
    setTimeout(() => {
      setupMediaTracking();
    }, 0);
  }
  
  // Clean up any existing interval when component is destroyed
  function cleanup() {
    if (playbackInterval) {
      clearInterval(playbackInterval);
      playbackInterval = null;
    }
    
    // Reset media playback state when component is destroyed
    mediaPlayback.set({
      isPlaying: false,
      currentItem: null,
      startTimestamp: null,
      currentTime: 0
    });
  }
  
  // Call cleanup when component is destroyed
  import { onDestroy } from 'svelte';
  onDestroy(cleanup);
  
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
  
  // Format time in seconds to MM:SS format
  function formatTime(seconds: number): string {
    const minutes = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  }
  
  // Seek to a specific time in the audio/video player
  function seekToTime(seconds: number): void {
    if (!item) return;
    
    // Find the audio or video element
    const element = document.querySelector('.media-preview audio, .media-preview video') as HTMLMediaElement;
    if (element) {
      element.currentTime = seconds;
      element.play().catch(err => console.error('Failed to play media:', err));
      
      // Update the media playback store
      mediaPlayback.update(state => ({
        ...state,
        currentTime: seconds,
        isPlaying: true
      }));
      
      // Dispatch an event to center the timeline on the playhead
      dispatch('center-playhead');
    }
  }
  
  // Setup media playback tracking
  function setupMediaTracking(): void {
    if (!item || !item.timestamp) return;
    
    // Find the audio or video element
    mediaElement = document.querySelector('.media-preview audio, .media-preview video') as HTMLMediaElement;
    
    if (mediaElement) {
      // Set up event listeners for media playback
      mediaElement.addEventListener('play', () => {
        // Start tracking playback
        mediaPlayback.set({
          isPlaying: true,
          currentItem: item?.id || null,
          startTimestamp: item?.timestamp || null,
          currentTime: mediaElement?.currentTime || 0
        });
        
        // Set up interval to update current time
        if (playbackInterval) clearInterval(playbackInterval);
        playbackInterval = setInterval(() => {
          if (mediaElement && !mediaElement.paused) {
            mediaPlayback.update(state => ({
              ...state,
              currentTime: mediaElement?.currentTime || 0
            }));
          }
        }, 100) as unknown as number; // Update 10 times per second
      });
      
      mediaElement.addEventListener('pause', () => {
        mediaPlayback.update(state => ({
          ...state,
          isPlaying: false
        }));
        
        // Clear interval when paused
        if (playbackInterval) {
          clearInterval(playbackInterval);
          playbackInterval = null;
        }
      });
      
      mediaElement.addEventListener('ended', () => {
        mediaPlayback.update(state => ({
          ...state,
          isPlaying: false,
          currentTime: 0
        }));
        
        // Clear interval when ended
        if (playbackInterval) {
          clearInterval(playbackInterval);
          playbackInterval = null;
        }
      });
      
      mediaElement.addEventListener('timeupdate', () => {
        mediaPlayback.update(state => ({
          ...state,
          currentTime: mediaElement?.currentTime || 0
        }));
      });
      
      mediaElement.addEventListener('seeking', () => {
        mediaPlayback.update(state => ({
          ...state,
          currentTime: mediaElement?.currentTime || 0
        }));
      });
    }
  }
</script>

{#if item}
  <div class="media-details">
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
      
      {#if item.transcripts && item.transcripts.length > 0}
        <div class="info-item transcription">
          <span class="label">Transcript:</span>
          <div class="value transcript-container">
            {#each item.transcripts as entry}
              <div class="transcript-entry" on:click={() => seekToTime(entry.start)}>
                <span class="transcript-time">{formatTime(entry.start)} - {formatTime(entry.end)}</span>
                <span class="transcript-text">{entry.text}</span>
              </div>
            {/each}
          </div>
        </div>
      {:else if item.transcription}
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
  <div class="empty-state">
    <p>Select an item from the timeline to view details</p>
  </div>
{/if}

<style>
  .media-details {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow-y: auto;
    padding: 1rem;
  }
  
  .empty-state {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100%;
    color: #999;
    font-style: italic;
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
  
  .transcript-container {
    background-color: #f5f5f5;
    border-radius: 4px;
    font-size: 0.875rem;
    line-height: 1.5;
    max-height: 250px;
    overflow-y: auto;
  }
  
  .transcript-entry {
    padding: 0.5rem 0.75rem;
    border-bottom: 1px solid #e0e0e0;
    cursor: pointer;
    transition: background-color 0.2s ease;
  }
  
  .transcript-entry:last-child {
    border-bottom: none;
  }
  
  .transcript-entry:hover {
    background-color: #e3f2fd;
  }
  
  .transcript-time {
    display: block;
    font-size: 0.75rem;
    color: #666;
    margin-bottom: 0.25rem;
  }
  
  .transcript-text {
    display: block;
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