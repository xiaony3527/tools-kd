/**
 * check-summary.mjs — TDD validation for renderSummary (Task 8)
 *
 * Builds the frontend, then uses tsx to run a TypeScript verification
 * that renderSummary produces the expected string output.
 */

import { execSync } from 'child_process';
import { fileURLToPath } from 'url';
import { dirname, resolve } from 'path';

const __dirname = dirname(fileURLToPath(import.meta.url));
const frontendRoot = resolve(__dirname, '..');

// 1. Build the frontend (TypeScript compilation check)
console.log('[check-summary] Building frontend...');
execSync('npm run build', { cwd: frontendRoot, stdio: 'inherit' });

// 2. Run the actual test via tsx (handles TypeScript natively)
console.log('[check-summary] Running renderSummary test...');
execSync('npx tsx src/check-summary-test.ts', {
  cwd: frontendRoot,
  stdio: 'inherit',
});

console.log('[check-summary] PASSED: renderSummary works correctly.');
