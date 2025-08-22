<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { FiltersConfig, Filter } from '../lib/types';
  import { reorderFilters, deleteFilter, pinFilter, unpinFilter } from '../lib/api';

  export let filtersConfig: FiltersConfig | null = null;

  const dispatch = createEventDispatcher();

  let draggedItem: Filter | null = null;
  let draggedOver: string | null = null;
  let localFilters: Filter[] = [];
  let isDragging = false;

  $: if (filtersConfig) {
    localFilters = [...filtersConfig.filters];
  }

  function handleDragStart(event: DragEvent, filter: Filter) {
    if (!event.dataTransfer) return;
    
    draggedItem = filter;
    isDragging = true;
    event.dataTransfer.effectAllowed = 'move';
    event.dataTransfer.setData('text/plain', filter.id);
    
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

  function handleDragOver(event: DragEvent, filterId: string) {
    event.preventDefault();
    if (!draggedItem || draggedItem.id === filterId) return;
    
    draggedOver = filterId;
    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = 'move';
    }
  }

  function handleDragLeave() {
    draggedOver = null;
  }

  async function handleDrop(event: DragEvent, targetFilter: Filter) {
    event.preventDefault();
    if (!draggedItem || draggedItem.id === targetFilter.id) return;

    const draggedIndex = localFilters.findIndex(f => f.id === draggedItem!.id);
    const targetIndex = localFilters.findIndex(f => f.id === targetFilter.id);

    if (draggedIndex === -1 || targetIndex === -1) return;

    // Create new array with reordered items
    const newFilters = [...localFilters];
    const [movedFilter] = newFilters.splice(draggedIndex, 1);
    if (movedFilter) {
      newFilters.splice(targetIndex, 0, movedFilter);
    }

    // Update local state immediately for responsive UI
    localFilters = newFilters;
    
    // Send reorder request to API
    const filterIds = newFilters.map(f => f.id);
    
    try {
      await reorderFilters(filterIds);
      // Emit event to parent to refresh filter config
      dispatch('filters-reordered');
    } catch (error) {
      console.error('Failed to save filter order:', error);
      // Revert local state on error
      if (filtersConfig) {
        localFilters = [...filtersConfig.filters];
      }
    }

    draggedOver = null;
  }

  function handleKeyDown(event: KeyboardEvent, _filter: Filter, index: number) {
    if (event.key === 'ArrowUp' && index > 0) {
      event.preventDefault();
      moveFilter(index, index - 1);
    } else if (event.key === 'ArrowDown' && index < localFilters.length - 1) {
      event.preventDefault();
      moveFilter(index, index + 1);
    }
  }

  async function moveFilter(fromIndex: number, toIndex: number) {
    const newFilters = [...localFilters];
    const [movedFilter] = newFilters.splice(fromIndex, 1);
    if (movedFilter) {
      newFilters.splice(toIndex, 0, movedFilter);
    }

    localFilters = newFilters;
    
    const filterIds = newFilters.map(f => f.id);
    
    try {
      await reorderFilters(filterIds);
      dispatch('filters-reordered');
    } catch (error) {
      console.error('Failed to save filter order:', error);
      if (filtersConfig) {
        localFilters = [...filtersConfig.filters];
      }
    }
  }

  async function handleDeleteFilter(filter: Filter) {
    const confirmation = confirm(`Are you sure you want to delete the filter "${filter.name}"? This action cannot be undone.`);
    if (!confirmation) return;

    try {
      await deleteFilter(filter.id);
      dispatch('filters-reordered');
    } catch (error) {
      console.error('Failed to delete filter:', error);
      alert('Failed to delete filter. Please try again.');
    }
  }

  async function handlePinFilter(filter: Filter) {
    try {
      await pinFilter(filter.id);
      dispatch('filters-reordered');
    } catch (error) {
      console.error('Failed to pin filter:', error);
      alert('Failed to pin filter. Please try again.');
    }
  }

  async function handleUnpinFilter(filter: Filter) {
    try {
      await unpinFilter(filter.id);
      dispatch('filters-reordered');
    } catch (error) {
      console.error('Failed to unpin filter:', error);
      alert('Failed to unpin filter. Please try again.');
    }
  }

  function isFilterPinned(filterId: string): boolean {
    return filtersConfig?.pinnedFilters?.includes(filterId) || false;
  }
</script>

