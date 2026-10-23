const relative = new Intl.RelativeTimeFormat('en', { numeric: 'auto' });

/** "Active now", "Last seen yesterday", "Never signed in". */
export function lastSeen(unixSeconds: number | null, now = Date.now()): string {
  if (unixSeconds === null) return 'Never signed in';

  const seconds = Math.round((now - unixSeconds * 1000) / 1000);
  if (seconds < 15 * 60) return 'Active now';

  const minutes = Math.round(seconds / 60);
  if (minutes < 60) return `Last seen ${relative.format(-minutes, 'minute')}`;
  const hours = Math.round(minutes / 60);
  if (hours < 24) return `Last seen ${relative.format(-hours, 'hour')}`;
  const days = Math.round(hours / 24);
  if (days < 30) return `Last seen ${relative.format(-days, 'day')}`;
  const months = Math.round(days / 30);
  if (months < 12) return `Last seen ${relative.format(-months, 'month')}`;
  return `Last seen ${relative.format(-Math.round(months / 12), 'year')}`;
}
