/** The letter shown on an app's tile. */
export function initial(name: string): string {
  const first = [...name.trim()][0];
  return first ? first.toUpperCase() : '?';
}

/**
 * A background colour for an app's tile, steady for the life of the app.
 * The golden angle spreads neighbouring ids across the colour wheel, and a
 * fixed lightness keeps the white letter readable in every theme.
 */
export function tileColour(id: number): string {
  const hue = Math.round((id * 137.508) % 360);
  return `hsl(${hue} 50% 36%)`;
}

export function greeting(date = new Date()): string {
  const hour = date.getHours();
  if (hour < 12) return 'Good morning';
  if (hour < 18) return 'Good afternoon';
  return 'Good evening';
}
