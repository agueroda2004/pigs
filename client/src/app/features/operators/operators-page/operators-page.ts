import { Component, OnInit, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Operator } from '../../../core/operators/operator.models';
import { OperatorsService } from '../../../core/operators/operators.service';
import { CreateOperatorModal } from '../create-operator-modal/create-operator-modal';
import { EditOperatorModal } from '../edit-operator-modal/edit-operator-modal';
import { OperatorCard } from '../operator-card/operator-card';

@Component({
  selector: 'app-operators-page',
  imports: [OperatorCard, CreateOperatorModal, EditOperatorModal],
  styleUrl: './operators-page.css',
  templateUrl: './operators-page.html',
})
export class OperatorsPage implements OnInit {
  private readonly operatorsService = inject(OperatorsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly modalOpen = signal(false);
  protected readonly editOpen = signal(false);
  protected readonly editingOperator = signal<Operator | null>(null);
  protected readonly operators = signal<Operator[]>([]);
  protected readonly loading = signal(false);
  protected readonly error = signal(false);

  async ngOnInit(): Promise<void> {
    await this.loadOperators();
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(name: string): Promise<void> {
    this.modalOpen.set(false);
    this.notifications.success(`Operador "${name}" creado correctamente`);
    await this.loadOperators();
  }

  protected openEdit(operator: Operator): void {
    this.editingOperator.set(operator);
    this.editOpen.set(true);
  }

  protected closeEdit(): void {
    this.editOpen.set(false);
  }

  protected async onUpdated(operator: Operator): Promise<void> {
    this.editOpen.set(false);
    this.notifications.success(`Operador "${operator.name}" actualizado correctamente`);
    await this.loadOperators();
  }

  protected async loadOperators(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      this.operators.set(await firstValueFrom(this.operatorsService.listOperators()));
    } catch {
      this.error.set(true);
    } finally {
      this.loading.set(false);
    }
  }
}
