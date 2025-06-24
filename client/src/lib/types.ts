export interface MediaItem {
  id: string;
  type: 'photo' | 'audio' | 'video';
  timestamp: string;
  duration?: number;
  filename: string;
  transcription: string;
  labels: string[];
}

export interface TimelineItem {
  id: string;
  content: string;
  start: string;
  end?: string;
  type: string;
  className?: string;
  mediaItem: MediaItem;
}

export interface Label {
  id: string;
  text: string;
}