<script lang="ts">
  export let item: any;
  
  $: mediaType = item?.type || 'unknown';
  $: mediaPath = item?.mediaPath || '';
</script>

<div class="media-viewer">
  <div class="media-info">
    <h3>{item.content}</h3>
    <p>Type: {mediaType}</p>
    <p>Start: {new Date(item.start).toLocaleString()}</p>
    {#if item.end}
      <p>End: {new Date(item.end).toLocaleString()}</p>
    {/if}
  </div>
  
  <div class="media-content">
    {#if mediaType === 'image'}
      <img src={mediaPath} alt={item.content} />
    {:else if mediaType === 'video'}
      <video controls>
        <source src={mediaPath} type="video/mp4">
        Your browser does not support the video tag.
      </video>
    {:else if mediaType === 'audio'}
      <audio controls>
        <source src={mediaPath} type="audio/mpeg">
        Your browser does not support the audio tag.
      </audio>
    {:else}
      <div class="placeholder">
        <p>Preview not available for this media type</p>
      </div>
    {/if}
  </div>
</div>

<style>
  .media-viewer {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding: 1rem;
    border: 1px solid #ddd;
    border-radius: 4px;
  }
  
  .media-info {
    padding-bottom: 1rem;
    border-bottom: 1px solid #eee;
  }
  
  .media-info h3 {
    margin-top: 0;
  }
  
  .media-content {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 200px;
  }
  
  img, video, audio {
    max-width: 100%;
    max-height: 400px;
  }
  
  .placeholder {
    display: flex;
    justify-content: center;
    align-items: center;
    width: 100%;
    height: 200px;
    background-color: #f5f5f5;
    border-radius: 4px;
  }
</style>