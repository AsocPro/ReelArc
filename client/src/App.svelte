<script lang="ts">
  import { onMount } from 'svelte';
  import Timeline from './components/Timeline.svelte';
  import SidePanel from './components/SidePanel.svelte';
  import UploadForm from './components/UploadForm.svelte';
  import TranscriptionStatus from './components/TranscriptionStatus.svelte';
  import type { MediaItem } from './lib/types';
  import { fetchMediaItems } from './lib/api';
  
  let mediaItems: MediaItem[] = [];
  let selectedItem: MediaItem | null = null;
  let sidePanelOpen = false;
  let loading = true;
  let error = '';
  let activeTab = 'timeline'; // 'timeline' or 'transcription'
  
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
    sidePanelOpen = true;
  }
  
  function handleSidePanelClose() {
    sidePanelOpen = false;
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
</script>

<main>
  <header>
    <h1>Timeline Media Viewer</h1>
  </header>
  
  <div class="container">
    <div class="upload-section">
      <h2>Upload Media</h2>
      <UploadForm on:upload-success={handleUploadSuccess} />
    </div>
    
    <div class="tabs">
      <button 
        class="tab-button" 
        class:active={activeTab === 'timeline'} 
        on:click={() => setActiveTab('timeline')}
      >
        Media Timeline
      </button>
      <button 
        class="tab-button" 
        class:active={activeTab === 'transcription'} 
        on:click={() => setActiveTab('transcription')}
      >
        Transcription Status
      </button>
    </div>
    
    {#if activeTab === 'timeline'}
      <div class="timeline-section">
        {#if loading}
          <div class="loading">Loading media data...</div>
        {:else if error}
          <div class="error">{error}</div>
        {:else}
          <Timeline 
            data={mediaItems} 
            on:item-select={handleItemSelect} 
          />
        {/if}
      </div>
    {:else if activeTab === 'transcription'}
      <div class="transcription-section">
        <TranscriptionStatus />
      </div>
    {/if}
  </div>
  
  <SidePanel 
    item={selectedItem} 
    open={sidePanelOpen}
    on:close={handleSidePanelClose}
    on:update={handleItemUpdate}
  />
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
    display: grid;
    grid-template-columns: 1fr;
    grid-gap: 2rem;
  }
  
  @media (min-width: 768px) {
    .container {
      grid-template-columns: 1fr;
    }
    
    .timeline-section, .transcription-section {
      grid-column: 1 / -1;
    }
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
    margin-bottom: 1rem;
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
  
  .tab-button:hover {
    color: #2196f3;
  }
  
  .tab-button.active {
    color: #2196f3;
    border-bottom-color: #2196f3;
  }
</style>