import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, input, output, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { NotificationService } from '../../../core/notifications/notification.service';
import { User } from '../../../core/users/user.models';
import { UsersService } from '../../../core/users/users.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-delete-user-modal',
  imports: [Modal],
  styleUrl: './delete-user-modal.css',
  templateUrl: './delete-user-modal.html',
})
export class DeleteUserModal {
  readonly open = input(false);
  readonly user = input<User | null>(null);
  readonly closed = output<void>();
  readonly deleted = output<string>();

  private readonly users = inject(UsersService);
  private readonly notifications = inject(NotificationService);

  protected readonly loading = signal(false);

  protected async confirm(): Promise<void> {
    const user = this.user();
    if (!user) {
      return;
    }

    this.loading.set(true);

    try {
      await firstValueFrom(this.users.deleteUser(user.id));
      this.deleted.emit(user.username);
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
        return 'El usuario no existe';
      }
    }
    return 'No se pudo eliminar el usuario';
  }
}
