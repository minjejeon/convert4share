export function EventsOn(eventName: string, callback: (data?: any) => void): () => void;
export function EventsOff(eventName: string, ...additionalEventNames: string[]): void;
export function EventsEmit(eventName: string, ...data: any[]): void;
export function WindowSetAlwaysOnTop(b: boolean): void;
export function WindowShow(): void;
export function WindowUnminimise(): void;
