import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, input, output, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { Weaging } from '../../../core/weagings/weaging.models';
import { WeagingsService } from '../../../core/weagings/weagings.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-delete-weaging-modal',
  imports: [Modal],
  styleUrl: './delete-weaging-modal.css',
  templateUrl: './delete-weaging-modal.html',
})
export class DeleteWeagingModal {
  readonly open = input(false);
  readonly weaging = input<Weaging | null>(null);
  readonly sowCode = input<string | null>(null);
  readonly closed = output<void>();
  readonly deleted = output<Weaging>();

  private readonly weagings = inject(WeagingsService);
  private readonly notifications = inject(NotificationService);

  protected readonly loading = signal(false);

  protected async confirm(): Promise<void> {
    const weaging = this.weaging();
    if (!weaging || this.loading()) {
      return;
    }

    this.loading.set(true);

    try {
      await firstValueFrom(this.weagings.deleteWeaging(weaging.id));
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
      if (error.status === 409) {
        return 'No se puede eliminar el destete porque la cerda tiene montas o bajas posteriores';
      }
      if (error.status === 404) {
        return 'El destete no existe';
      }
    }
    return 'No se pudo eliminar el destete';
  }
}
