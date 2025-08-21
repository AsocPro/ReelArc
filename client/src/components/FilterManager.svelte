<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { FiltersConfig, Filter } from '../lib/types';
  import { reorderFilters } from '../lib/api';

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

          <div class="filter-status">
            <span class="status-badge" class:enabled={filter.enabled}>
              {filter.enabled ? 'Enabled' : 'Disabled'}
            </span>
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

  .filter-status {
    display: flex;
    align-items: center;
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

    .filter-status {
      justify-content: flex-end;
    }
  }
</style>