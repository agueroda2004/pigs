import { Injectable, Signal, signal } from '@angular/core';

export type NotificationType = 'info' | 'success' | 'error';

export interface Notification {
  id: number;
  type: NotificationType;
  message: string;
}

export interface NotificationOptions {
  duration?: number;
}

const DEFAULT_DURATIONS: Record<NotificationType, number> = {
  info: 4000,
  success: 4000,
  error: 6000,
};

@Injectable({ providedIn: 'root' })
export class NotificationService {
  private readonly items = signal<Notification[]>([]);
  private readonly timers = new Map<number, ReturnType<typeof setTimeout>>();
  private nextId = 1;

  readonly notifications: Signal<Notification[]> = this.items.asReadonly();

  show(type: NotificationType, message: string, options?: NotificationOptions): number {
    const id = this.nextId++;
    this.items.update((current) => [...current, { id, type, message }]);

    const duration = options?.duration ?? DEFAULT_DURATIONS[type];
    if (duration > 0) {
      this.timers.set(
        id,
        setTimeout(() => this.dismiss(id), duration),
      );
    }

    return id;
  }

  info(message: string, options?: NotificationOptions): number {
    return this.show('info', message, options);
  }

  success(message: string, options?: NotificationOptions): number {
    return this.show('success', message, options);
  }

  error(message: string, options?: NotificationOptions): number {
    return this.show('error', message, options);
  }

  dismiss(id: number): void {
    const timer = this.timers.get(id);
    if (timer) {
      clearTimeout(timer);
      this.timers.delete(id);
    }
    this.items.update((current) => current.filter((notification) => notification.id !== id));
  }
}
