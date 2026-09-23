import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { CreateBreedModal } from './create-breed-modal';

class BreedsStub {
  createBreed = vi.fn(() => of({}));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('CreateBreedModal', () => {
  let stub: BreedsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new BreedsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [CreateBreedModal],
      providers: [
        { provide: BreedsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(CreateBreedModal).componentInstance as any;

  it('does not submit when the form is empty', async () => {
    const component = create();
    await component.submit();
    expect(stub.createBreed).not.toHaveBeenCalled();
  });

  it('creates a breed and emits the name', async () => {
    const component = create();
    const created = vi.fn();
    component.created.subscribe(created);
    component.form.setValue({ name: 'Duroc' });

    await component.submit();

    expect(stub.createBreed).toHaveBeenCalledWith({ name: 'Duroc' });
    expect(created).toHaveBeenCalledWith('Duroc');
  });

  it('shows an error toast when the name is taken', async () => {
    stub.createBreed = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 409 })));
    const component = create();
    component.form.setValue({ name: 'Duroc' });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('El nombre de la raza ya existe');
  });
});