<div class="filter-manager">
  <div class="manager-header">
    <h3>Quick Filter Management</h3>
    <p class="manager-description">
      Drag and drop to reorder filters, or use keyboard arrows when focused. Changes are saved automatically.
    </p>
  </div>

  {#if localFilters.length > 0}
    <div class="filters-list" class:is-dragging={isDragging}>
      {#each localFilters as filter, index (filter.id)}
        <div
          class="filter-item"
          class:drag-over={draggedOver === filter.id}
          class:disabled={!filter.enabled}
          draggable="true"
          tabindex="0"
          role="button"
          aria-label="Reorder filter: {filter.name}. Current position: {index + 1} of {localFilters.length}"
          on:dragstart={(e) => handleDragStart(e, filter)}
          on:dragend={handleDragEnd}
          on:dragover={(e) => handleDragOver(e, filter.id)}
          on:dragleave={handleDragLeave}
          on:drop={(e) => handleDrop(e, filter)}
          on:keydown={(e) => handleKeyDown(e, filter, index)}
        >
          <div class="drag-handle" aria-hidden="true">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M10 13a1 1 0 1 1-2 0 1 1 0 0 1 2 0zM10 6a1 1 0 1 1-2 0 1 1 0 0 1 2 0zM10 9.5a1 1 0 1 1-2 0 1 1 0 0 1 2 0zM8 13a1 1 0 1 1-2 0 1 1 0 0 1 2 0zM8 6a1 1 0 1 1-2 0 1 1 0 0 1 2 0zM8 9.5a1 1 0 1 1-2 0 1 1 0 0 1 2 0z"/>
            </svg>
          </div>
          
          <div class="filter-content">
            <div class="filter-info">
              <span class="filter-icon">{filter.icon}</span>
              <span class="filter-name">{filter.name}</span>
              <span class="filter-position">#{index + 1}</span>
            </div>
            <div class="filter-description">{filter.description}</div>
          </div>

          <div class="filter-actions">
            <div class="filter-status">
              <span class="status-badge" class:enabled={filter.enabled}>
                {filter.enabled ? 'Enabled' : 'Disabled'}
              </span>
              {#if isFilterPinned(filter.id)}
                <span class="pin-badge" title="Pinned to collapsed filter bar">📌</span>
              {/if}
            </div>
            <div class="action-buttons">
              {#if isFilterPinned(filter.id)}
                <button 
                  class="unpin-button"
                  title="Unpin from collapsed filter bar"
                  on:click|stopPropagation={() => handleUnpinFilter(filter)}
                >
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M9.5 12.5a3.5 3.5 0 1 1-7 0 3.5 3.5 0 0 1 7 0zm-.646-4.854.646.647.646-.647a.5.5 0 0 1 .708.708L10.207 9l.647.646a.5.5 0 0 1-.708.708L9.5 9.707l-.646.647a.5.5 0 0 1-.708-.708L8.793 9l-.647-.646a.5.5 0 0 1 .708-.708z"/>
                    <path d="M16 8A8 8 0 1 1 0 8a8 8 0 0 1 16 0zM5.354 4.646a.5.5 0 1 0-.708.708L7.293 8l-2.647 2.646a.5.5 0 0 0 .708.708L8 8.707l2.646 2.647a.5.5 0 0 0 .708-.708L8.707 8l2.647-2.646a.5.5 0 0 0-.708-.708L8 7.293 5.354 4.646z"/>
                  </svg>
                </button>
              {:else}
                <button 
                  class="pin-button"
                  title="Pin to collapsed filter bar"
                  on:click|stopPropagation={() => handlePinFilter(filter)}
                >
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M9.828.722a.5.5 0 0 1 .354.146l4.95 4.95a.5.5 0 0 1 0 .707c-.48.48-1.072.588-1.503.588-.177 0-.335-.018-.46-.039l-3.134 3.134a5.927 5.927 0 0 1 .16 1.013c.046.702-.032 1.687-.72 2.375a.5.5 0 0 1-.707 0l-2.829-2.828-3.182 3.182c-.195.195-.511.195-.707 0L1 11.414a.5.5 0 0 1 0-.707l3.182-3.182L1.353 4.696a.5.5 0 0 1 0-.707c.688-.688 1.673-.767 2.375-.72a5.922 5.922 0 0 1 1.013.16l3.134-3.133c-.021-.126-.039-.284-.039-.461 0-.431.108-1.023.589-1.503a.5.5 0 0 1 .353-.146zm.122 2.112L6.878 6.878a.5.5 0 0 1-.233.131l-1.337.267a1.5 1.5 0 0 0-.728.214l-.009.009-1.6 1.6a.5.5 0 0 1-.707 0L2.5 9.793l-.646-.647a.5.5 0 0 1 0-.707l1.293-1.293a.5.5 0 0 1 .707 0l.647.647L9.146 2.146a.5.5 0 0 1 .708 0l.353.353a.5.5 0 0 1 0 .708L6.561 6.853a.5.5 0 0 1-.708-.708L9.5 2.498z"/>
                  </svg>
                </button>
              {/if}
              <button 
                class="delete-button"
                title="Delete filter"
                on:click|stopPropagation={() => handleDeleteFilter(filter)}
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
      <p>No filters available to manage.</p>
    </div>
  {/if}
</div>

<style>
  .filter-manager {
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
    margin: 0;
    font-size: 0.9rem;
    color: #6c757d;
    line-height: 1.4;
  }

  .filters-list {
    padding: 1rem;
  }

  .filters-list.is-dragging {
    background: #f8f9fa;
  }

  .filter-item {
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

  .filter-item:last-child {
    margin-bottom: 0;
  }

  .filter-item:hover {
    border-color: #2196f3;
    box-shadow: 0 2px 8px rgba(33, 150, 243, 0.1);
    transform: translateY(-1px);
  }

  .filter-item:focus {
    outline: none;
    border-color: #2196f3;
    box-shadow: 0 0 0 3px rgba(33, 150, 243, 0.1);
  }

  .filter-item.drag-over {
    border-color: #2196f3;
    background: #f3f9ff;
    transform: scale(1.02);
  }

  .filter-item.disabled {
    opacity: 0.6;
    background: #f8f9fa;
  }

  .filter-item:global(.dragging) {
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

  .filter-item:hover .drag-handle {
    color: #2196f3;
  }

  .filter-content {
    flex: 1;
    min-width: 0;
  }

  .filter-info {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.25rem;
  }

  .filter-icon {
    font-size: 1.25rem;
    display: flex;
    align-items: center;
  }

  .filter-name {
    font-weight: 600;
    color: #2c3e50;
    flex: 1;
  }

  .filter-position {
    font-size: 0.8rem;
    color: #6c757d;
    background: #e9ecef;
    padding: 0.25rem 0.5rem;
    border-radius: 12px;
    min-width: 32px;
    text-align: center;
  }

  .filter-description {
    font-size: 0.9rem;
    color: #6c757d;
    line-height: 1.3;
  }

  .filter-actions {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.5rem;
  }

  .filter-status {
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

  .pin-badge {
    font-size: 0.9rem;
    padding: 0.2rem;
    background: #fff3cd;
    border: 1px solid #ffeaa7;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .pin-button, .unpin-button, .delete-button {
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
  }

  .pin-button {
    border-color: #28a745;
    color: #28a745;
  }

  .pin-button:hover {
    background: #28a745;
    color: white;
    transform: scale(1.05);
  }

  .unpin-button {
    border-color: #ffc107;
    color: #ffc107;
  }

  .unpin-button:hover {
    background: #ffc107;
    color: #212529;
    transform: scale(1.05);
  }

  .delete-button {
    border-color: #dc3545;
    color: #dc3545;
  }

  .delete-button:hover {
    background: #dc3545;
    color: white;
    transform: scale(1.05);
  }

  .pin-button:focus, .unpin-button:focus, .delete-button:focus {
    outline: none;
  }

  .pin-button:focus {
    box-shadow: 0 0 0 2px rgba(40, 167, 69, 0.2);
  }

  .unpin-button:focus {
    box-shadow: 0 0 0 2px rgba(255, 193, 7, 0.2);
  }

  .delete-button:focus {
    box-shadow: 0 0 0 2px rgba(220, 53, 69, 0.2);
  }

  .pin-button:active, .unpin-button:active, .delete-button:active {
    transform: scale(0.95);
  }

  .empty-state {
    padding: 2rem;
    text-align: center;
    color: #6c757d;
  }

  .empty-state p {
    margin: 0;
    font-size: 1rem;
  }

  /* Responsive design */
  @media (max-width: 768px) {
    .filter-item {
      flex-direction: column;
      align-items: stretch;
      gap: 0.75rem;
    }

    .filter-info {
      justify-content: space-between;
    }

    .drag-handle {
      align-self: flex-start;
    }

    .filter-actions {
      flex-direction: row;
      justify-content: space-between;
      align-items: center;
    }

    .filter-status {
      justify-content: flex-start;
    }

    .pin-button, .unpin-button, .delete-button {
      min-width: 28px;
      height: 28px;
      padding: 0.375rem;
    }
  }
</style>