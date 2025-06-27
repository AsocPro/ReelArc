<script lang="ts">
  import { onMount } from 'svelte';
  import Timeline from './components/Timeline.svelte';

  import UploadForm from './components/UploadForm.svelte';
  import TranscriptionStatus from './components/TranscriptionStatus.svelte';
  import MediaDetails from './components/MediaDetails.svelte';
  import type { MediaItem } from './lib/types';
  import { fetchMediaItems } from './lib/api';
  
  let mediaItems: MediaItem[] = [];
  let selectedItem: MediaItem | null = null;
  let loading = true;
  let error = '';
  let activeTab = 'upload'; // 'transcription', 'upload', or 'details'
  let timelineComponent: any;
  
  onMount(async () => {
    await loadMediaItems();
  });
  
  async function loadMediaItems() {
    try {
      loading = true;
      mediaItems = await fetchMediaItems();
      loading = false;
    } catch (err) {
      loading = false;
      error = 'Failed to load media items';
      console.error(error, err);
    }
  }
  
  function handleItemSelect(event: CustomEvent<MediaItem>) {
    selectedItem = event.detail;
    // Switch to the details tab
    setActiveTab('details');
  }
  

  
  function handleItemUpdate(event: CustomEvent<MediaItem>) {
    const updatedItem = event.detail;
    // Update the item in our array
    mediaItems = mediaItems.map(item => 
      item.id === updatedItem.id ? updatedItem : item
    );
    // Update the selected item
    selectedItem = updatedItem;
  }
  
  function handleUploadSuccess() {
    // Refresh media items after successful upload
    loadMediaItems();
  }
  
  function setActiveTab(tab: string) {
    activeTab = tab;
  }
  
  function handleCenterPlayhead() {
    if (timelineComponent && typeof timelineComponent.centerOnPlayhead === 'function') {
      timelineComponent.centerOnPlayhead();
    }
  }
</script>

<main>
  <header>
    <h1>Timeline Media Viewer</h1>
  </header>
  
  <div class="container">
    <!-- Timeline component always visible at the top -->
    <div class="timeline-container">
      {#if loading}
        <div class="loading">Loading media data...</div>
      {:else if error}
        <div class="error">{error}</div>
      {:else}
        <Timeline 
          data={mediaItems} 
          on:item-select={handleItemSelect}
          on:center-playhead
          bind:this={timelineComponent}
        />
      {/if}
    </div>

    <div class="tabs">
      <button 
        class="tab-button" 
        class:active={activeTab === 'upload'} 
        on:click={() => setActiveTab('upload')}
      >
        Upload Media
      </button>
      <button 
        class="tab-button" 
        class:active={activeTab === 'details'} 
        on:click={() => setActiveTab('details')}
        disabled={!selectedItem}
      >
        Media Details
      </button>
      <button 
        class="tab-button" 
        class:active={activeTab === 'transcription'} 
        on:click={() => setActiveTab('transcription')}
      >
        Transcription Status
      </button>
    </div>
    
    <div class="content-section">
      {#if activeTab === 'upload'}
        <div class="upload-section">
          <UploadForm on:upload-success={handleUploadSuccess} />
        </div>
      {:else if activeTab === 'details'}
        <div class="details-section">
          <MediaDetails 
            item={selectedItem} 
            on:update={handleItemUpdate}
            on:center-playhead={handleCenterPlayhead}
          />
        </div>
      {:else if activeTab === 'transcription'}
        <div class="transcription-section">
          <TranscriptionStatus />
        </div>
      {/if}
    </div>
  </div>
  

</main>

<style>
  :global(body) {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen,
      Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
    margin: 0;
    padding: 0;
    box-sizing: border-box;
  }
  
  main {
    max-width: 1200px;
    margin: 0 auto;
    padding: 1rem;
  }
  
  header {
    margin-bottom: 2rem;
    border-bottom: 1px solid #eee;
  }
  
  h1 {
    color: #333;
  }
  
  .container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  
  .content-section {
    flex: 1;
    min-height: 400px;
    border: 1px solid #eee;
    border-radius: 4px;
    overflow: hidden;
  }
  
  .timeline-container {
    margin-bottom: 1rem;
    border: 1px solid #eee;
    border-radius: 4px;
    overflow: hidden;
    min-height: 200px;
  }
  
  .upload-section, 
  .details-section, 
  .transcription-section {
    height: 100%;
    min-height: 400px;
  }
  
  h2 {
    margin-top: 0;
    color: #555;
  }
  
  .loading, .error {
    padding: 2rem;
    text-align: center;
    background-color: #f5f5f5;
    border-radius: 4px;
    color: #666;
  }
  
  .error {
    background-color: #ffebee;
    color: #d32f2f;
  }
  
  .tabs {
    display: flex;
    margin-bottom: 0;
    border-bottom: 1px solid #eee;
  }
  
  .tab-button {
    padding: 0.75rem 1.5rem;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    cursor: pointer;
    font-size: 1rem;
    color: #666;
    transition: all 0.2s ease;
  }
  
  .tab-button:hover:not(:disabled) {
    color: #2196f3;
  }
  
  .tab-button.active {
    color: #2196f3;
    border-bottom-color: #2196f3;
  }
  
  .tab-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>