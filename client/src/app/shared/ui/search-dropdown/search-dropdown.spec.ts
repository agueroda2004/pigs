import { TestBed } from '@angular/core/testing';
import { vi } from 'vitest';

import { SearchDropdown } from './search-dropdown';

const options = [
  { value: '1', label: 'C-001' },
  { value: '2', label: 'C-002' },
  { value: '3', label: 'V-100' },
];

describe('SearchDropdown', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [SearchDropdown] }).compileComponents();
  });

  const create = () => {
    const fixture = TestBed.createComponent(SearchDropdown);
    fixture.componentRef.setInput('options', options);
    fixture.componentRef.setInput('placeholder', 'Seleccionar');
    fixture.detectChanges();
    return fixture;
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const toggle = (fixture: any) => {
    (fixture.nativeElement.querySelector('button') as HTMLButtonElement).click();
    fixture.detectChanges();
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const search = (fixture: any, term: string) => {
    const input = fixture.nativeElement.querySelector('input') as HTMLInputElement;
    input.value = term;
    input.dispatchEvent(new Event('input'));
    fixture.detectChanges();
    return input;
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const optionCount = (fixture: any) =>
    fixture.nativeElement.querySelectorAll('li[role="option"]').length;

  it('shows the placeholder when there is no value', () => {
    const fixture = create();

    expect(fixture.nativeElement.textContent).toContain('Seleccionar');
  });

  it('shows the selected label', () => {
    const fixture = create();
    fixture.componentInstance.writeValue('2');
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('C-002');
  });

  it('lists every option when opened without a query', () => {
    const fixture = create();
    toggle(fixture);

    expect(optionCount(fixture)).toBe(3);
  });

  it('filters the options by the typed term', () => {
    const fixture = create();
    toggle(fixture);

    search(fixture, 'C-00');

    expect(optionCount(fixture)).toBe(2);
  });

  it('shows the empty state when nothing matches', () => {
    const fixture = create();
    toggle(fixture);

    search(fixture, 'zzz');
    fixture.detectChanges();

    expect(optionCount(fixture)).toBe(0);
    expect(fixture.nativeElement.textContent).toContain('Sin resultados');
  });

  it('selects an option on click and closes the list', () => {
    const fixture = create();
    const changed = vi.fn();
    fixture.componentInstance.registerOnChange(changed);
    toggle(fixture);

    const first = fixture.nativeElement.querySelector(
      'li[role="option"] button',
    ) as HTMLButtonElement;
    first.click();
    fixture.detectChanges();

    expect(changed).toHaveBeenCalledWith('1');
    expect(fixture.nativeElement.querySelector('input')).toBeNull();
  });

  it('selects the highlighted option with the keyboard', () => {
    const fixture = create();
    const changed = vi.fn();
    fixture.componentInstance.registerOnChange(changed);
    toggle(fixture);
    const input = fixture.nativeElement.querySelector('input') as HTMLInputElement;

    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown' }));
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter' }));
    fixture.detectChanges();

    expect(changed).toHaveBeenCalledWith('2');
  });

  it('closes on escape', () => {
    const fixture = create();
    toggle(fixture);

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }));
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('input')).toBeNull();
  });

  it('does not open when disabled', () => {
    const fixture = create();
    fixture.componentInstance.setDisabledState(true);
    fixture.detectChanges();

    toggle(fixture);

    expect(fixture.nativeElement.querySelector('input')).toBeNull();
  });
});
