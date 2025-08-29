<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { MediaItem, SwimlanesConfig } from '../lib/types';
  import { updateLabels } from '../lib/api';

  export let data: MediaItem[] = [];
  export let loading = false;
  export let error = '';
  export let swimlanesConfig: SwimlanesConfig | null = null;

  const dispatch = createEventDispatcher<{
    'item-select': MediaItem;
    'item-update': MediaItem;
  }>();

  function handleItemClick(item: MediaItem) {
    dispatch('item-select', item);
  }

  // Drag and drop state
  let draggedItem: MediaItem | null = null;
  let draggedOverColumn: string | null = null;

  function handleDragStart(event: DragEvent, item: MediaItem, sourceLabel: string) {
    draggedItem = item;
    event.dataTransfer!.effectAllowed = 'move';
    event.dataTransfer!.setData('text/plain', JSON.stringify({
      itemId: item.id,
      sourceLabel: sourceLabel
    }));

    // Add dragging class to the card
    const target = event.target as HTMLElement;
    target.classList.add('dragging');
  }

  function handleDragEnd(_event: DragEvent) {
    draggedItem = null;
    draggedOverColumn = null;

    // Remove dragging class from all cards
    const cards = document.querySelectorAll('.kanban-card');
    cards.forEach(card => card.classList.remove('dragging'));
  }

  function handleDragOver(event: DragEvent, targetLabel: string) {
    event.preventDefault();
    event.dataTransfer!.dropEffect = 'move';

    draggedOverColumn = targetLabel;

    // Add drag-over class to the column
    const target = event.currentTarget as HTMLElement;
    target.classList.add('drag-over');
  }

  function handleDragLeave(event: DragEvent) {
    // Remove drag-over class when leaving the column
    const target = event.currentTarget as HTMLElement;
    target.classList.remove('drag-over');
  }

  async function handleDrop(event: DragEvent, targetLabel: string) {
    event.preventDefault();

    if (!draggedItem) return;

    // Remove drag-over class
    const target = event.currentTarget as HTMLElement;
    target.classList.remove('drag-over');

    // Parse drag data
    const dragData = JSON.parse(event.dataTransfer!.getData('text/plain'));
    const sourceLabel = dragData.sourceLabel;

    // Only proceed if dropping in a different column
    if (sourceLabel !== targetLabel) {
      try {
        // Calculate new labels array
        let newLabels: string[] = [];

        if (targetLabel === 'Unlabeled') {
          // Remove all labels for Unlabeled lane
          newLabels = [];
        } else {
          // Remove source label and add target label
          newLabels = [...draggedItem.labels];

          // Remove source label if it exists
          if (sourceLabel !== 'Unlabeled') {
            const sourceIndex = newLabels.indexOf(sourceLabel);
            if (sourceIndex > -1) {
              newLabels.splice(sourceIndex, 1);
            }
          }

          // Add target label if not already present
          if (!newLabels.includes(targetLabel)) {
            newLabels.push(targetLabel);
          }
        }

        // Update the item's labels via API
        const updatedItem = await updateLabels(draggedItem.id, newLabels);
        if (updatedItem) {
          // Update local data state immediately
          data = data.map(item =>
            item.id === updatedItem.id ? updatedItem : item
          );

          dispatch('item-update', updatedItem);
        } else {
          // Handle API error
          error = 'Failed to update item labels. Please try again.';
          console.error('Failed to update labels: API returned null');
        }
      } catch (apiError) {
        // Handle API error gracefully
        error = 'Failed to update item labels. Please check your connection and try again.';
        console.error('Failed to update labels:', apiError);
      }
    }

    draggedItem = null;
    draggedOverColumn = null;
  }

  function formatTimestamp(timestamp: string): string {
    return new Date(timestamp).toLocaleString();
  }

  function formatDuration(duration?: number): string {
    if (!duration) return '-';
    const minutes = Math.floor(duration / 60);
    const seconds = Math.floor(duration % 60);
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
  }



  // Group items by swimlane labels for Kanban columns
  $: groupedItems = data.reduce((groups, item) => {
    // If swimlanes are configured, use the first matching swimlane
    if (swimlanesConfig && swimlanesConfig.swimlanes.length > 0) {
      // Find the first enabled swimlane that matches an item label
      let matchedSwimlane = null;
      for (const swimlane of swimlanesConfig.swimlanes) {
        if (swimlane.enabled && item.labels && item.labels.includes(swimlane.name)) {
          matchedSwimlane = swimlane;
          break;
        }
      }

      // If no swimlane matches, use the first enabled swimlane as default
      if (!matchedSwimlane) {
        matchedSwimlane = swimlanesConfig.swimlanes.find(s => s.enabled) || swimlanesConfig.swimlanes[0];
      }

      if (matchedSwimlane) {
        if (!groups[matchedSwimlane.name]) {
          groups[matchedSwimlane.name] = [];
        }
        groups[matchedSwimlane.name].push(item);
      }
    } else {
      // Fallback to original logic if no swimlanes configured
      const primaryLabel: string = (item.labels && item.labels.length > 0 && item.labels[0]) ? item.labels[0] : 'Unlabeled';

      if (!groups[primaryLabel]) {
        groups[primaryLabel] = [];
      }
      groups[primaryLabel].push(item);
    }

    return groups;
  }, {} as Record<string, MediaItem[]>);

  // Get swimlane order from configuration
  $: swimlaneOrder = swimlanesConfig && swimlanesConfig.swimlanes.length > 0
    ? swimlanesConfig.swimlanes
        .filter(s => s.enabled)
        .sort((a, b) => a.order - b.order)
        .map(s => s.name)
    : [];

  // Define column order based on swimlanes or fallback to original logic
  $: columnOrder = swimlaneOrder.length > 0 ? swimlaneOrder : (() => {
    const availableLabels = Array.from(
      new Set(
        data.flatMap(item => item.labels.length > 0 ? item.labels : ['Unlabeled'])
      )
    ).sort();
    return ['Unlabeled', ...availableLabels.filter(label => label !== 'Unlabeled')];
  })();

  function getColumnTitle(label: string): string {
    // If swimlanes are configured, find the swimlane and use its name
    if (swimlanesConfig) {
      const swimlane = swimlanesConfig.swimlanes.find(s => s.name === label);
      if (swimlane) {
        return swimlane.name;
      }
    }

    // Fallback to original logic
    if (label === 'Unlabeled') {
      return 'Unlabeled';
    }
    return label;
  }

  function getColumnColor(label: string): string {
    // If swimlanes are configured, use the swimlane's color
    if (swimlanesConfig) {
      const swimlane = swimlanesConfig.swimlanes.find(s => s.name === label);
      if (swimlane && swimlane.color) {
        return swimlane.color;
      }
    }

    // Fallback to original logic
    if (label === 'Unlabeled') {
      return '#f5f5f5'; // Light gray for unlabeled
    }

    // Generate consistent colors based on label hash
    const hash = label.split('').reduce((a: number, b: string) => {
      const charCode = b.charCodeAt(0);
      a = ((a << 5) - a) + (isNaN(charCode) ? 0 : charCode);
      return a & a;
    }, 0);

    const colors = [
      '#e8f5e8', // Light green
      '#e3f2fd', // Light blue
      '#fff3e0', // Light orange
      '#f3e5f5', // Light purple
      '#e8f4fd', // Light cyan
      '#f1f8e9', // Light lime
      '#fce4ec', // Light pink
      '#f3e5f5', // Light purple (duplicate for variety)
    ];

    const colorIndex = Math.abs(hash) % colors.length;
    return colors[colorIndex] || '#f5f5f5';
  }

  function getColumnIcon(label: string): string {
    // If swimlanes are configured, use the swimlane's icon
    if (swimlanesConfig) {
      const swimlane = swimlanesConfig.swimlanes.find(s => s.name === label);
      if (swimlane && swimlane.icon) {
        return swimlane.icon;
      }
    }

    // Default icon
    return '🏷️';
  }
