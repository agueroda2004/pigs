import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { User } from '../../../core/users/user.models';
import { UsersService } from '../../../core/users/users.service';
import { CreateUserModal } from '../create-user-modal/create-user-modal';
import { DeleteUserModal } from '../delete-user-modal/delete-user-modal';
import { EditUserModal } from '../edit-user-modal/edit-user-modal';
import { UserCard } from '../user-card/user-card';

@Component({
  selector: 'app-users-page',
  imports: [CreateUserModal, EditUserModal, DeleteUserModal, UserCard],
  styleUrl: './users-page.css',
  templateUrl: './users-page.html',
})
export class UsersPage implements OnInit {
  private readonly usersService = inject(UsersService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);

  protected readonly currentUserId = computed(() => this.auth.user()?.id ?? null);
  protected readonly modalOpen = signal(false);
  protected readonly editingUser = signal<User | null>(null);
  protected readonly editOpen = signal(false);
  protected readonly deletingUser = signal<User | null>(null);
  protected readonly deleteOpen = signal(false);
  protected readonly users = signal<User[]>([]);
  protected readonly loading = signal(false);
  protected readonly error = signal(false);

  async ngOnInit(): Promise<void> {
    await this.loadUsers();
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(username: string): Promise<void> {
    this.modalOpen.set(false);
    this.notifications.success(`Usuario "${username}" creado correctamente`);
    await this.loadUsers();
  }

  protected openEdit(user: User): void {
    this.editingUser.set(user);
    this.editOpen.set(true);
  }

  protected closeEdit(): void {
    this.editOpen.set(false);
  }

  protected async onUpdated(): Promise<void> {
    this.editOpen.set(false);
    this.notifications.success('Usuario actualizado correctamente');
    await this.loadUsers();
  }

  protected openDelete(user: User): void {
    this.deletingUser.set(user);
    this.deleteOpen.set(true);
  }

  protected closeDelete(): void {
    this.deleteOpen.set(false);
  }

  protected async onDeleted(username: string): Promise<void> {
    this.deleteOpen.set(false);
    this.notifications.success(`Usuario "${username}" eliminado correctamente`);
    await this.loadUsers();
  }

  protected async loadUsers(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      this.users.set(await firstValueFrom(this.usersService.listUsers()));
    } catch {
      this.error.set(true);
    } finally {
      this.loading.set(false);
    }
  }
}
