/// <reference types="svelte" />
/// <reference types="vite/client" />

declare module 'vis-timeline/standalone' {
  export class Timeline {
    constructor(container: HTMLElement, items: any, options?: any);
    on(event: string, callback: (properties: any) => void): void;
    setItems(items: any): void;
    fit(): void;
    destroy(): void;
  }
  
  export class DataSet {
    constructor(data: any[]);
  }
}