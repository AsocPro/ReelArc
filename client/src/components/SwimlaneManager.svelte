<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { SwimlanesConfig, Swimlane } from '../lib/types';
  import { reorderSwimlanes, deleteSwimlane, saveSwimlane } from '../lib/api';

  export let swimlanesConfig: SwimlanesConfig | null = null;

  const dispatch = createEventDispatcher();

  let draggedItem: Swimlane | null = null;
  let draggedOver: string | null = null;
  let localSwimlanes: Swimlane[] = [];
  let isDragging = false;
  let showAddDialog = false;
  let newSwimlaneName = '';
  let newSwimlaneDescription = '';
  let newSwimlaneIcon = '🏷️';
  let newSwimlaneColor = '#f5f5f5';

  $: if (swimlanesConfig) {
    localSwimlanes = [...swimlanesConfig.swimlanes].sort((a, b) => a.order - b.order);
  }

  function handleDragStart(event: DragEvent, swimlane: Swimlane) {
    if (!event.dataTransfer) return;

    draggedItem = swimlane;
    isDragging = true;
    event.dataTransfer.effectAllowed = 'move';
    event.dataTransfer.setData('text/plain', swimlane.id);

    // Add visual feedback to the dragged element
    if (event.target instanceof Element) {
      event.target.classList.add('dragging');
    }
  }

  function handleDragEnd(event: DragEvent) {
    draggedItem = null;
    draggedOver = null;
    isDragging = false;

    // Remove visual feedback
    if (event.target instanceof Element) {
      event.target.classList.remove('dragging');
    }
  }

  function handleDragOver(event: DragEvent, swimlaneId: string) {
    event.preventDefault();
    if (!draggedItem || draggedItem.id === swimlaneId) return;

    draggedOver = swimlaneId;
    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = 'move';
    }
  }

  function handleDragLeave() {
    draggedOver = null;
  }

  async function handleDrop(event: DragEvent, targetSwimlane: Swimlane) {
    event.preventDefault();
    if (!draggedItem || draggedItem.id === targetSwimlane.id) return;

    const draggedIndex = localSwimlanes.findIndex(s => s.id === draggedItem!.id);
    const targetIndex = localSwimlanes.findIndex(s => s.id === targetSwimlane.id);

    if (draggedIndex === -1 || targetIndex === -1) return;

    // Create new array with reordered items
    const newSwimlanes = [...localSwimlanes];
    const [movedSwimlane] = newSwimlanes.splice(draggedIndex, 1);
    if (movedSwimlane) {
      newSwimlanes.splice(targetIndex, 0, movedSwimlane);
    }

    // Update local state immediately for responsive UI
    localSwimlanes = newSwimlanes;

    // Update order values
    newSwimlanes.forEach((swimlane, index) => {
      swimlane.order = index;
    });

    // Send reorder request to API
    const swimlaneIds = newSwimlanes.map(s => s.id);

    try {
      await reorderSwimlanes(swimlaneIds);
      // Emit event to parent to refresh swimlanes config
      dispatch('swimlanes-reordered');
    } catch (error) {
      console.error('Failed to save swimlane order:', error);
      // Revert local state on error
      if (swimlanesConfig) {
        localSwimlanes = [...swimlanesConfig.swimlanes].sort((a, b) => a.order - b.order);
      }
    }

    draggedOver = null;
  }

  function handleKeyDown(event: KeyboardEvent, _swimlane: Swimlane, index: number) {
    if (event.key === 'ArrowUp' && index > 0) {
      event.preventDefault();
      moveSwimlane(index, index - 1);
    } else if (event.key === 'ArrowDown' && index < localSwimlanes.length - 1) {
      event.preventDefault();
      moveSwimlane(index, index + 1);
    }
  }

  async function moveSwimlane(fromIndex: number, toIndex: number) {
    const newSwimlanes = [...localSwimlanes];
    const [movedSwimlane] = newSwimlanes.splice(fromIndex, 1);
    if (movedSwimlane) {
      newSwimlanes.splice(toIndex, 0, movedSwimlane);
    }

    // Update order values
    newSwimlanes.forEach((swimlane, index) => {
      swimlane.order = index;
    });

    localSwimlanes = newSwimlanes;

    const swimlaneIds = newSwimlanes.map(s => s.id);

    try {
      await reorderSwimlanes(swimlaneIds);
      dispatch('swimlanes-reordered');
    } catch (error) {
      console.error('Failed to save swimlane order:', error);
      if (swimlanesConfig) {
        localSwimlanes = [...swimlanesConfig.swimlanes].sort((a, b) => a.order - b.order);
      }
    }
  }

  async function handleDeleteSwimlane(swimlane: Swimlane) {
    const confirmation = confirm(`Are you sure you want to delete the swimlane "${swimlane.name}"? This action cannot be undone.`);
    if (!confirmation) return;

    try {
      await deleteSwimlane(swimlane.id);
      dispatch('swimlanes-reordered');
    } catch (error) {
      console.error('Failed to delete swimlane:', error);
      alert('Failed to delete swimlane. Please try again.');
    }
  }

  function openAddDialog() {
    showAddDialog = true;
    newSwimlaneName = '';
    newSwimlaneDescription = '';
    newSwimlaneIcon = '🏷️';
    newSwimlaneColor = '#f5f5f5';
  }

  function closeAddDialog() {
    showAddDialog = false;
    newSwimlaneName = '';
    newSwimlaneDescription = '';
    newSwimlaneIcon = '🏷️';
    newSwimlaneColor = '#f5f5f5';
  }

  async function saveNewSwimlane() {
    if (!newSwimlaneName.trim()) {
      alert('Please enter a swimlane name.');
      return;
    }

    try {
      // Generate unique ID based on name and timestamp
      const swimlaneId = newSwimlaneName.toLowerCase().replace(/[^a-z0-9]+/g, '_') + '_' + Date.now();

      const swimlane: Swimlane = {
        id: swimlaneId,
        name: newSwimlaneName,
        description: newSwimlaneDescription,
        icon: newSwimlaneIcon,
        enabled: true,
        color: newSwimlaneColor,
        order: localSwimlanes.length
      };

      await saveSwimlane(swimlane);

      // Reload swimlanes configuration to show the new swimlane
      dispatch('swimlanes-reordered');

      // Close the dialog
      closeAddDialog();

      // Show success message
      alert(`Swimlane "${newSwimlaneName}" created successfully!`);

    } catch (error) {
      console.error('Failed to save swimlane:', error);
      alert('Failed to save swimlane. Please try again.');
    }
  }
