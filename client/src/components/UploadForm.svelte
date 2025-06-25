<script lang="ts">
  import { createEventDispatcher, onDestroy } from 'svelte';
  
  const dispatch = createEventDispatcher();
  
  let fileInput: HTMLInputElement;
  let uploading = false;
  let error = '';
  let success = '';
  let selectedFiles: File[] = [];
  let uploadProgress: {[key: string]: number} = {};
  let overallProgress = 0;
  let uploadStartTime: number;
  let timeRemaining: string = '';
  let uploadSpeed: string = '';
  let activeXHR: XMLHttpRequest | null = null;
  
  // For time estimation
  let updateIntervalId: number;
  
  onDestroy(() => {
    if (updateIntervalId) {
      clearInterval(updateIntervalId);
    }
    
    if (activeXHR) {
      activeXHR.abort();
    }
  });
  
  function formatBytes(bytes: number, decimals: number = 1): string {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(decimals)) + ' ' + sizes[i];
  }
  
  function formatTime(seconds: number): string {
    if (!isFinite(seconds) || seconds < 0) return 'calculating...';
    
    if (seconds < 60) {
      return `${Math.round(seconds)}s`;
    } else if (seconds < 3600) {
      const minutes = Math.floor(seconds / 60);
      const remainingSeconds = Math.round(seconds % 60);
      return `${minutes}m ${remainingSeconds}s`;
    } else {
      const hours = Math.floor(seconds / 3600);
      const minutes = Math.floor((seconds % 3600) / 60);
      return `${hours}h ${minutes}m`;
    }
  }
  
  function handleFileSelect() {
    if (!fileInput.files || fileInput.files.length === 0) {
      selectedFiles = [];
      return;
    }
    
    selectedFiles = Array.from(fileInput.files);
    // Initialize progress for each file
    uploadProgress = {};
    selectedFiles.forEach(file => {
      uploadProgress[file.name] = 0;
    });
    
    overallProgress = 0;
    timeRemaining = '';
    uploadSpeed = '';
  }
  
  function updateTimeEstimates(loaded: number, total: number) {
    const currentTime = Date.now();
    const elapsedTime = (currentTime - uploadStartTime) / 1000; // in seconds
    
    if (elapsedTime > 0 && loaded > 0) {
      // Calculate upload speed in bytes per second
      const bytesPerSecond = loaded / elapsedTime;
      uploadSpeed = formatBytes(bytesPerSecond) + '/s';
      
      // Calculate time remaining
      if (bytesPerSecond > 0) {
        const remainingBytes = total - loaded;
        const remainingTime = remainingBytes / bytesPerSecond;
        timeRemaining = formatTime(remainingTime);
      }
    }
  }
  
  async function handleSubmit() {
    if (selectedFiles.length === 0) {
      error = 'Please select files to upload';
      return;
    }
    
    const formData = new FormData();
    selectedFiles.forEach(file => {
      formData.append('files', file);
    });
    
    uploading = true;
    error = '';
    success = '';
    uploadStartTime = Date.now();
    
    // Set up interval to update time estimates
    if (updateIntervalId) {
      clearInterval(updateIntervalId);
    }
    
    updateIntervalId = setInterval(() => {
      if (overallProgress > 0 && overallProgress < 100) {
        // This will update the time remaining display even when we don't get progress events
        const currentTime = Date.now();
        const elapsedTime = (currentTime - uploadStartTime) / 1000;
        if (elapsedTime > 0) {
          // Estimate total time based on current progress
          const estimatedTotalTime = elapsedTime / (overallProgress / 100);
          const remainingTime = estimatedTotalTime - elapsedTime;
          timeRemaining = formatTime(remainingTime);
        }
      }
    }, 1000) as unknown as number;
    
    try {
      const xhr = new XMLHttpRequest();
      activeXHR = xhr;
      
      xhr.upload.addEventListener('progress', (event) => {
        if (event.lengthComputable) {
          // Calculate overall progress
          overallProgress = Math.round((event.loaded / event.total) * 100);
          
          // Update time estimates
          updateTimeEstimates(event.loaded, event.total);
          
          // For multiple files, we don't have individual progress
          // So we set all files to the same progress
          Object.keys(uploadProgress).forEach(filename => {
            uploadProgress[filename] = overallProgress;
          });
          uploadProgress = {...uploadProgress}; // Trigger reactivity
        }
      });
      
      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          const result = JSON.parse(xhr.responseText);
          success = `${result.count} files uploaded successfully`;
          fileInput.value = '';
          selectedFiles = [];
          uploadProgress = {};
          
          if (updateIntervalId) {
            clearInterval(updateIntervalId);
          }
          
          dispatch('upload-success', result);
        } else {
          try {
            const errorData = JSON.parse(xhr.responseText);
            error = errorData.message || 'Upload failed';
          } catch (e) {
            error = 'Upload failed';
          }
        }
        uploading = false;
        activeXHR = null;
      });
      
      xhr.addEventListener('error', () => {
        error = 'Error uploading files';
        console.error('Upload error');
        uploading = false;
        activeXHR = null;
        
        if (updateIntervalId) {
          clearInterval(updateIntervalId);
        }
      });
      
      xhr.addEventListener('abort', () => {
        error = 'Upload aborted';
        uploading = false;
        activeXHR = null;
        
        if (updateIntervalId) {
          clearInterval(updateIntervalId);
        }
      });
      
      xhr.open('POST', '/api/upload');
      xhr.send(formData);
      
    } catch (err) {
      error = 'Error uploading files';
      console.error('Upload error:', err);
      uploading = false;
      activeXHR = null;
      
      if (updateIntervalId) {
        clearInterval(updateIntervalId);
      }
    }
  }
