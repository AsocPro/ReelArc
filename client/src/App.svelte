<script lang="ts">
  import { onMount } from 'svelte';
  import TimelineViewer from './components/TimelineViewer.svelte';

  import UploadForm from './components/UploadForm.svelte';
  import TranscriptionStatus from './components/TranscriptionStatus.svelte';
  import MediaDetails from './components/MediaDetails.svelte';
  import type { MediaItem, MediaFilters } from './lib/types';
  import { fetchMediaItems } from './lib/api';
  
  let mediaItems: MediaItem[] = [];
  let selectedItem: MediaItem | null = null;
  let loading = true;
  let error = '';
  let activeTab = 'upload'; // 'transcription', 'upload', or 'details'
  let timelineViewerComponent: any;
  
  // Filter state
  let filters: MediaFilters = {};
  let startDate = '';
  let endDate = '';
  let labelFilter = '';
  let availableLabels: string[] = [];
  let showFilters = false;
  
  onMount(async () => {
    await loadMediaItems();
  });
  
  async function loadMediaItems() {
    try {
      loading = true;
      mediaItems = await fetchMediaItems(filters);
      updateAvailableLabels();
      loading = false;
      
      // Force refresh the timeline after data loads
      setTimeout(() => {
        if (timelineViewerComponent && typeof timelineViewerComponent.refreshTimeline === 'function') {
          timelineViewerComponent.refreshTimeline();
        }
      }, 100);
    } catch (err) {
      loading = false;
      error = 'Failed to load media items';
      console.error(error, err);
    }
  }
  
  function updateAvailableLabels() {
    const labelSet = new Set<string>();
    mediaItems.forEach(item => {
      item.labels.forEach(label => labelSet.add(label));
    });
    availableLabels = Array.from(labelSet).sort();
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
    if (timelineViewerComponent && typeof timelineViewerComponent.centerOnPlayhead === 'function') {
      timelineViewerComponent.centerOnPlayhead();
    }
  }
  
  function applyFilters() {
    filters = {};
    
    if (startDate) {
      filters.startDate = new Date(startDate).toISOString();
    }
    
    if (endDate) {
      filters.endDate = new Date(endDate).toISOString();
    }
    
    if (labelFilter.trim()) {
      filters.labels = labelFilter.split(',').map(label => label.trim()).filter(label => label);
    }
    
    loadMediaItems();
  }
  
  function clearFilters() {
    startDate = '';
    endDate = '';
    labelFilter = '';
    filters = {};
    loadMediaItems();
  }
  
  function toggleFilters() {
    showFilters = !showFilters;
  }
</script>

<main>
  <header>
    <h1>Timeline Media Viewer</h1>
  </header>
  
  <div class="container">
    <!-- Filter controls -->
    <div class="filter-section">
      <div class="filter-header">
        <button class="filter-toggle" on:click={toggleFilters}>
          {showFilters ? '▼' : '▶'} Filters
        </button>
        {#if Object.keys(filters).length > 0}
          <span class="filter-indicator">({Object.keys(filters).length} active)</span>
        {/if}
      </div>
      
      {#if showFilters}
        <div class="filter-controls">
          <div class="filter-row">
            <div class="filter-group">
              <label for="start-date">Start Date:</label>
              <input 
                id="start-date"
                type="date" 
                bind:value={startDate}
                class="filter-input"
              />
            </div>
            
            <div class="filter-group">
              <label for="end-date">End Date:</label>
              <input 
                id="end-date"
                type="date" 
                bind:value={endDate}
                class="filter-input"
              />
            </div>
          </div>
          
          <div class="filter-row">
            <div class="filter-group full-width">
              <label for="label-filter">Labels (comma-separated):</label>
              <input 
                id="label-filter"
                type="text" 
                bind:value={labelFilter}
                placeholder="e.g., meeting, personal, work"
                class="filter-input"
              />
              {#if availableLabels.length > 0}
                <div class="available-labels">
                  <span class="label-hint">Available: </span>
                  {#each availableLabels as label}
                    <button 
                      class="label-tag" 
                      on:click={() => {
                        if (!labelFilter.includes(label)) {
                          labelFilter = labelFilter ? `${labelFilter}, ${label}` : label;
                        }
                      }}
                    >
                      {label}
                    </button>
                  {/each}
                </div>
              {/if}
            </div>
          </div>
          
          <div class="filter-actions">
            <button class="apply-btn" on:click={applyFilters}>Apply Filters</button>
            <button class="clear-btn" on:click={clearFilters}>Clear All</button>
          </div>
        </div>
      {/if}
    </div>

    <!-- Timeline viewer with tabbed interface always visible at the top -->
    <div class="timeline-container">
      <TimelineViewer 
        data={mediaItems} 
        {loading}
        {error}
        on:item-select={handleItemSelect}
        on:center-playhead
        bind:this={timelineViewerComponent}
      />
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
    min-height: 400px;
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
  
  /* Filter styles */
  .filter-section {
    margin-bottom: 1rem;
    border: 1px solid #eee;
    border-radius: 4px;
    background: #fafafa;
  }
  
  .filter-header {
    padding: 0.75rem 1rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    border-bottom: 1px solid #eee;
  }
  
  .filter-toggle {
    background: none;
    border: none;
    cursor: pointer;
    font-size: 1rem;
    color: #666;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  
  .filter-toggle:hover {
    color: #2196f3;
  }
  
  .filter-indicator {
    color: #2196f3;
    font-size: 0.9rem;
    font-weight: 500;
  }
  
  .filter-controls {
    padding: 1rem;
    background: white;
  }
  
  .filter-row {
    display: flex;
    gap: 1rem;
    margin-bottom: 1rem;
    flex-wrap: wrap;
  }
  
  .filter-row:last-child {
    margin-bottom: 0;
  }
  
  .filter-group {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    min-width: 200px;
  }
  
  .filter-group.full-width {
    flex: 1;
    min-width: 300px;
  }
  
  .filter-group label {
    font-size: 0.9rem;
    font-weight: 500;
    color: #555;
  }
  
  .filter-input {
    padding: 0.5rem;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 1rem;
  }
  
  .filter-input:focus {
    outline: none;
    border-color: #2196f3;
    box-shadow: 0 0 0 2px rgba(33, 150, 243, 0.1);
  }
  
  .available-labels {
    margin-top: 0.5rem;
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
    align-items: center;
  }
  
  .label-hint {
    font-size: 0.8rem;
    color: #666;
    margin-right: 0.5rem;
  }
  
  .label-tag {
    background: #e3f2fd;
    color: #1976d2;
    border: 1px solid #bbdefb;
    border-radius: 12px;
    padding: 0.25rem 0.5rem;
    font-size: 0.8rem;
    cursor: pointer;
    transition: all 0.2s ease;
  }
  
  .label-tag:hover {
    background: #bbdefb;
    transform: translateY(-1px);
  }
  
  .filter-actions {
    display: flex;
    gap: 0.5rem;
    margin-top: 1rem;
    padding-top: 1rem;
    border-top: 1px solid #eee;
  }
  
  .apply-btn, .clear-btn {
    padding: 0.5rem 1rem;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.9rem;
    transition: all 0.2s ease;
  }
  
  .apply-btn {
    background: #2196f3;
    color: white;
  }
  
  .apply-btn:hover {
    background: #1976d2;
  }
  
  .clear-btn {
    background: #f5f5f5;
    color: #666;
    border: 1px solid #ddd;
  }
  
  .clear-btn:hover {
    background: #eeeeee;
  }
</style>