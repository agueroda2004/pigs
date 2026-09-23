import { TestBed } from '@angular/core/testing';
import { vi } from 'vitest';

import { NotificationService } from './notification.service';

describe('NotificationService', () => {
  let service: NotificationService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(NotificationService);
  });

  afterEach(() => vi.useRealTimers());

  it('adds notifications with the matching type', () => {
    service.info('info message');
    service.success('success message');
    service.error('error message');

    expect(service.notifications().map((notification) => notification.type)).toEqual([
      'info',
      'success',
      'error',
    ]);
  });

  it('auto-dismisses info and success after 4 seconds', () => {
    vi.useFakeTimers();
    service.success('done');

    vi.advanceTimersByTime(4000);

    expect(service.notifications()).toHaveLength(0);
  });

  it('keeps error notifications until 6 seconds', () => {
    vi.useFakeTimers();
    service.error('boom');

    vi.advanceTimersByTime(4000);
    expect(service.notifications()).toHaveLength(1);

    vi.advanceTimersByTime(2000);
    expect(service.notifications()).toHaveLength(0);
  });

  it('dismisses immediately and cancels the pending timer', () => {
    vi.useFakeTimers();
    const id = service.info('hi');

    service.dismiss(id);
    expect(service.notifications()).toHaveLength(0);

    vi.advanceTimersByTime(10000);
    expect(service.notifications()).toHaveLength(0);
  });

  it('honours a custom duration', () => {
    vi.useFakeTimers();
    service.show('info', 'quick', { duration: 1000 });

    vi.advanceTimersByTime(1000);

    expect(service.notifications()).toHaveLength(0);
  });
});
