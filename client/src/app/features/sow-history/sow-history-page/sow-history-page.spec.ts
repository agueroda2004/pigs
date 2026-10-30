import { TestBed } from '@angular/core/testing';

import { SowHistoryPage } from './sow-history-page';

describe('SowHistoryPage', () => {
  it('renders the placeholder heading', () => {
    const fixture = TestBed.createComponent(SowHistoryPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Historial de cerdas');
  });
});
