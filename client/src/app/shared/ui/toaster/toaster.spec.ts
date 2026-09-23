import { TestBed } from '@angular/core/testing';

import { NotificationService } from '../../../core/notifications/notification.service';
import { Toaster } from './toaster';

describe('Toaster', () => {
  let service: NotificationService;

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [Toaster] }).compileComponents();
    service = TestBed.inject(NotificationService);
  });

  it('renders one toast per notification', () => {
    service.info('info message');
    service.error('error message');

    const fixture = TestBed.createComponent(Toaster);
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelectorAll('p')).toHaveLength(2);
  });

  it('applies the colour classes for the notification type', () => {
    service.success('ok');
    service.error('nope');

    const fixture = TestBed.createComponent(Toaster);
    fixture.detectChanges();

    const html: string = fixture.nativeElement.innerHTML;
    expect(html).toContain('border-l-success');
    expect(html).toContain('border-l-danger');
  });

  it('dismisses the notification when the close button is clicked', () => {
    service.info('info message');

    const fixture = TestBed.createComponent(Toaster);
    fixture.detectChanges();
    (fixture.nativeElement.querySelector('button') as HTMLButtonElement).click();
    fixture.detectChanges();

    expect(service.notifications()).toHaveLength(0);
    expect(fixture.nativeElement.querySelectorAll('p')).toHaveLength(0);
  });
});
