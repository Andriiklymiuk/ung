import * as assert from 'node:assert';
import { Formatter } from '../../utils/formatting';

suite('Formatter Test Suite', () => {
  test('formatDate returns YYYY-MM-DD format', () => {
    const date = new Date('2025-01-15');
    const formatted = Formatter.formatDate(date);
    assert.strictEqual(formatted, '2025-01-15');
  });

  test('formatCurrency formats USD correctly', () => {
    const formatted = Formatter.formatCurrency(1234.56, 'USD');
    assert.strictEqual(formatted, '$1234.56');
  });

  test('formatCurrency formats EUR correctly', () => {
    const formatted = Formatter.formatCurrency(1234.56, 'EUR');
    assert.strictEqual(formatted, '€1234.56');
  });

  test('formatCurrency formats UAH correctly', () => {
    const formatted = Formatter.formatCurrency(1234.56, 'UAH');
    assert.strictEqual(formatted, '₴1234.56');
  });

  test('formatDuration converts seconds to human readable', () => {
    assert.strictEqual(Formatter.formatDuration(3661), '1h 1m');
    assert.strictEqual(Formatter.formatDuration(90), '1m 30s');
    assert.strictEqual(Formatter.formatDuration(45), '45s');
  });

  test('formatHours converts decimal hours', () => {
    assert.strictEqual(Formatter.formatHours(2.5), '2h 30m');
    assert.strictEqual(Formatter.formatHours(1.0), '1h');
    assert.strictEqual(Formatter.formatHours(0.5), '30m');
  });

  test('truncate shortens long text', () => {
    assert.strictEqual(Formatter.truncate('Hello World', 8), 'Hello...');
    assert.strictEqual(Formatter.truncate('Short', 10), 'Short');
  });

  test('truncate handles exact length', () => {
    assert.strictEqual(Formatter.truncate('12345', 5), '12345');
    assert.strictEqual(Formatter.truncate('123456', 5), 'tr...');
  });

  test('parseDate returns YYYY-MM-DD format', () => {
    const parsed = Formatter.parseDate('2025-01-15');
    assert.strictEqual(parsed, '2025-01-15');
  });

  test('getStatusBadge returns correct emoji', () => {
    assert.strictEqual(Formatter.getStatusBadge('pending'), '⏳');
    assert.strictEqual(Formatter.getStatusBadge('paid'), '✅');
    assert.strictEqual(Formatter.getStatusBadge('overdue'), '⚠️');
    assert.strictEqual(Formatter.getStatusBadge('draft'), '📝');
    assert.strictEqual(Formatter.getStatusBadge('cancelled'), '❌');
    assert.strictEqual(Formatter.getStatusBadge('unknown'), '•');
  });
});
