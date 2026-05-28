/**
 * check-summary-test.ts — Imports renderSummary and verifies output
 *
 * This file is executed by tsx (via check-summary.mjs) and will fail
 * until renderSummary is added to App.ts (TDD red → green cycle).
 */

import { renderSummary } from './App';
import { main } from './wailsjs/go/models';

const value = renderSummary(new main.QuoteState({
  summary: { totalCount: 3, displayWeight: 4 },
  sto: { price: 8.9, billableWeight: 4, available: true, recommended: false, note: '' },
  bs: { price: 7.8, billableWeight: 4, available: true, recommended: true, note: '' },
  destination: { province: '', city: '', cities: [] },
  packages: [],
}));
if (value !== '3|8.9|7.8') {
  throw new Error(`summary render mismatch: got "${value}", expected "3|8.9|7.8"`);
}

console.log('  renderSummary output:', value);
