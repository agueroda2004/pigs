export interface FlipUpOptions {
  gap?: number;
  viewportHeight?: number;
}

// shouldFlipUp decides whether an overlay should open upwards instead of downwards.
// It flips only when the panel does not fit below the trigger and there is more room above.
export function shouldFlipUp(
  trigger: HTMLElement,
  panel: HTMLElement,
  options: FlipUpOptions = {},
): boolean {
  const gap = options.gap ?? 8;
  const viewportHeight =
    options.viewportHeight ?? (typeof window === 'undefined' ? 0 : window.innerHeight);

  const triggerRect = trigger.getBoundingClientRect();
  const panelHeight = panel.getBoundingClientRect().height;
  const spaceBelow = viewportHeight - triggerRect.bottom - gap;
  const spaceAbove = triggerRect.top - gap;

  return spaceBelow < panelHeight && spaceAbove > spaceBelow;
}
