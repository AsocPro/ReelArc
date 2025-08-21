import type { MediaItem, TranscriptionStatus, MediaFilters, FiltersConfig } from './types';

/**
 * Fetches media items from the API
 * @param filters Optional filters for date range and labels
 * @returns Promise with array of media items
 */
export async function fetchMediaItems(filters?: MediaFilters): Promise<MediaItem[]> {
  try {
    const url = new URL('/api/media', window.location.origin);
    
    if (filters) {
      if (filters.filter) {
        // Use new filter spec with filter ID
        url.searchParams.set('filter', filters.filter);
      } else {
        // Handle relative date parameters by sending as JSON in a special parameter
        if (filters.dateRangeType === 'relative') {
          const dateRangeFilter: any = {
            type: 'relative'
          };
          
          // Handle start relative date
          if (filters.startPeriod) {
            if (filters.endPeriod) {
              // Range mode with start and end
              dateRangeFilter.startRelative = {
                period: filters.startPeriod,
                offset: filters.startOffset,
                anchor: filters.startAnchor || 'start',
                count: filters.startCount,
                direction: filters.startDirection
              };
              dateRangeFilter.endRelative = {
                period: filters.endPeriod,
                offset: filters.endOffset,
                anchor: filters.endAnchor || 'start',
                count: filters.endCount,
                direction: filters.endDirection
              };
            } else {
              // Simple mode (backward compatibility)
              dateRangeFilter.period = filters.startPeriod;
              dateRangeFilter.offset = filters.startOffset;
              dateRangeFilter.anchor = filters.startAnchor || 'start';
              dateRangeFilter.count = filters.startCount;
              dateRangeFilter.direction = filters.startDirection;
            }
          }
          
          // Send the date range as a JSON parameter that server can parse
          url.searchParams.set('dateRangeFilter', JSON.stringify(dateRangeFilter));
        } else {
          // Fallback to legacy fixed date parameters
          if (filters.startDate) {
            url.searchParams.set('startDate', filters.startDate);
          }
          if (filters.endDate) {
            url.searchParams.set('endDate', filters.endDate);
          }
        }
        
        if (filters.labels && filters.labels.length > 0) {
          url.searchParams.set('labels', filters.labels.join(','));
        }
        if (filters.mediaTypes && filters.mediaTypes.length > 0) {
          url.searchParams.set('mediaTypes', filters.mediaTypes.join(','));
        }
      }
    }
    
    const response = await fetch(url.toString());
    if (!response.ok) {
      throw new Error(`Failed to fetch media items: ${response.statusText}`);
    }
    return await response.json();
  } catch (error) {
    console.error('Error fetching media items:', error);
    return [];
  }
}

/**
 * Fetches transcription status from the API
 * @returns Promise with array of transcription statuses
 */
export async function fetchTranscriptionStatus(): Promise<TranscriptionStatus[]> {
  try {
    const response = await fetch('/api/transcription/status');
    if (!response.ok) {
      throw new Error(`Failed to fetch transcription status: ${response.statusText}`);
    }
    const data = await response.json();
    return Array.isArray(data) ? data : [];
  } catch (error) {
    console.error('Error fetching transcription status:', error);
    return [];
  }
}

/**
 * Updates labels for a media item
 * @param id Media item ID
 * @param labels New labels array
 * @returns Promise with updated media item
 */
export async function updateLabels(id: string, labels: string[]): Promise<MediaItem | null> {
  try {
    const response = await fetch('/api/labels/update', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        id,
        labels
      })
    });

    if (!response.ok) {
      throw new Error(`Failed to update labels: ${response.statusText}`);
    }

    return await response.json();
  } catch (error) {
    console.error('Error updating labels:', error);
    return null;
  }
}

/**
 * Fetches the filters configuration from the API
 * @returns Promise with filters configuration
 */
export async function fetchFiltersConfig(): Promise<FiltersConfig | null> {
  try {
    const response = await fetch('/api/filters');
    if (!response.ok) {
      throw new Error(`Failed to fetch filters config: ${response.statusText}`);
    }
    return await response.json();
  } catch (error) {
    console.error('Error fetching filters config:', error);
    return null;
  }
}