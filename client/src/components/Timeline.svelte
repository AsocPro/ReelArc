<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  import { DataSet, Timeline } from 'vis-timeline/standalone';
  import 'vis-timeline/styles/vis-timeline-graph2d.css';
  import type { MediaItem, TimelineItem } from '../lib/types';
  import { fetchMediaItems } from '../lib/api';
  
  export let data: MediaItem[] = [];
  
  const dispatch = createEventDispatcher<{
    'item-select': MediaItem;
  }>();
  
  let container: HTMLElement;
  let timeline: any;
  let timelineItems: TimelineItem[] = [];
  let loading = true;
  let error = '';
  
  // Watch for changes in data and container
  $: if (data.length > 0 && container && !timeline) {
    // Initialize timeline if container is available but timeline isn't created yet
    initializeTimeline();
  } else if (data.length > 0 && timeline) {
    // Update existing timeline with new data
    convertToTimelineItems();
    updateTimeline();
    loading = false; // Ensure loading is set to false when data is available
  }
  
  // Watch for container changes
  $: if (container && data.length > 0 && !timeline) {
    initializeTimeline();
  }
  
  onMount(async () => {
    // If no data is provided, fetch it
    if (data.length === 0) {
      try {
        loading = true;
        data = await fetchMediaItems();
      } catch (err) {
        loading = false;
        error = 'Failed to load media items';
        console.error(error, err);
      }
    }
    
    // Wait for the next tick to ensure container is rendered
    setTimeout(() => {
      if (container && data.length > 0) {
        initializeTimeline();
      } else if (data.length > 0) {
        // If we have data but no container yet, we'll wait for the container to be bound
        loading = false;
      }
    }, 0);
    
    return () => {
      if (timeline) {
        timeline.destroy();
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
      height: '300px',
      minHeight: '300px',
      stack: true,
      showCurrentTime: true,
      zoomable: true,
      zoomMin: 1000 * 60 , // one minute in milliseconds
      zoomMax: 1000 * 60 * 60 * 24 , // about one day in milliseconds
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
      
      // Set loading to false now that timeline is initialized
      loading = false;
    } catch (err) {
      console.error('Failed to initialize timeline:', err);
      error = 'Failed to initialize timeline';
      loading = false; // Set loading to false even if there's an error
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
    height: 300px;
    margin: 1rem 0;
    border: 1px solid #ddd;
    border-radius: 4px;
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
</style>
