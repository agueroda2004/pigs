import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, input, output, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { PartialWeaging } from '../../../core/partial-weagings/partial-weaging.models';
import { PartialWeagingsService } from '../../../core/partial-weagings/partial-weagings.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-delete-partial-weaging-modal',
  imports: [Modal],
  styleUrl: './delete-partial-weaging-modal.css',
  templateUrl: './delete-partial-weaging-modal.html',
})
export class DeletePartialWeagingModal {
  readonly open = input(false);
  readonly weaging = input<PartialWeaging | null>(null);
  readonly sowCode = input<string | null>(null);
  readonly closed = output<void>();
  readonly deleted = output<PartialWeaging>();

  private readonly partialWeagings = inject(PartialWeagingsService);
  private readonly notifications = inject(NotificationService);

  protected readonly loading = signal(false);

  protected async confirm(): Promise<void> {
    const weaging = this.weaging();
    if (!weaging) {
      return;
    }

    this.loading.set(true);

    try {
      await firstValueFrom(this.partialWeagings.deletePartialWeaging(weaging.id));
      this.deleted.emit(weaging);
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
        return 'El destete parcial no existe';
      }
    }
    return 'No se pudo eliminar el destete parcial';
  }
}
