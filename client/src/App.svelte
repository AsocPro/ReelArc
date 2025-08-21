<script lang="ts">
  import { onMount } from 'svelte';
  import TimelineViewer from './components/TimelineViewer.svelte';

  import UploadForm from './components/UploadForm.svelte';
  import TranscriptionStatus from './components/TranscriptionStatus.svelte';
  import MediaDetails from './components/MediaDetails.svelte';
  import type { MediaItem, MediaFilters, FiltersConfig, Filter } from './lib/types';
  import { fetchMediaItems, fetchFiltersConfig } from './lib/api';
  
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
  let mediaTypeFilter: string[] = [];
  let availableLabels: string[] = [];
  let showFilters = false;
  let activeQuickFilter: string | null = null;
  let filtersConfig: FiltersConfig | null = null;
  let availableMediaTypes = ['photo', 'audio', 'video', 'note'];
  
  // Relative filtering state
  let dateFilterType: 'fixed' | 'relative' = 'fixed';
  let relativeRangeType: 'simple' | 'range' = 'simple';
  // Simple relative (single period)
  let relativePeriod: string = 'days';
  let relativeOffset: number = -7;
  let relativeCount: number = 7;
  let relativeDirection: string = 'backward';
  // Start relative date (for ranges)
  let startRelativePeriod: string = 'days';
  let startRelativeOffset: number = -7;
  let startRelativeCount: number = 7;
  let startRelativeDirection: string = 'backward';
  // End relative date (for ranges)
  let endRelativePeriod: string = 'days';
  let endRelativeOffset: number = 0;
  let endRelativeCount: number = 1;
  let endRelativeDirection: string = 'backward';
  
  onMount(async () => {
    await loadFiltersConfig();
    await loadMediaItems();
  });
  
  async function loadFiltersConfig() {
    try {
      filtersConfig = await fetchFiltersConfig();
      // Set default filter if one is specified
      if (filtersConfig && filtersConfig.defaultFilter) {
        activeQuickFilter = filtersConfig.defaultFilter;
        filters = { filter: filtersConfig.defaultFilter };
      }
    } catch (err) {
      console.error('Failed to load filters config:', err);
    }
  }
  
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
    
    // Handle date filtering based on type
    if (dateFilterType === 'fixed') {
      if (startDate) {
        filters.startDate = new Date(startDate).toISOString();
      }
      if (endDate) {
        filters.endDate = new Date(endDate).toISOString();
      }
    } else if (dateFilterType === 'relative') {
      // Send relative date parameters directly to server
      filters.dateRangeType = 'relative';
      
      if (relativeRangeType === 'simple') {
        // Simple relative date (backward compatibility)
        filters.startPeriod = relativePeriod;
        filters.startAnchor = 'start';
        
        if (relativePeriod === 'days') {
          filters.startCount = relativeCount;
          filters.startDirection = relativeDirection;
        } else {
          filters.startOffset = relativeOffset;
        }
      } else if (relativeRangeType === 'range') {
        // Relative date range with start and end
        // Start date
        filters.startPeriod = startRelativePeriod;
        filters.startAnchor = 'start';
        if (startRelativePeriod === 'days') {
          filters.startCount = startRelativeCount;
          filters.startDirection = startRelativeDirection;
        } else {
          filters.startOffset = startRelativeOffset;
        }
        
        // End date
        filters.endPeriod = endRelativePeriod;
        filters.endAnchor = 'start';
        if (endRelativePeriod === 'days') {
          filters.endCount = endRelativeCount;
          filters.endDirection = endRelativeDirection;
        } else {
          filters.endOffset = endRelativeOffset;
        }
      }
    }
    
    if (labelFilter.trim()) {
      filters.labels = labelFilter.split(',').map(label => label.trim()).filter(label => label);
    }
    
    if (mediaTypeFilter.length > 0) {
      filters.mediaTypes = mediaTypeFilter;
    }
    
    // Clear active quick filter since we're using manual filters
    activeQuickFilter = null;
    
    loadMediaItems();
  }
  
  function clearFilters() {
    startDate = '';
    endDate = '';
    labelFilter = '';
    mediaTypeFilter = [];
    activeQuickFilter = null;
    dateFilterType = 'fixed';
    relativeRangeType = 'simple';
    // Simple relative
    relativePeriod = 'days';
    relativeOffset = -7;
    relativeCount = 7;
    relativeDirection = 'backward';
    // Start relative
    startRelativePeriod = 'days';
    startRelativeOffset = -7;
    startRelativeCount = 7;
    startRelativeDirection = 'backward';
    // End relative
    endRelativePeriod = 'days';
    endRelativeOffset = 0;
    endRelativeCount = 1;
    endRelativeDirection = 'backward';
    filters = {};
    loadMediaItems();
  }
  
  function applyQuickFilter(filterId: string) {
    activeQuickFilter = filterId;
    // Clear manual filters
    startDate = '';
    endDate = '';
    labelFilter = '';
    mediaTypeFilter = [];
    filters = { filter: filterId };
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
        {#if Object.keys(filters).length > 0 || activeQuickFilter}
          <span class="filter-indicator">
            {#if activeQuickFilter}
              Quick filter: {filtersConfig?.filters.find(f => f.id === activeQuickFilter)?.name || activeQuickFilter}
            {:else}
              {Object.keys(filters).length} manual filters active
            {/if}
          </span>
        {/if}
      </div>
      
      {#if showFilters}
        <div class="filter-controls">
          <!-- Quick Filters -->
          {#if filtersConfig && filtersConfig.filters.length > 0}
            <div class="quick-filters-section">
              <h3>Quick Filters</h3>
              <div class="quick-filters">
                {#each filtersConfig.filters.filter(f => f.enabled) as filter}
                  <button 
                    class="quick-filter-btn"
                    class:active={activeQuickFilter === filter.id}
                    on:click={() => applyQuickFilter(filter.id)}
                    title={filter.description}
                  >
                    <span class="filter-icon">{filter.icon}</span>
                    <span class="filter-name">{filter.name}</span>
                  </button>
                {/each}
              </div>
            </div>
            
            <div class="divider">
              <span>or use manual filters</span>
            </div>
          {/if}
          
          <!-- Manual Filters -->
          <div class="manual-filters-section">
            <!-- Date Filter Type Selection -->
            <div class="filter-row">
              <div class="filter-group full-width">
                <label>Date Filter Type:</label>
                <div class="radio-group">
                  <label class="radio-label">
                    <input 
                      type="radio" 
                      bind:group={dateFilterType} 
                      value="fixed"
                    />
                    <span>Fixed Dates</span>
                  </label>
                  <label class="radio-label">
                    <input 
                      type="radio" 
                      bind:group={dateFilterType} 
                      value="relative"
                    />
                    <span>Relative Dates</span>
                  </label>
                </div>
              </div>
            </div>
            
            <!-- Fixed Date Filters -->
            {#if dateFilterType === 'fixed'}
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
            {/if}
            
            <!-- Relative Date Filters -->
            {#if dateFilterType === 'relative'}
              <!-- Relative Range Type Selection -->
              <div class="filter-row">
                <div class="filter-group full-width">
                  <label>Relative Filter Type:</label>
                  <div class="radio-group">
                    <label class="radio-label">
                      <input 
                        type="radio" 
                        bind:group={relativeRangeType} 
                        value="simple"
                      />
                      <span>Simple Relative (single period)</span>
                    </label>
                    <label class="radio-label">
                      <input 
                        type="radio" 
                        bind:group={relativeRangeType} 
                        value="range"
                      />
                      <span>Relative Range (start and end)</span>
                    </label>
                  </div>
                </div>
              </div>
              
              <!-- Simple Relative Date Filters -->
              {#if relativeRangeType === 'simple'}
                <div class="filter-row">
                  <div class="filter-group">
                    <label for="relative-period">Period:</label>
                    <select 
                      id="relative-period"
                      bind:value={relativePeriod}
                      class="filter-input"
                    >
                      <option value="day">Day</option>
                      <option value="days">Days (Range)</option>
                      <option value="week">Week</option>
                      <option value="month">Month</option>
                    </select>
                  </div>
                  
                  {#if relativePeriod === 'days'}
                    <div class="filter-group">
                      <label for="relative-count">Number of Days:</label>
                      <input 
                        id="relative-count"
                        type="number" 
                        bind:value={relativeCount}
                        min="1"
                        max="365"
                        class="filter-input"
                      />
                    </div>
                    
                    <div class="filter-group">
                      <label for="relative-direction">Direction:</label>
                      <select 
                        id="relative-direction"
                        bind:value={relativeDirection}
                        class="filter-input"
                      >
                        <option value="backward">Past (Backward)</option>
                        <option value="forward">Future (Forward)</option>
                      </select>
                    </div>
                  {:else}
                    <div class="filter-group">
                      <label for="relative-offset">Offset:</label>
                      <input 
                        id="relative-offset"
                        type="number" 
                        bind:value={relativeOffset}
                        class="filter-input"
                        placeholder="0 = current, -1 = previous, 1 = next"
                      />
                    </div>
                    
                    <div class="filter-group">
                      <span class="helper-text">
                        {#if relativeOffset === 0}
                          Current {relativePeriod}
                        {:else if relativeOffset < 0}
                          {Math.abs(relativeOffset)} {relativePeriod}(s) ago
                        {:else}
                          {relativeOffset} {relativePeriod}(s) from now
                        {/if}
                      </span>
                    </div>
                  {/if}
                </div>
              {/if}
              
              <!-- Relative Range Date Filters -->
              {#if relativeRangeType === 'range'}
                <div class="relative-range-section">
                  <!-- Start Date -->
                  <div class="range-section">
                    <h4>Start Date</h4>
                    <div class="filter-row">
                      <div class="filter-group">
                        <label for="start-relative-period">Period:</label>
                        <select 
                          id="start-relative-period"
                          bind:value={startRelativePeriod}
                          class="filter-input"
                        >
                          <option value="day">Day</option>
                          <option value="days">Days</option>
                          <option value="week">Week</option>
                          <option value="month">Month</option>
                        </select>
                      </div>
                      
                      {#if startRelativePeriod === 'days'}
                        <div class="filter-group">
                          <label for="start-relative-count">Days:</label>
                          <input 
                            id="start-relative-count"
                            type="number" 
                            bind:value={startRelativeCount}
                            min="1"
                            max="365"
                            class="filter-input"
                          />
                        </div>
                        
                        <div class="filter-group">
                          <label for="start-relative-direction">Direction:</label>
                          <select 
                            id="start-relative-direction"
                            bind:value={startRelativeDirection}
                            class="filter-input"
                          >
                            <option value="backward">Past</option>
                            <option value="forward">Future</option>
                          </select>
                        </div>
                      {:else}
                        <div class="filter-group">
                          <label for="start-relative-offset">Offset:</label>
                          <input 
                            id="start-relative-offset"
                            type="number" 
                            bind:value={startRelativeOffset}
                            class="filter-input"
                          />
                        </div>
                        
                        <div class="filter-group">
                          <span class="helper-text">
                            {#if startRelativeOffset === 0}
                              Current {startRelativePeriod}
                            {:else if startRelativeOffset < 0}
                              {Math.abs(startRelativeOffset)} {startRelativePeriod}(s) ago
                            {:else}
                              {startRelativeOffset} {startRelativePeriod}(s) from now
                            {/if}
                          </span>
                        </div>
                      {/if}
                    </div>
                  </div>
                  
                  <!-- End Date -->
                  <div class="range-section">
                    <h4>End Date</h4>
                    <div class="filter-row">
                      <div class="filter-group">
                        <label for="end-relative-period">Period:</label>
                        <select 
                          id="end-relative-period"
                          bind:value={endRelativePeriod}
                          class="filter-input"
                        >
                          <option value="day">Day</option>
                          <option value="days">Days</option>
                          <option value="week">Week</option>
                          <option value="month">Month</option>
                        </select>
                      </div>
                      
                      {#if endRelativePeriod === 'days'}
                        <div class="filter-group">
                          <label for="end-relative-count">Days:</label>
                          <input 
                            id="end-relative-count"
                            type="number" 
                            bind:value={endRelativeCount}
                            min="1"
                            max="365"
                            class="filter-input"
                          />
                        </div>
                        
                        <div class="filter-group">
                          <label for="end-relative-direction">Direction:</label>
                          <select 
                            id="end-relative-direction"
                            bind:value={endRelativeDirection}
                            class="filter-input"
                          >
                            <option value="backward">Past</option>
                            <option value="forward">Future</option>
                          </select>
                        </div>
                      {:else}
                        <div class="filter-group">
                          <label for="end-relative-offset">Offset:</label>
                          <input 
                            id="end-relative-offset"
                            type="number" 
                            bind:value={endRelativeOffset}
                            class="filter-input"
                          />
                        </div>
                        
                        <div class="filter-group">
                          <span class="helper-text">
                            {#if endRelativeOffset === 0}
                              Current {endRelativePeriod}
                            {:else if endRelativeOffset < 0}
                              {Math.abs(endRelativeOffset)} {endRelativePeriod}(s) ago
                            {:else}
                              {endRelativeOffset} {endRelativePeriod}(s) from now
                            {/if}
                          </span>
                        </div>
                      {/if}
                    </div>
                  </div>
                </div>
              {/if}
            {/if}
            
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
            
            <div class="filter-row">
              <div class="filter-group full-width">
                <label for="media-type-filter">Media Types:</label>
                <div class="media-type-checkboxes">
                  {#each availableMediaTypes as mediaType}
                    <label class="checkbox-label">
                      <input 
                        type="checkbox" 
                        value={mediaType}
                        bind:group={mediaTypeFilter}
                      />
                      <span class="checkbox-text">{mediaType}</span>
                    </label>
                  {/each}
                </div>
              </div>
            </div>
            
            <div class="filter-actions">
              <button class="apply-btn" on:click={applyFilters}>Apply Manual Filters</button>
              <button class="clear-btn" on:click={clearFilters}>Clear All</button>
            </div>
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
  
  /* Quick Filters Styles */
  .quick-filters-section {
    margin-bottom: 1.5rem;
  }
  
  .quick-filters-section h3 {
    margin: 0 0 0.75rem 0;
    font-size: 1rem;
    font-weight: 600;
    color: #333;
  }
  
  .quick-filters {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
  }
  
  .quick-filter-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    border: 2px solid #e0e0e0;
    border-radius: 8px;
    background: white;
    cursor: pointer;
    transition: all 0.2s ease;
    font-size: 0.9rem;
    color: #555;
    min-width: 120px;
  }
  
  .quick-filter-btn:hover {
    border-color: #2196f3;
    background: #f3f9ff;
    transform: translateY(-1px);
  }
  
  .quick-filter-btn.active {
    border-color: #2196f3;
    background: #2196f3;
    color: white;
  }
  
  .filter-icon {
    font-size: 1.2rem;
  }
  
  .filter-name {
    font-weight: 500;
  }
  
  .divider {
    margin: 1.5rem 0;
    text-align: center;
    position: relative;
  }
  
  .divider::before {
    content: '';
    position: absolute;
    top: 50%;
    left: 0;
    right: 0;
    height: 1px;
    background: #ddd;
    z-index: 1;
  }
  
  .divider span {
    background: white;
    padding: 0 1rem;
    color: #666;
    font-size: 0.9rem;
    position: relative;
    z-index: 2;
  }
  
  /* Manual Filters Styles */
  .manual-filters-section {
    /* Add some visual separation from quick filters */
  }
  
  .media-type-checkboxes {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    margin-top: 0.5rem;
  }
  
  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    font-size: 0.9rem;
    color: #555;
  }
  
  .checkbox-label input[type="checkbox"] {
    width: 16px;
    height: 16px;
    cursor: pointer;
  }
  
  .checkbox-text {
    text-transform: capitalize;
  }
  
  /* Relative Filter Styles */
  .radio-group {
    display: flex;
    gap: 1.5rem;
    margin-top: 0.5rem;
  }
  
  .radio-label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    font-size: 0.9rem;
    color: #555;
  }
  
  .radio-label input[type="radio"] {
    width: 16px;
    height: 16px;
    cursor: pointer;
  }
  
  .helper-text {
    font-size: 0.8rem;
    color: #666;
    font-style: italic;
    display: flex;
    align-items: center;
    background: #f8f9fa;
    padding: 0.5rem;
    border-radius: 4px;
    border: 1px solid #e9ecef;
  }
  
  select.filter-input {
    background: white;
    cursor: pointer;
  }
  
  input[type="number"].filter-input {
    min-width: 100px;
  }
  
  /* Relative Range Styles */
  .relative-range-section {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    margin-top: 1rem;
    padding: 1rem;
    background: #fafbfc;
    border: 1px solid #e1e5e9;
    border-radius: 6px;
  }
  
  .range-section {
    padding: 1rem;
    background: white;
    border: 1px solid #e1e5e9;
    border-radius: 4px;
  }
  
  .range-section h4 {
    margin: 0 0 1rem 0;
    font-size: 1rem;
    font-weight: 600;
    color: #2c3e50;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid #ecf0f1;
  }
</style>