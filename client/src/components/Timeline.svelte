<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  import { DataSet, Timeline } from 'vis-timeline/standalone';
  import 'vis-timeline/styles/vis-timeline-graph2d.css';
  import type { MediaItem, TimelineItem } from '../lib/types';

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
  let unsubscribe: () => void;
  
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
    
    // Return a cleanup function
    return () => {
      if (timeline) {
        timeline.destroy();
      }
      
      // Unsubscribe from store when component is destroyed
      if (unsubscribe) {
        unsubscribe();
      }
    };
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
</script>

<div class="timeline-container">
  {#if loading}
    <div class="loading">Loading timeline data...</div>
  {:else if error}
    <div class="error">{error}</div>
  {:else if data.length === 0}
    <div class="empty">No media items found</div>
  {:else}
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
</style>
