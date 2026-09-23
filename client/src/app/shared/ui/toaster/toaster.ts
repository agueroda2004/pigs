import { Component, inject } from '@angular/core';

import {
  NotificationService,
  NotificationType,
} from '../../../core/notifications/notification.service';

@Component({
  selector: 'app-toaster',
  template: `
    <div
      class="pointer-events-none fixed top-4 right-4 z-[100] flex w-full max-w-sm flex-col gap-3"
    >
      @for (notification of notifications.notifications(); track notification.id) {
        <div
          [attr.role]="role(notification.type)"
          [attr.aria-live]="ariaLive(notification.type)"
          class="pointer-events-auto flex items-start gap-3 rounded-lg border border-l-4 border-border bg-card p-4 shadow-lg"
          [class]="accentClass(notification.type)"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="h-5 w-5 shrink-0"
            [class]="textClass(notification.type)"
            aria-hidden="true"
          >
            @switch (notification.type) {
              @case ('success') {
                <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
                <path d="m9 11 3 3L22 4" />
              }
              @case ('error') {
                <circle cx="12" cy="12" r="10" />
                <path d="m15 9-6 6M9 9l6 6" />
              }
              @default {
                <circle cx="12" cy="12" r="10" />
                <path d="M12 16v-4M12 8h.01" />
              }
            }
          </svg>

          <p class="flex-1 text-sm">{{ notification.message }}</p>

          <button
            type="button"
            (click)="notifications.dismiss(notification.id)"
            aria-label="Cerrar notificación"
            class="inline-flex h-5 w-5 shrink-0 items-center justify-center rounded text-muted-foreground transition hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="h-4 w-4"
            >
              <path d="M18 6 6 18M6 6l12 12" />
            </svg>
          </button>
        </div>
      }
    </div>
  `,
})
export class Toaster {
  protected readonly notifications = inject(NotificationService);

  protected accentClass(type: NotificationType): string {
    switch (type) {
      case 'success':
        return 'border-l-success';
      case 'error':
        return 'border-l-danger';
      default:
        return 'border-l-info';
    }
  }

  protected textClass(type: NotificationType): string {
    switch (type) {
      case 'success':
        return 'text-success';
      case 'error':
        return 'text-danger';
      default:
        return 'text-info';
    }
  }

  protected role(type: NotificationType): string {
    return type === 'error' ? 'alert' : 'status';
  }

  protected ariaLive(type: NotificationType): string {
    return type === 'error' ? 'assertive' : 'polite';
  }
}
