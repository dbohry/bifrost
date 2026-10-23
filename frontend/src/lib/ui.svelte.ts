/**
 * Things one part of the screen asks another to do. The bottom bar on a phone
 * is outside the dashboard, but its Search and Add buttons act on it.
 */
class Ui {
  /** Set by the bottom bar. The dashboard does it, then clears it. */
  request = $state<'search' | 'add' | null>(null);
}

export const ui = new Ui();
