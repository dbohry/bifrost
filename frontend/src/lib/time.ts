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

/** How long ago, in a few words: "just now", "3 min", "2 h", "4 days". */
export function span(unixSeconds: number, now = Date.now()): string {
  const seconds = Math.max(0, Math.round((now - unixSeconds * 1000) / 1000));
  if (seconds < 60) return 'under a minute';
  const minutes = Math.round(seconds / 60);
  if (minutes < 60) return `${minutes} min`;
  const hours = Math.round(minutes / 60);
  if (hours < 24) return `${hours} h`;
  const days = Math.round(hours / 24);
  return `${days} ${days === 1 ? 'day' : 'days'}`;
}
