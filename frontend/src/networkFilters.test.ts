import { describe, test, expect } from 'vitest';
import { resolveNetworkFilters, DEFAULT_NETWORK_FILTERS } from './networkFilters';

describe('resolveNetworkFilters', () => {
  test('returns defaults when nothing is stored', () => {
    expect(resolveNetworkFilters(null)).toEqual(DEFAULT_NETWORK_FILTERS);
    expect(resolveNetworkFilters(undefined)).toEqual(DEFAULT_NETWORK_FILTERS);
    expect(resolveNetworkFilters('')).toEqual(DEFAULT_NETWORK_FILTERS);
  });

  test('restores a full stored filter set', () => {
    const stored = JSON.stringify({
      showRelationships: true,
      showActivities: false,
    });
    expect(resolveNetworkFilters(stored)).toEqual({
      showRelationships: true,
      showActivities: false,
    });
  });

  test('falls back per field when the stored object is partial', () => {
    const stored = JSON.stringify({ showActivities: true });
    expect(resolveNetworkFilters(stored)).toEqual({
      ...DEFAULT_NETWORK_FILTERS,
      showActivities: true,
    });
  });

  test('returns defaults for corrupt JSON', () => {
    expect(resolveNetworkFilters('{not json')).toEqual(DEFAULT_NETWORK_FILTERS);
    expect(resolveNetworkFilters('null')).toEqual(DEFAULT_NETWORK_FILTERS);
  });

  test('does not share state between calls', () => {
    const first = resolveNetworkFilters(null);
    first.showActivities = true;
    expect(resolveNetworkFilters(null).showActivities).toBe(false);
  });
});