</script>

<form on:submit|preventDefault={handleSubmit}>
  <div class="form-group">
    <label for="file">Select media files:</label>
    <input 
      type="file" 
      id="file" 
      bind:this={fileInput} 
      accept="image/*,video/*,audio/*"
      disabled={uploading}
      multiple
      on:change={handleFileSelect}
    />
  </div>
  
  {#if selectedFiles.length > 0}
    <div class="selected-files">
      <p>Selected {selectedFiles.length} file{selectedFiles.length !== 1 ? 's' : ''}:</p>
      <ul>
        {#each selectedFiles as file}
          <li>{file.name} ({(file.size / 1024).toFixed(1)} KB)</li>
        {/each}
      </ul>
    </div>
  {/if}
  
  <button type="submit" disabled={uploading || selectedFiles.length === 0}>
    {#if uploading}
      Uploading... ({overallProgress}%)
    {:else}
      Upload
    {/if}
  </button>
  
  {#if uploading}
    <div class="upload-progress">
      <div class="progress-info">
        <span>Overall progress: {overallProgress}%</span>
        {#if uploadSpeed}
          <span>Speed: {uploadSpeed}</span>
        {/if}
        {#if timeRemaining}
          <span>Time remaining: {timeRemaining}</span>
        {/if}
      </div>
      
      <div class="progress-container">
        <div class="progress-bar" style="width: {overallProgress}%"></div>
      </div>
      
      {#if selectedFiles.length > 1}
        <div class="file-progress-list">
          <h4>File Progress:</h4>
          {#each selectedFiles as file}
            <div class="file-progress">
              <div class="file-progress-name" title={file.name}>{file.name}</div>
              <div class="file-progress-bar">
                <div class="file-progress-fill" style="width: {uploadProgress[file.name] || 0}%"></div>
              </div>
              <div class="file-progress-percent">{uploadProgress[file.name] || 0}%</div>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
  
  {#if error}
    <div class="error">{error}</div>
  {/if}
  
  {#if success}
    <div class="success">{success}</div>
  {/if}
</form>

<style>
  form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  
  label {
    font-weight: bold;
  }
  
  button {
    padding: 0.5rem 1rem;
    background-color: #4caf50;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-weight: bold;
  }
  
  button:hover {
    background-color: #45a049;
  }
  
  button:disabled {
    background-color: #cccccc;
    cursor: not-allowed;
  }
  
  .error {
    color: #f44336;
    padding: 0.5rem;
    background-color: #ffebee;
    border-radius: 4px;
  }
  
  .success {
    color: #4caf50;
    padding: 0.5rem;
    background-color: #e8f5e9;
    border-radius: 4px;
  }
  
  .selected-files {
    margin-top: 0.5rem;
    padding: 0.5rem;
    background-color: #f5f5f5;
    border-radius: 4px;
  }
  
  .selected-files p {
    margin: 0 0 0.5rem 0;
    font-weight: bold;
  }
  
  .selected-files ul {
    margin: 0;
    padding-left: 1.5rem;
  }
  
  .selected-files li {
    margin-bottom: 0.25rem;
  }
  
  .progress-container {
    width: 100%;
    height: 20px;
    background-color: #e0e0e0;
    border-radius: 4px;
    overflow: hidden;
    margin-top: 0.5rem;
  }
  
  .progress-bar {
    height: 100%;
    background-color: #4caf50;
    transition: width 0.3s ease;
  }
  
  .file-progress {
    display: flex;
    align-items: center;
    margin-bottom: 0.25rem;
  }
  
  .file-progress-name {
    flex: 1;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    margin-right: 0.5rem;
  }
  
  .file-progress-bar {
    flex: 2;
    height: 10px;
    background-color: #e0e0e0;
    border-radius: 2px;
    overflow: hidden;
    position: relative;
  }
  
  .file-progress-fill {
    height: 100%;
    background-color: #4caf50;
    transition: width 0.3s ease;
  }
  
  .file-progress-percent {
    min-width: 40px;
    text-align: right;
    font-size: 0.8rem;
  }
  
  .upload-progress {
    margin-top: 1rem;
    padding: 1rem;
    background-color: #f9f9f9;
    border-radius: 4px;
    border: 1px solid #e0e0e0;
  }
  
  .progress-info {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0.5rem;
    font-size: 0.9rem;
    color: #555;
  }
  
  .file-progress-list {
    margin-top: 1rem;
    max-height: 200px;
    overflow-y: auto;
    padding-right: 0.5rem;
  }
  
  .file-progress-list h4 {
    margin-top: 0;
    margin-bottom: 0.5rem;
    font-size: 0.9rem;
    color: #555;
  }
</style>