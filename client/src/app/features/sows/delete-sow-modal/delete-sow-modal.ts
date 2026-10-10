import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, input, output, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { NotificationService } from '../../../core/notifications/notification.service';
import { Sow } from '../../../core/sows/sow.models';
import { SowsService } from '../../../core/sows/sows.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-delete-sow-modal',
  imports: [Modal],
  styleUrl: './delete-sow-modal.css',
  templateUrl: './delete-sow-modal.html',
})
export class DeleteSowModal {
  readonly open = input(false);
  readonly sow = input<Sow | null>(null);
  readonly closed = output<void>();
  readonly deleted = output<Sow>();

  private readonly sows = inject(SowsService);
  private readonly notifications = inject(NotificationService);

  protected readonly loading = signal(false);

  protected async confirm(): Promise<void> {
    const sow = this.sow();
    if (!sow) {
      return;
    }

    this.loading.set(true);

    try {
      await firstValueFrom(this.sows.deleteSow(sow.id));
      this.deleted.emit(sow);
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
        return 'La cerda no existe';
      }
    }
    return 'No se pudo eliminar la cerda';
  }
}
