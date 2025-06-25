export interface TranscriptEntry {
  start: number;
  end: number;
  text: string;
  segment: number;
  speaker?: string;
  metadata?: string;
}

export interface MediaItem {
  id: string;
  type: 'photo' | 'audio' | 'video';
  timestamp: string;
  duration?: number;
  filename: string;
  transcription: string;
  labels: string[];
  transcripts?: TranscriptEntry[];
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

export interface TranscriptionStatus {
  filename: string;
  status: 'queued' | 'processing' | 'completed' | 'failed';
  error?: string;
  timestamp: string;
}