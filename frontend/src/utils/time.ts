// Normalize a <input type="time"> value to canonical 24h "HH:MM". The field
// renders according to the user's locale, so an en-US browser can hand it back
// as "06:00 PM" or "6:00" instead of "18:00".
export function normalizeTime24(value: string): string {
  const compact = value.trim().replace(/\s+/g, '');
  const m24 = /^([01]?[0-9]|2[0-3]):([0-5][0-9])$/.exec(compact);
  if (m24) {
    return `${String(Number(m24[1])).padStart(2, '0')}:${m24[2]}`;
  }
  const m12 = /^(0?[1-9]|1[0-2]):([0-5][0-9])(am|pm)$/i.exec(compact);
  if (m12) {
    let h = Number(m12[1]) % 12;
    if (m12[3].toLowerCase() === 'pm') {
      h += 12;
    }
    return `${String(h).padStart(2, '0')}:${m12[2]}`;
  }
  return value;
}