</script>

<div class="swimlane-manager">
  <div class="manager-header">
    <h3>Swimlane Management</h3>
    <p class="manager-description">
      Drag and drop to reorder swimlanes, or use keyboard arrows when focused. Changes are saved automatically.
    </p>
    <button class="add-swimlane-btn" on:click={openAddDialog}>
      <span class="add-icon">+</span>
      Add Swimlane
    </button>
  </div>

  {#if localSwimlanes.length > 0}
    <div class="swimlanes-list" class:is-dragging={isDragging}>
      {#each localSwimlanes as swimlane, index (swimlane.id)}
        <div
          class="swimlane-item"
          class:drag-over={draggedOver === swimlane.id}
          draggable="true"
          tabindex="0"
          role="button"
          aria-label="Reorder swimlane: {swimlane.name}. Current position: {index + 1} of {localSwimlanes.length}"
          on:dragstart={(e) => handleDragStart(e, swimlane)}
          on:dragend={handleDragEnd}
          on:dragover={(e) => handleDragOver(e, swimlane.id)}
          on:dragleave={handleDragLeave}
          on:drop={(e) => handleDrop(e, swimlane)}
          on:keydown={(e) => handleKeyDown(e, swimlane, index)}
        >
          <div class="drag-handle" aria-hidden="true">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M10 13a1 1 0 1 1-2 0 1 1 0 0 1 2 0zM10 6a1 1 0 1 1-2 0 1 1 0 0 1 2 0zM10 9.5a1 1 0 1 1-2 0 1 1 0 0 1 2 0zM8 13a1 1 0 1 1-2 0 1 1 0 0 1 2 0zM8 6a1 1 0 1 1-2 0 1 1 0 0 1 2 0zM8 9.5a1 1 0 1 1-2 0 1 1 0 0 1 2 0z"/>
            </svg>
          </div>

          <div class="swimlane-content">
            <div class="swimlane-info">
              <span class="swimlane-icon">{swimlane.icon}</span>
              <span class="swimlane-name">{swimlane.name}</span>
              <span class="swimlane-position">#{index + 1}</span>
            </div>
            <div class="swimlane-description">{swimlane.description}</div>
            {#if swimlane.color}
              <div class="swimlane-color" style="background-color: {swimlane.color}"></div>
            {/if}
          </div>

          <div class="swimlane-actions">
            <div class="swimlane-status">
              <span class="status-badge" class:enabled={swimlane.enabled}>
                {swimlane.enabled ? 'Enabled' : 'Disabled'}
              </span>
            </div>
            <div class="action-buttons">
              <button
                class="delete-button"
                title="Delete swimlane"
                on:click|stopPropagation={() => handleDeleteSwimlane(swimlane)}
              >
                <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                  <path d="M11 1.5v1h3.5a.5.5 0 0 1 0 1h-.538l-.853 10.66A2 2 0 0 1 11.115 16h-6.23a2 2 0 0 1-1.994-1.84L2.038 3.5H1.5a.5.5 0 0 1 0-1H5v-1A1.5 1.5 0 0 1 6.5 0h3A1.5 1.5 0 0 1 11 1.5Zm-5 0v1h4v-1a.5.5 0 0 0-.5-.5h-3a.5.5 0 0 0-.5.5ZM4.5 5.029l.5 8.5a.5.5 0 1 0 .998-.06l-.5-8.5a.5.5 0 1 0-.998.06Zm6.53-.528a.5.5 0 0 0-.528.47l-.5 8.5a.5.5 0 0 0 .998.058l.5-8.5a.5.5 0 0 0-.47-.528ZM8 4.5a.5.5 0 0 0-.5.5v8.5a.5.5 0 0 0 1 0V5a.5.5 0 0 0-.5-.5Z"/>
                </svg>
              </button>
            </div>
          </div>
        </div>
      {/each}
    </div>
  {:else}
    <div class="empty-state">
      <p>No swimlanes available to manage.</p>
      <button class="add-first-swimlane-btn" on:click={openAddDialog}>
        Create your first swimlane
      </button>
    </div>
  {/if}
</div>

<!-- Add Swimlane Dialog -->
{#if showAddDialog}
  <div class="modal-overlay" on:click={closeAddDialog}>
    <div class="add-dialog" on:click|stopPropagation>
      <div class="dialog-header">
        <h3>Add New Swimlane</h3>
        <button class="close-btn" on:click={closeAddDialog}>×</button>
      </div>

      <div class="dialog-content">
        <div class="form-group">
          <label for="swimlane-name">Swimlane Name:</label>
          <input
            id="swimlane-name"
            type="text"
            bind:value={newSwimlaneName}
            placeholder="Enter swimlane name"
            class="dialog-input"
          />
        </div>

        <div class="form-group">
          <label for="swimlane-description">Description:</label>
          <textarea
            id="swimlane-description"
            bind:value={newSwimlaneDescription}
            placeholder="Enter swimlane description"
            class="dialog-textarea"
            rows="3"
          ></textarea>
        </div>

        <div class="form-group">
          <label for="swimlane-icon">Icon:</label>
          <input
            id="swimlane-icon"
            type="text"
            bind:value={newSwimlaneIcon}
            placeholder="🏷️"
            class="dialog-input icon-input"
            maxlength="2"
          />
          <span class="icon-hint">Choose an emoji icon for your swimlane</span>
        </div>

        <div class="form-group">
          <label for="swimlane-color">Color:</label>
          <input
            id="swimlane-color"
            type="color"
            bind:value={newSwimlaneColor}
            class="dialog-input color-input"
          />
        </div>
      </div>

      <div class="dialog-actions">
        <button class="dialog-btn cancel-btn" on:click={closeAddDialog}>Cancel</button>
        <button class="dialog-btn save-btn" on:click={saveNewSwimlane}>Create Swimlane</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .swimlane-manager {
    background: white;
    border: 1px solid #e1e5e9;
    border-radius: 8px;
    overflow: hidden;
  }

  .manager-header {
    padding: 1.5rem;
    background: #f8f9fa;
    border-bottom: 1px solid #e1e5e9;
  }

  .manager-header h3 {
    margin: 0 0 0.5rem 0;
    font-size: 1.1rem;
    font-weight: 600;
    color: #2c3e50;
  }

  .manager-description {
    margin: 0 0 1rem 0;
    font-size: 0.9rem;
    color: #6c757d;
    line-height: 1.4;
  }

  .add-swimlane-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    background: #4caf50;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.9rem;
    transition: all 0.2s ease;
  }

  .add-swimlane-btn:hover {
    background: #45a049;
    transform: translateY(-1px);
  }

  .add-icon {
    font-size: 1.2rem;
    font-weight: bold;
  }

  .swimlanes-list {
    padding: 1rem;
  }

  .swimlanes-list.is-dragging {
    background: #f8f9fa;
  }

  .swimlane-item {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem;
    margin-bottom: 0.5rem;
    background: white;
    border: 2px solid #e9ecef;
    border-radius: 8px;
    cursor: grab;
    transition: all 0.2s ease;
    user-select: none;
  }

  .swimlane-item:last-child {
    margin-bottom: 0;
  }

  .swimlane-item:hover {
    border-color: #2196f3;
    box-shadow: 0 2px 8px rgba(33, 150, 243, 0.1);
    transform: translateY(-1px);
  }

  .swimlane-item:focus {
    outline: none;
    border-color: #2196f3;
    box-shadow: 0 0 0 3px rgba(33, 150, 243, 0.1);
  }

  .swimlane-item.drag-over {
    border-color: #2196f3;
    background: #f3f9ff;
    transform: scale(1.02);
  }

  .swimlane-item:global(.dragging) {
    opacity: 0.5;
    transform: rotate(5deg);
    cursor: grabbing;
  }

  .drag-handle {
    display: flex;
    align-items: center;
    color: #6c757d;
    transition: color 0.2s ease;
  }

  .swimlane-item:hover .drag-handle {
    color: #2196f3;
  }

  .swimlane-content {
    flex: 1;
    min-width: 0;
  }

  .swimlane-info {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.25rem;
  }

  .swimlane-icon {
    font-size: 1.25rem;
    display: flex;
    align-items: center;
  }

  .swimlane-name {
    font-weight: 600;
    color: #2c3e50;
    flex: 1;
  }

  .swimlane-position {
    font-size: 0.8rem;
    color: #6c757d;
    background: #e9ecef;
    padding: 0.25rem 0.5rem;
    border-radius: 12px;
    min-width: 32px;
    text-align: center;
  }

  .swimlane-description {
    font-size: 0.9rem;
    color: #6c757d;
    line-height: 1.3;
  }

  .swimlane-color {
    width: 20px;
    height: 20px;
    border-radius: 4px;
    border: 1px solid #ddd;
    margin-top: 0.5rem;
  }

  .swimlane-actions {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.5rem;
  }

  .swimlane-status {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .action-buttons {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .status-badge {
    padding: 0.375rem 0.75rem;
    font-size: 0.8rem;
    font-weight: 500;
    border-radius: 16px;
    background: #dc3545;
    color: white;
  }

  .status-badge.enabled {
    background: #28a745;
  }

  .delete-button {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.5rem;
    background: transparent;
    border: 1px solid;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.2s ease;
    min-width: 32px;
    height: 32px;
    border-color: #dc3545;
    color: #dc3545;
  }

  .delete-button:hover {
    background: #dc3545;
    color: white;
    transform: scale(1.05);
  }

  .delete-button:focus {
    outline: none;
  }

  .empty-state {
    padding: 2rem;
    text-align: center;
    color: #6c757d;
  }

  .empty-state p {
    margin: 0 0 1rem 0;
    font-size: 1rem;
  }

  .add-first-swimlane-btn {
    padding: 0.75rem 1.5rem;
    background: #4caf50;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.9rem;
    transition: all 0.2s ease;
  }

  .add-first-swimlane-btn:hover {
    background: #45a049;
    transform: translateY(-1px);
  }

  /* Modal Styles */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    animation: fadeIn 0.2s ease;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .add-dialog {
    background: white;
    border-radius: 8px;
    width: 90%;
    max-width: 500px;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
    animation: slideUp 0.2s ease;
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .dialog-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1.5rem;
    border-bottom: 1px solid #e1e5e9;
  }

  .dialog-header h3 {
    margin: 0;
    font-size: 1.2rem;
    font-weight: 600;
    color: #2c3e50;
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
    color: #6c757d;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    transition: all 0.2s ease;
  }

  .close-btn:hover {
    background: #f8f9fa;
    color: #2c3e50;
  }

  .dialog-content {
    padding: 1.5rem;
  }

  .form-group {
    margin-bottom: 1.5rem;
  }

  .form-group:last-child {
    margin-bottom: 0;
  }

  .form-group label {
    display: block;
    font-weight: 500;
    color: #2c3e50;
    margin-bottom: 0.5rem;
    font-size: 0.9rem;
  }

  .dialog-input,
  .dialog-textarea {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 1rem;
    font-family: inherit;
    transition: border-color 0.2s ease;
  }

  .dialog-input:focus,
  .dialog-textarea:focus {
    outline: none;
    border-color: #2196f3;
    box-shadow: 0 0 0 3px rgba(33, 150, 243, 0.1);
  }

  .icon-input {
    width: 80px;
    text-align: center;
    font-size: 1.2rem;
  }

  .color-input {
    width: 60px;
    height: 40px;
    padding: 0;
    border: none;
    cursor: pointer;
  }

  .icon-hint {
    display: block;
    font-size: 0.8rem;
    color: #6c757d;
    margin-top: 0.25rem;
  }

  .dialog-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 0.75rem;
    padding: 1.5rem;
    border-top: 1px solid #e1e5e9;
    background: #f8f9fa;
  }

  .dialog-btn {
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.9rem;
    font-weight: 500;
    transition: all 0.2s ease;
    min-width: 100px;
  }

  .cancel-btn {
    background: #f5f5f5;
    color: #666;
    border: 1px solid #ddd;
  }

  .cancel-btn:hover {
    background: #eeeeee;
    color: #555;
  }

  .dialog-btn.save-btn {
    background: #4caf50;
    color: white;
  }

  .dialog-btn.save-btn:hover {
    background: #45a049;
  }

  /* Responsive design */
  @media (max-width: 768px) {
    .swimlane-item {
      flex-direction: column;
      align-items: stretch;
      gap: 0.75rem;
    }

    .swimlane-info {
      justify-content: space-between;
    }

    .drag-handle {
      align-self: flex-start;
    }

    .swimlane-actions {
      flex-direction: row;
      justify-content: space-between;
      align-items: center;
    }

    .swimlane-status {
      justify-content: flex-start;
    }

    .delete-button {
      min-width: 28px;
      height: 28px;
      padding: 0.375rem;
    }

    .add-dialog {
      width: 95%;
      margin: 1rem;
    }

    .dialog-header,
    .dialog-content,
    .dialog-actions {
      padding: 1rem;
    }

    .dialog-actions {
      flex-direction: column-reverse;
      gap: 0.5rem;
    }

    .dialog-btn {
      width: 100%;
    }
  }
</style>