import { TestBed } from '@angular/core/testing';

import { ReportsPage } from './reports-page';

describe('ReportsPage', () => {
  it('renders the placeholder heading', () => {
    const fixture = TestBed.createComponent(ReportsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Reportes');
  });
});
