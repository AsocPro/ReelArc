<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  
  const dispatch = createEventDispatcher();
  
  let fileInput: HTMLInputElement;
  let uploading = false;
  let error = '';
  let success = '';
  
  async function handleSubmit() {
    if (!fileInput.files || fileInput.files.length === 0) {
      error = 'Please select a file to upload';
      return;
    }
    
    const file = fileInput.files[0];
    const formData = new FormData();
    formData.append('file', file);
    
    uploading = true;
    error = '';
    success = '';
    
    try {
      const response = await fetch('/api/upload', {
        method: 'POST',
        body: formData
      });
      
      if (response.ok) {
        const result = await response.json();
        success = `File uploaded successfully: ${file.name}`;
        fileInput.value = '';
        dispatch('upload-success', result);
      } else {
        const errorData = await response.json();
        error = errorData.message || 'Upload failed';
      }
    } catch (err) {
      error = 'Error uploading file';
      console.error('Upload error:', err);
    } finally {
      uploading = false;
    }
  }
</script>

<form on:submit|preventDefault={handleSubmit}>
  <div class="form-group">
    <label for="file">Select media file:</label>
    <input 
      type="file" 
      id="file" 
      bind:this={fileInput} 
      accept="image/*,video/*,audio/*"
      disabled={uploading}
    />
  </div>
  
  <button type="submit" disabled={uploading}>
    {#if uploading}
      Uploading...
    {:else}
      Upload
    {/if}
  </button>
  
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
</style>