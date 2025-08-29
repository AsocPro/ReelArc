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
  type: 'photo' | 'audio' | 'video' | 'note';
  timestamp: string;
  duration?: number;
  filename: string;
  transcription: string;
  notes: string;
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

export interface ViewConfig {
  id: string;
  label: string;
  component: any;
  icon?: string;
}

export interface TabConfig {
  id: string;
  label: string;
  disabled?: boolean;
}

export interface MediaFilters {
  startDate?: string;
  endDate?: string;
  labels?: string[];
  mediaTypes?: string[];
  filter?: string;
  // Relative date parameters
  dateRangeType?: string;
  // Start relative date
  startPeriod?: string;
  startOffset?: number;
  startAnchor?: string;
  startCount?: number;
  startDirection?: string;
  // End relative date
  endPeriod?: string;
  endOffset?: number;
  endAnchor?: string;
  endCount?: number;
  endDirection?: string;
}

export interface FilterDateRange {
  type: 'fixed' | 'relative';
  // For fixed dates
  startDate?: string;
  endDate?: string;
  // For simple relative dates (backward compatibility)
  period?: string;
  offset?: number;
  anchor?: string;
  count?: number;
  direction?: string;
  // For complex relative date ranges
  startRelative?: {
    period: string;
    offset?: number;
    anchor?: string;
    count?: number;
    direction?: string;
  };
  endRelative?: {
    period: string;
    offset?: number;
    anchor?: string;
    count?: number;
    direction?: string;
  };
}

export interface FilterCriteria {
  dateRange?: FilterDateRange;
  labels: string[];
  mediaTypes: string[];
}

export interface Filter {
  id: string;
  name: string;
  description: string;
  type: string;
  icon: string;
  enabled: boolean;
  criteria: FilterCriteria;
}

export interface FiltersConfig {
  version: string;
  metadata: {
    name: string;
    description: string;
    created: string;
    updated: string;
  };
  defaultFilter: string;
  pinnedFilters: string[];
  filters: Filter[];
}

export interface ZoomLevel {
  id: string;
  label: string;
  duration: number; // Duration in milliseconds
  snapTo: 'hour' | 'day' | 'week' | 'month' | 'year';
}

// Swimlane types
export interface Swimlane {
  id: string;
  name: string;
  description: string;
  icon: string;
  enabled: boolean;
  color?: string;
  order: number;
}

export interface SwimlanesConfig {
  version: string;
  metadata: {
    name: string;
    description: string;
    created: string;
    updated: string;
  };
  swimlanes: Swimlane[];
}