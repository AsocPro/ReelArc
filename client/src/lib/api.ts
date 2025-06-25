import type { MediaItem, TranscriptionStatus } from './types';

/**
 * Fetches media items from the API
 * @returns Promise with array of media items
 */
export async function fetchMediaItems(): Promise<MediaItem[]> {
  try {
    const response = await fetch('/api/media');
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
    return await response.json();
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
    // In a real implementation, this would be a PUT or PATCH request
    // For now, we'll just return a mock response
    return {
      id,
      labels,
      type: 'photo', // This would come from the server in a real implementation
      timestamp: new Date().toISOString(),
      filename: 'mock.jpg',
      transcription: ''
    };
  } catch (error) {
    console.error('Error updating labels:', error);
    return null;
  }
}