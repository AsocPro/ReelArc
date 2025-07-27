<script lang="ts">
  import { onMount, createEventDispatcher, onDestroy } from 'svelte';
  import { DataSet, Timeline } from 'vis-timeline/standalone';
  import 'vis-timeline/styles/vis-timeline-graph2d.css';
  import type { MediaItem, TimelineItem, ZoomLevel } from '../lib/types';

  import { mediaPlayback } from '../lib/stores';
  import { get } from 'svelte/store';
  
  export let data: MediaItem[] = [];
  export let loading = false;
  export let error = '';
  
  const dispatch = createEventDispatcher<{
    'item-select': MediaItem;
    'center-playhead': void;
  }>();
  
  let container: HTMLElement;
  let timeline: any;
  let timelineItems: TimelineItem[] = [];
  let playheadLine: any = null;
  let currentTimeMarker: any = null;
  let unsubscribe: () => void;
  let currentTimeInterval: any = null;
  let currentZoomLevel: ZoomLevel;
  
  // Zoom level configuration
  const zoomLevels: ZoomLevel[] = [
    { id: 'hour', label: '1 Hour', duration: 60 * 60 * 1000, snapTo: 'hour' },
    { id: 'day', label: '1 Day', duration: 24 * 60 * 60 * 1000, snapTo: 'day' },
    { id: 'week', label: '1 Week', duration: 7 * 24 * 60 * 60 * 1000, snapTo: 'week' },
    { id: 'month', label: '1 Month', duration: 30 * 24 * 60 * 60 * 1000, snapTo: 'month' },
    { id: 'year', label: '1 Year', duration: 365 * 24 * 60 * 60 * 1000, snapTo: 'year' }
  ];
  
  // Initialize with day view
  currentZoomLevel = zoomLevels[1] as ZoomLevel;
  
  // Watch for changes in data and container
  $: if (data.length > 0 && container && !timeline) {
    // Initialize timeline if container is available but timeline isn't created yet
    initializeTimeline();
  } else if (data.length > 0 && timeline) {
    // Update existing timeline with new data
    convertToTimelineItems();
    updateTimeline();
  }
  
  // Watch for container changes
  $: if (container && data.length > 0 && !timeline) {
    initializeTimeline();
  }
  
  onMount(() => {
    // Wait for the next tick to ensure container is rendered
    setTimeout(() => {
      if (container && data.length > 0) {
        initializeTimeline();
      }
    }, 0);
    
    // Subscribe to media playback store to update playhead
    unsubscribe = mediaPlayback.subscribe(state => {
      if (timeline && container) {
        if (state.isPlaying && !playheadLine) {
          // Create playhead marker if it doesn't exist
          createPlayheadMarker();
        }
        
        // Update playhead position
        updatePlayheadPosition(state);
      }
    });
    
    // Start current time marker updates
    startCurrentTimeUpdates();
  });
  
  onDestroy(() => {
    if (timeline) {
      timeline.destroy();
    }
    
    // Unsubscribe from store when component is destroyed
    if (unsubscribe) {
      unsubscribe();
    }
    
    // Clear current time interval
    if (currentTimeInterval) {
      clearInterval(currentTimeInterval);
    }
  });
  
  function initializeTimeline() {
    if (!container) {
      console.error('Timeline container element is not available');
      return;
    }
    
    // Don't reinitialize if timeline already exists
    if (timeline) {
      return;
    }
    
    // Initialize an empty timeline
    const options = {
      height: '100%',
      minHeight: '400px',
      stack: true,
      showCurrentTime: true,
      zoomable: true,
      zoomMin: 1000 * 60 , // one minute in milliseconds
      zoomMax: 1000 * 60 * 60 * 24 * 365, // about one year in milliseconds
      type: 'box',
      orientation: {
        axis: 'top',
        item: 'top'
      }
    };
    
    try {
      // Create the timeline instance
      timeline = new Timeline(container, [], options);
      
      timeline.on('select', function(properties: any) {
        if (properties.items && properties.items.length) {
          const selectedId = properties.items[0];
          const selectedItem = data.find(item => item.id === selectedId);
          if (selectedItem) {
            dispatch('item-select', selectedItem);
          }
        }
      });
      
      // Add data to the timeline if available
      if (data.length > 0) {
        convertToTimelineItems();
        updateTimeline();
      }
      
      // Timeline is now initialized
      
      // Create current time marker
      createCurrentTimeMarker();
      
      // Check if we need to create a playhead marker
      const playbackState = get(mediaPlayback);
      if (playbackState.isPlaying) {
        createPlayheadMarker();
        updatePlayheadPosition(playbackState);
      }
    } catch (err) {
      console.error('Failed to initialize timeline:', err);
    }
  }
  
  function convertToTimelineItems() {
    timelineItems = data.map(item => {
      const timelineItem: TimelineItem = {
        id: item.id,
        content: item.filename,
        start: item.timestamp,
        type: getTimelineItemType(item.type),
        className: `item-${item.type}`,
        mediaItem: item
      };
      
      // Add end time for range items (audio/video)
      if (item.duration) {
        const startDate = new Date(item.timestamp);
        const endDate = new Date(startDate.getTime() + item.duration * 1000);
        timelineItem.end = endDate.toISOString();
      } else {
        timelineItem.end = item.timestamp;
      }

      return timelineItem;
    });
  }
  
  function getTimelineItemType(mediaType: string): string {
    switch (mediaType) {
      case 'photo':
        return 'box';
      case 'audio':
      case 'video':
        return 'range';
      default:
        return 'box';
    }
  }
  
  function updateTimeline() {
    if (!timeline) {
      console.error('Cannot update timeline: timeline is not initialized');
      return;
    }
    
    // Convert data to format expected by vis-timeline
    const items = new DataSet(timelineItems);
    
    try {
      timeline.setItems(items);
      timeline.fit();
    } catch (err) {
      console.error('Error updating timeline:', err);
    }
  }
  
  // Create a custom playhead marker line
  function createPlayheadMarker() {
    if (!timeline) return;
    
    // Remove existing playhead if it exists
    if (playheadLine) {
      timeline.removeCustomTime(playheadLine);
      playheadLine = null;
    }
    
    // Create a new custom time line
    playheadLine = 'playhead-' + Date.now(); // Unique ID
    try {
      timeline.addCustomTime(new Date(), playheadLine);
      timeline.setCustomTimeMarker('Media Playhead', playheadLine, false);
      
      // Style the playhead line
      const playheadElement = container.querySelector(`.vis-custom-time.${playheadLine}`);
      if (playheadElement) {
        playheadElement.classList.add('playhead-marker');
      }
    } catch (err) {
      console.error('Error creating playhead marker:', err);
    }
  }
  
  // Update the playhead position based on media playback
  function updatePlayheadPosition(playbackState: any) {
    if (!timeline || !playheadLine || !playbackState.isPlaying || !playbackState.startTimestamp) return;
    
    try {
      // Calculate the current timestamp based on the start timestamp and current playback time
      const startTime = new Date(playbackState.startTimestamp);
      const currentTime = new Date(startTime.getTime() + (playbackState.currentTime * 1000));
      
      // Update the custom time line position
      timeline.setCustomTime(currentTime, playheadLine);
    } catch (err) {
      console.error('Error updating playhead position:', err);
    }
  }
  
  // Center the timeline view on the playhead
  export function centerOnPlayhead() {
    if (!timeline || !playheadLine) return;
    
    try {
      const playheadTime = timeline.getCustomTime(playheadLine);
      if (playheadTime) {
        // Center the timeline on the playhead time with some padding
        const windowTime = 60 * 1000; // 1 minute window
        const start = new Date(playheadTime.getTime() - windowTime / 2);
        const end = new Date(playheadTime.getTime() + windowTime / 2);
        
        timeline.setWindow(start, end, { animation: true });
        dispatch('center-playhead');
      }
    } catch (err) {
      console.error('Error centering on playhead:', err);
    }
  }
  
  // Start updating current time marker
  function startCurrentTimeUpdates() {
    // Create initial current time marker
    createCurrentTimeMarker();
    
    // Update every 30 seconds
    currentTimeInterval = setInterval(() => {
      updateCurrentTimeMarker();
    }, 30000);
  }
  
  // Create current time marker
  function createCurrentTimeMarker() {
    if (!timeline) return;
    
    // Remove existing marker if it exists
    if (currentTimeMarker) {
      timeline.removeCustomTime(currentTimeMarker);
      currentTimeMarker = null;
    }
    
    // Create a new current time marker
    currentTimeMarker = 'current-time-' + Date.now();
    try {
      timeline.addCustomTime(new Date(), currentTimeMarker);
      timeline.setCustomTimeMarker('Current Time', currentTimeMarker, false);
      
      // Style the current time marker
      const markerElement = container.querySelector(`.vis-custom-time.${currentTimeMarker}`);
      if (markerElement) {
        markerElement.classList.add('current-time-marker');
      }
    } catch (err) {
      console.error('Error creating current time marker:', err);
    }
  }
  
  // Update current time marker position
  function updateCurrentTimeMarker() {
    if (!timeline || !currentTimeMarker) return;
    
    try {
      timeline.setCustomTime(new Date(), currentTimeMarker);
    } catch (err) {
      console.error('Error updating current time marker:', err);
    }
  }
  
  // Snap time to appropriate interval based on zoom level
  function snapTime(date: Date, snapTo: string): Date {
    const snapped = new Date(date);
    
    switch (snapTo) {
      case 'hour':
        snapped.setMinutes(0, 0, 0);
        break;
      case 'day':
        snapped.setHours(0, 0, 0, 0);
        break;
      case 'week':
        const dayOfWeek = snapped.getDay();
        const diff = snapped.getDate() - dayOfWeek;
        snapped.setDate(diff);
        snapped.setHours(0, 0, 0, 0);
        break;
      case 'month':
        snapped.setDate(1);
        snapped.setHours(0, 0, 0, 0);
        break;
      case 'year':
        snapped.setMonth(0, 1);
        snapped.setHours(0, 0, 0, 0);
        break;
    }
    
    return snapped;
  }
  
  // Set zoom level and update view
  function setZoomLevel(zoomLevel: ZoomLevel) {
    if (!timeline) return;
    
    currentZoomLevel = zoomLevel;
    
    // Get current center time or use current time
    const currentWindow = timeline.getWindow();
    const centerTime = new Date((currentWindow.start.getTime() + currentWindow.end.getTime()) / 2);
    
    // Snap to appropriate interval
    const snappedStart = snapTime(centerTime, zoomLevel.snapTo);
    const snappedEnd = new Date(snappedStart.getTime() + zoomLevel.duration);
    
    // Set the new window
    timeline.setWindow(snappedStart, snappedEnd, { animation: true });
  }
  
  // Navigate to previous time period
  function navigatePrevious() {
    if (!timeline) return;
    
    const currentWindow = timeline.getWindow();
    const newStart = new Date(currentWindow.start.getTime() - currentZoomLevel.duration);
    
    // Snap the new start time
    const snappedStart = snapTime(newStart, currentZoomLevel.snapTo);
    const snappedEnd = new Date(snappedStart.getTime() + currentZoomLevel.duration);
    
    timeline.setWindow(snappedStart, snappedEnd, { animation: true });
  }
  
  // Navigate to next time period
  function navigateNext() {
    if (!timeline) return;
    
    const currentWindow = timeline.getWindow();
    const newStart = new Date(currentWindow.start.getTime() + currentZoomLevel.duration);
    
    // Snap the new start time
    const snappedStart = snapTime(newStart, currentZoomLevel.snapTo);
    const snappedEnd = new Date(snappedStart.getTime() + currentZoomLevel.duration);
    
    timeline.setWindow(snappedStart, snappedEnd, { animation: true });
  }
  
  // Jump to current time
  function jumpToNow() {
    if (!timeline) return;
    
    const now = new Date();
    const snappedStart = snapTime(now, currentZoomLevel.snapTo);
    const snappedEnd = new Date(snappedStart.getTime() + currentZoomLevel.duration);
    
    timeline.setWindow(snappedStart, snappedEnd, { animation: true });
  }
