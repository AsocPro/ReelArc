import { writable } from 'svelte/store';

// Store for tracking current media playback
export interface MediaPlaybackState {
  isPlaying: boolean;
  currentItem: string | null; // ID of the currently playing media item
  startTimestamp: string | null; // ISO timestamp when the media started
  currentTime: number; // Current playback position in seconds
}

// Initialize with default values
export const mediaPlayback = writable<MediaPlaybackState>({
  isPlaying: false,
  currentItem: null,
  startTimestamp: null,
  currentTime: 0
});