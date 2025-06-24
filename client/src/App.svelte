<script lang="ts">
  import { onMount } from 'svelte';
  import Timeline from './components/Timeline.svelte';
  import MediaViewer from './components/MediaViewer.svelte';
  import UploadForm from './components/UploadForm.svelte';
  
  let timelineData: any[] = [];
  let selectedItem: any = null;
  
  onMount(async () => {
    try {
      const response = await fetch('/api/timeline');
      if (response.ok) {
        timelineData = await response.json();
      } else {
        console.error('Failed to fetch timeline data');
      }
    } catch (error) {
      console.error('Error fetching timeline data:', error);
    }
  });
  
  function handleItemSelect(event: CustomEvent) {
    selectedItem = event.detail;
  }
  
  function handleUploadSuccess(event: CustomEvent) {
    // Refresh timeline data after successful upload
    fetch('/api/timeline')
      .then(response => response.json())
      .then(data => {
        timelineData = data;
      })
      .catch(error => {
        console.error('Error refreshing timeline data:', error);
      });
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
    
    <div class="timeline-section">
      <h2>Media Timeline</h2>
      <Timeline 
        data={timelineData} 
        on:item-select={handleItemSelect} 
      />
    </div>
    
    <div class="viewer-section">
      <h2>Media Viewer</h2>
      {#if selectedItem}
        <MediaViewer item={selectedItem} />
      {:else}
        <p>Select an item from the timeline to view</p>
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
    display: grid;
    grid-template-columns: 1fr;
    grid-gap: 2rem;
  }
  
  @media (min-width: 768px) {
    .container {
      grid-template-columns: 1fr 2fr;
    }
    
    .timeline-section {
      grid-column: 1 / -1;
    }
  }
  
  h2 {
    margin-top: 0;
    color: #555;
  }
</style>