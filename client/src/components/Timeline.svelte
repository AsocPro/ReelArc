<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  import { DataSet, Timeline } from 'vis-timeline/standalone';
  import 'vis-timeline/styles/vis-timeline-graph2d.css';
  
  export let data: any[] = [];
  
  const dispatch = createEventDispatcher();
  let container: HTMLElement;
  let timeline: any;
  
  $: if (data.length > 0 && timeline) {
    updateTimeline();
  }
  
  onMount(() => {
    // Initialize an empty timeline
    const options = {
      height: '300px',
      minHeight: '300px',
      stack: true,
      showCurrentTime: true,
      zoomable: true,
      zoomMin: 1000 * 60 * 60 * 24, // one day in milliseconds
      zoomMax: 1000 * 60 * 60 * 24 * 31 * 3, // about three months in milliseconds
      type: 'box',
      orientation: {
        axis: 'top',
        item: 'top'
      }
    };
    
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
    
    if (data.length > 0) {
      updateTimeline();
    }
    
    return () => {
      if (timeline) {
        timeline.destroy();
      }
    };
  });
  
  function updateTimeline() {
    // Convert data to format expected by vis-timeline
    const items = new DataSet(
      data.map(item => ({
        id: item.id,
        content: item.content,
        start: item.start,
        end: item.end,
        type: item.type === 'image' ? 'box' : 'range',
        className: `item-${item.type || 'default'}`
      }))
    );
    
    timeline.setItems(items);
    timeline.fit();
  }
</script>

<div class="timeline-container" bind:this={container}></div>

<style>
  .timeline-container {
    width: 100%;
    height: 300px;
    margin: 1rem 0;
    border: 1px solid #ddd;
    border-radius: 4px;
    overflow: hidden;
  }
  
  :global(.item-audio) {
    background-color: #6ecfff !important;
    border-color: #2196F3 !important;
  }
  
  :global(.item-video) {
    background-color: #ffab91 !important;
    border-color: #FF5722 !important;
  }
  
  :global(.item-image) {
    background-color: #aed581 !important;
    border-color: #8BC34A !important;
  }
  
  :global(.item-default) {
    background-color: #e0e0e0 !important;
    border-color: #9e9e9e !important;
  }
</style>