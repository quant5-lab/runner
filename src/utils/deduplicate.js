/* Deduplicate array items using custom key extraction */

export function deduplicate(items, keyGetter) {
  return Array.from(
    items.reduce((map, item) => {
      const key = keyGetter(item);
      if (!map.has(key)) map.set(key, item);
      return map;
    }, new Map()).values(),
  );
}
