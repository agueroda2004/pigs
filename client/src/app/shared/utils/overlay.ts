export interface FlipUpOptions {
  gap?: number;
  viewportHeight?: number;
}

// OverlayPlacement holds the fixed coordinates where an overlay panel is shown.
// It keeps the panel anchored to its trigger while escaping any scroll container.
export interface OverlayPlacement {
  top: number;
  left: number;
  width: number;
  dropUp: boolean;
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

// computeOverlayPlacement returns the fixed top, left and width for an overlay panel.
// Positioning the panel with these viewport coordinates keeps it visible even when
// an ancestor scroll container would otherwise clip it.
export function computeOverlayPlacement(
  trigger: HTMLElement,
  panel: HTMLElement,
  options: FlipUpOptions = {},
): OverlayPlacement {
  const gap = options.gap ?? 8;
  const triggerRect = trigger.getBoundingClientRect();
  const panelHeight = panel.getBoundingClientRect().height;
  const dropUp = shouldFlipUp(trigger, panel, options);
  const top = dropUp ? triggerRect.top - panelHeight - gap : triggerRect.bottom + gap;

  return {
    top,
    left: triggerRect.left,
    width: triggerRect.width,
    dropUp,
  };
}
