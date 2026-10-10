import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, input, output, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { Abortion } from '../../../core/abortions/abortion.models';
import { AbortionsService } from '../../../core/abortions/abortions.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-delete-abortion-modal',
  imports: [Modal],
  styleUrl: './delete-abortion-modal.css',
  templateUrl: './delete-abortion-modal.html',
})
export class DeleteAbortionModal {
  readonly open = input(false);
  readonly abortion = input<Abortion | null>(null);
  readonly closed = output<void>();
  readonly deleted = output<Abortion>();

  private readonly abortions = inject(AbortionsService);
  private readonly notifications = inject(NotificationService);

  protected readonly loading = signal(false);

  protected async confirm(): Promise<void> {
    const abortion = this.abortion();
    if (!abortion) {
      return;
    }

    this.loading.set(true);

    try {
      await firstValueFrom(this.abortions.deleteAbortion(abortion.id));
      this.deleted.emit(abortion);
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
        return 'El aborto no existe';
      }
    }
    return 'No se pudo eliminar el aborto';
  }
}
