/**
 * check-address-test.ts — Tests shouldAnalyzeAddress
 */

import { shouldAnalyzeAddress } from './App';

const test1 = shouldAnalyzeAddress('杭州');
if (test1 !== false) {
  throw new Error(`short text '杭州' should not analyze, got ${test1}`);
}

const test2 = shouldAnalyzeAddress('浙江省杭州市');
if (test2 !== true) {
  throw new Error(`full address '浙江省杭州市' should analyze, got ${test2}`);
}

// Edge: exactly 5 chars (should pass)
const test3 = shouldAnalyzeAddress('12345');
if (test3 !== true) {
  throw new Error(`5-char string should analyze, got ${test3}`);
}

// Edge: exactly 4 chars (should fail)
const test4 = shouldAnalyzeAddress('1234');
if (test4 !== false) {
  throw new Error(`4-char string should not analyze, got ${test4}`);
}

// Whitespace-only
const test5 = shouldAnalyzeAddress('   ');
if (test5 !== false) {
  throw new Error(`whitespace-only should not analyze, got ${test5}`);
}

console.log('  shouldAnalyzeAddress results:', { test1, test2, test3, test4, test5 });