</script>

<div class="timeline-container">
  {#if loading}
    <div class="loading">Loading timeline data...</div>
  {:else if error}
    <div class="error">{error}</div>
  {:else if data.length === 0}
    <div class="empty">No media items found</div>
  {:else}
    <div class="timeline-controls">
      <div class="zoom-controls">
        <span class="control-label">Zoom:</span>
        {#each zoomLevels as zoomLevel}
          <button 
            class="zoom-btn" 
            class:active={currentZoomLevel.id === zoomLevel.id}
            on:click={() => setZoomLevel(zoomLevel)}
          >
            {zoomLevel.label}
          </button>
        {/each}
      </div>
      
      <div class="navigation-controls">
        <button class="nav-btn" on:click={navigatePrevious} title="Previous {currentZoomLevel.label}">
          ← Previous
        </button>
        <button class="nav-btn now-btn" on:click={jumpToNow} title="Jump to current time">
          Now
        </button>
        <button class="nav-btn" on:click={navigateNext} title="Next {currentZoomLevel.label}">
          Next →
        </button>
      </div>
    </div>
    <div class="timeline" bind:this={container}></div>
  {/if}
</div>

<style>
  .timeline-container {
    width: 100%;
    height: 100%;
    min-height: 400px;
    border: none;
    overflow: hidden;
    position: relative;
  }
  
  .timeline {
    width: 100%;
    height: 100%;
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
  
  :global(.item-photo) {
    background-color: #aed581 !important;
    border-color: #8BC34A !important;
  }
  
  :global(.item-audio) {
    background-color: #6ecfff !important;
    border-color: #2196F3 !important;
  }
  
  :global(.item-video) {
    background-color: #ffab91 !important;
    border-color: #FF5722 !important;
  }
  
  :global(.item-default) {
    background-color: #e0e0e0 !important;
    border-color: #9e9e9e !important;
  }
  
  :global(.playhead-marker) {
    border-color: #ff5722 !important;
    border-width: 2px !important;
    z-index: 10 !important;
  }
  
  :global(.vis-custom-time.playhead-marker::after) {
    content: '';
    position: absolute;
    top: 0;
    left: -5px;
    width: 10px;
    height: 10px;
    background-color: #ff5722;
    border-radius: 50%;
  }
  
  :global(.current-time-marker) {
    border-color: #2196f3 !important;
    border-width: 2px !important;
    z-index: 9 !important;
  }
  
  :global(.vis-custom-time.current-time-marker::after) {
    content: '';
    position: absolute;
    top: 0;
    left: -4px;
    width: 8px;
    height: 8px;
    background-color: #2196f3;
    border-radius: 50%;
  }
  
  .timeline-controls {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
    background-color: #f5f5f5;
    border-bottom: 1px solid #e0e0e0;
    gap: 16px;
    flex-wrap: wrap;
  }
  
  .zoom-controls {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }
  
  .control-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: #666;
    margin-right: 4px;
  }
  
  .zoom-btn {
    padding: 4px 12px;
    border: 1px solid #ddd;
    background-color: #fff;
    border-radius: 4px;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s ease;
  }
  
  .zoom-btn:hover {
    background-color: #f0f0f0;
    border-color: #bbb;
  }
  
  .zoom-btn.active {
    background-color: #2196f3;
    color: white;
    border-color: #2196f3;
  }
  
  .navigation-controls {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  
  .nav-btn {
    padding: 6px 12px;
    border: 1px solid #ddd;
    background-color: #fff;
    border-radius: 4px;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s ease;
  }
  
  .nav-btn:hover {
    background-color: #f0f0f0;
    border-color: #bbb;
  }
  
  .now-btn {
    background-color: #4caf50;
    color: white;
    border-color: #4caf50;
  }
  
  .now-btn:hover {
    background-color: #45a049;
    border-color: #45a049;
  }
</style>
