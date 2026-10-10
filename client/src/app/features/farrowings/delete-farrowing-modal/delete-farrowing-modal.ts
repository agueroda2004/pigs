import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, input, output, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { Farrowing } from '../../../core/farrowings/farrowing.models';
import { FarrowingsService } from '../../../core/farrowings/farrowings.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-delete-farrowing-modal',
  imports: [Modal],
  styleUrl: './delete-farrowing-modal.css',
  templateUrl: './delete-farrowing-modal.html',
})
export class DeleteFarrowingModal {
  readonly open = input(false);
  readonly farrowing = input<Farrowing | null>(null);
  readonly closed = output<void>();
  readonly deleted = output<Farrowing>();

  private readonly farrowings = inject(FarrowingsService);
  private readonly notifications = inject(NotificationService);

  protected readonly loading = signal(false);

  protected async confirm(): Promise<void> {
    const farrowing = this.farrowing();
    if (!farrowing) {
      return;
    }

    this.loading.set(true);

    try {
      await firstValueFrom(this.farrowings.deleteFarrowing(farrowing.id));
      this.deleted.emit(farrowing);
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
        return 'El parto no existe';
      }
    }
    return 'No se pudo eliminar el parto';
  }
}
