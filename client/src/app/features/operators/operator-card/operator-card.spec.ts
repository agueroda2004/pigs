import { TestBed } from '@angular/core/testing';

import { Operator } from '../../../core/operators/operator.models';
import { OperatorCard } from './operator-card';

function buildOperator(active = true): Operator {
  return {
    id: '1',
    name: 'Juan Pérez',
    active,
  };
}

describe('OperatorCard', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [OperatorCard],
    }).compileComponents();
  });

  it('renders the name and status', () => {
    const fixture = TestBed.createComponent(OperatorCard);
    fixture.componentRef.setInput('operator', buildOperator(true));
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Juan Pérez');
    expect(text).toContain('Activo');
  });

  it('shows the inactive status for inactive operators', () => {
    const fixture = TestBed.createComponent(OperatorCard);
    fixture.componentRef.setInput('operator', buildOperator(false));
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Inactivo');
  });

  it('hides the action buttons when editing is not allowed', () => {
    const fixture = TestBed.createComponent(OperatorCard);
    fixture.componentRef.setInput('operator', buildOperator());
    fixture.componentRef.setInput('canEdit', false);
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('button')).toBeNull();
  });

  it('emits the operator when the edit button is clicked', () => {
    const operator = buildOperator();
    const fixture = TestBed.createComponent(OperatorCard);
    fixture.componentRef.setInput('operator', operator);
    fixture.componentRef.setInput('canEdit', true);
    fixture.detectChanges();

    const emitted: Operator[] = [];
    fixture.componentInstance.editRequested.subscribe((value) => emitted.push(value));
    (fixture.nativeElement.querySelector('button') as HTMLButtonElement).click();

    expect(emitted).toEqual([operator]);
  });

  it('emits the operator when the delete button is clicked', () => {
    const operator = buildOperator();
    const fixture = TestBed.createComponent(OperatorCard);
    fixture.componentRef.setInput('operator', operator);
    fixture.componentRef.setInput('canEdit', true);
    fixture.detectChanges();

    const emitted: Operator[] = [];
    fixture.componentInstance.deleteRequested.subscribe((value) => emitted.push(value));
    const buttons = fixture.nativeElement.querySelectorAll('button');
    (buttons[1] as HTMLButtonElement).click();

    expect(emitted).toEqual([operator]);
  });
});