</script>

<div class="kanban-container">
  {#if loading}
    <div class="loading">Loading kanban data...</div>
  {:else if error}
    <div class="error">{error}</div>
  {:else if data.length === 0}
    <div class="empty">No media items found</div>
  {:else}
    <div class="kanban-board">
      {#each columnOrder as label}
        {@const items = groupedItems[label] || []}
        <div class="kanban-column" style="background-color: {getColumnColor(label)}">
          <div class="column-header">
            <h3 class="column-title">
              <span class="type-icon">{getColumnIcon(label)}</span>
              {getColumnTitle(label)}
            </h3>
            <span class="item-count">{items.length}</span>
          </div>
           <div
            class="column-content"
            class:drag-over={draggedOverColumn === label}
            on:dragover={(e) => handleDragOver(e, label)}
            on:dragleave={handleDragLeave}
            on:drop={(e) => handleDrop(e, label)}
          >
             {#each items as item}
               <div
                 class="kanban-card"
                 class:has-transcription={item.transcription}
                 class:dragging={draggedItem?.id === item.id}
                 draggable="true"
                 on:click={() => handleItemClick(item)}
                  on:dragstart={(e) => handleDragStart(e, item, label)}
                 on:dragend={handleDragEnd}
               >
                <div class="card-header">
                  <div class="filename" title={item.filename}>
                    {item.filename}
                  </div>
                  <div class="timestamp">
                    {formatTimestamp(item.timestamp)}
                  </div>
                </div>

                {#if item.duration}
                  <div class="duration">
                    Duration: {formatDuration(item.duration)}
                  </div>
                {/if}

                {#if item.labels && item.labels.length > 0}
                  <div class="labels">
                    {#each item.labels.slice(0, 2) as label}
                      <span class="label-tag">{label}</span>
                    {/each}
                    {#if item.labels.length > 2}
                      <span class="label-more">+{item.labels.length - 2}</span>
                    {/if}
                  </div>
                {/if}

                {#if item.transcription}
                  <div class="transcription-preview" title={item.transcription}>
                    {item.transcription.substring(0, 80)}{item.transcription.length > 80 ? '...' : ''}
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .kanban-container {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .kanban-board {
    display: flex;
    gap: 16px;
    padding: 16px;
    height: 100%;
    overflow-x: auto;
    overflow-y: hidden;
  }

  .kanban-column {
    flex: 1;
    min-width: 280px;
    max-width: 320px;
    border-radius: 8px;
    border: 1px solid #e0e0e0;
    display: flex;
    flex-direction: column;
    background-color: #f9f9f9;
  }

  .column-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid #e0e0e0;
    background-color: rgba(255, 255, 255, 0.8);
    border-radius: 8px 8px 0 0;
  }

  .column-title {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: #333;
  }

  .type-icon {
    font-size: 1.1rem;
  }

  .item-count {
    background-color: #666;
    color: white;
    padding: 2px 8px;
    border-radius: 12px;
    font-size: 0.75rem;
    font-weight: 500;
  }

  .column-content {
    flex: 1;
    padding: 12px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .kanban-card {
    background-color: white;
    border: 1px solid #e0e0e0;
    border-radius: 6px;
    padding: 12px;
    cursor: pointer;
    transition: all 0.2s ease;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .kanban-card:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
    transform: translateY(-1px);
  }

  .kanban-card.dragging {
    opacity: 0.5;
    transform: rotate(5deg);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }

  .column-content.drag-over {
    background-color: rgba(33, 150, 243, 0.1);
    border: 2px dashed #2196f3;
    border-radius: 6px;
  }

  .kanban-card.has-transcription {
    border-left: 3px solid #4caf50;
  }

  .card-header {
    margin-bottom: 8px;
  }

  .filename {
    font-weight: 600;
    font-size: 0.9rem;
    color: #333;
    margin-bottom: 4px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .timestamp {
    font-size: 0.75rem;
    color: #666;
  }

  .duration {
    font-size: 0.8rem;
    color: #666;
    margin-bottom: 8px;
    font-family: monospace;
  }

  .labels {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    margin-bottom: 8px;
  }

  .label-tag {
    background-color: #e3f2fd;
    color: #1976d2;
    padding: 2px 6px;
    border-radius: 10px;
    font-size: 0.7rem;
    white-space: nowrap;
  }

  .label-more {
    color: #666;
    font-size: 0.7rem;
    font-style: italic;
  }

  .transcription-preview {
    font-size: 0.8rem;
    color: #333;
    line-height: 1.4;
    padding: 6px;
    background-color: #f8f9fa;
    border-radius: 4px;
    border-left: 2px solid #4caf50;
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

  /* Responsive adjustments */
  @media (max-width: 768px) {
    .kanban-board {
      padding: 8px;
      gap: 8px;
    }

    .kanban-column {
      min-width: 240px;
      max-width: 280px;
    }
  }
</style>