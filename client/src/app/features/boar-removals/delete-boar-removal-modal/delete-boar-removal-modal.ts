import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, input, output, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { BoarRemoval } from '../../../core/boar-removals/boar-removal.models';
import { BoarRemovalsService } from '../../../core/boar-removals/boar-removals.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-delete-boar-removal-modal',
  imports: [Modal],
  styleUrl: './delete-boar-removal-modal.css',
  templateUrl: './delete-boar-removal-modal.html',
})
export class DeleteBoarRemovalModal {
  readonly open = input(false);
  readonly removal = input<BoarRemoval | null>(null);
  readonly boarCode = input<string | null>(null);
  readonly closed = output<void>();
  readonly deleted = output<void>();

  private readonly removals = inject(BoarRemovalsService);
  private readonly notifications = inject(NotificationService);

  protected readonly loading = signal(false);

  protected async confirm(): Promise<void> {
    const removal = this.removal();
    if (!removal) {
      return;
    }

    this.loading.set(true);

    try {
      await firstValueFrom(this.removals.deleteBoarRemoval(removal.id));
      this.deleted.emit();
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
        return 'La baja no existe';
      }
      if (error.status === 409) {
        return 'El verraco ya no está en el estado de la baja';
      }
    }
    return 'No se pudo eliminar la baja';
  }
}
