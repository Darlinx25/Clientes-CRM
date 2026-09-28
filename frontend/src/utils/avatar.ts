// Subdued palette for contact/activity avatars, picked by a stable hash of the name.
const AVATAR_COLORS = [
  '#5c6bc0', // indigo
  '#26a69a', // teal
  '#ef5350', // red
  '#ffa726', // orange
  '#8d6e63', // brown
  '#ec407a', // pink
  '#66bb6a', // green
  '#42a5f5', // blue
];

// "Barbosa, Guillermo" -> "GB": the surname-first form is reordered so the
// given name leads the initials.
export function getInitials(name: string): string {
  let ordered = name;
  const commaIdx = name.indexOf(',');
  if (commaIdx !== -1) {
    ordered = name.slice(commaIdx + 1).trim() + ' ' + name.slice(0, commaIdx).trim();
  }
  const words = ordered.split(/\s+/).filter(Boolean).slice(0, 2);
  return words
    .map((w) => {
      const cp = w.codePointAt(0);
      return cp != null ? String.fromCodePoint(cp).toLocaleUpperCase() : '';
    })
    .join('');
}

export function avatarColorFor(name: string): string {
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = (hash * 31 + name.charCodeAt(i)) >>> 0;
  }
  return AVATAR_COLORS[hash % AVATAR_COLORS.length];
}