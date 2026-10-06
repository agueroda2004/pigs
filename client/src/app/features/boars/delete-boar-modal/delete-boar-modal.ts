import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, input, output, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { Boar } from '../../../core/boars/boar.models';
import { BoarsService } from '../../../core/boars/boars.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-delete-boar-modal',
  imports: [Modal],
  styleUrl: './delete-boar-modal.css',
  templateUrl: './delete-boar-modal.html',
})
export class DeleteBoarModal {
  readonly open = input(false);
  readonly boar = input<Boar | null>(null);
  readonly closed = output<void>();
  readonly deleted = output<Boar>();

  private readonly boars = inject(BoarsService);
  private readonly notifications = inject(NotificationService);

  protected readonly loading = signal(false);

  protected async confirm(): Promise<void> {
    const boar = this.boar();
    if (!boar) {
      return;
    }

    this.loading.set(true);

    try {
      await firstValueFrom(this.boars.deleteBoar(boar.id));
      this.deleted.emit(boar);
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      const serverMessage = (error.error as { error?: string } | null)?.error;
      if (serverMessage) {
        return serverMessage;
      }
      if (error.status === 404) {
        return 'El verraco no existe';
      }
    }
    return 'No se pudo eliminar el verraco';
  }
}
