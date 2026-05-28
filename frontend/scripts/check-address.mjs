/**
 * check-address.mjs — TDD validation for shouldAnalyzeAddress (Task 9)
 */

import { execSync } from 'child_process';
import { fileURLToPath } from 'url';
import { dirname, resolve } from 'path';

const __dirname = dirname(fileURLToPath(import.meta.url));
const frontendRoot = resolve(__dirname, '..');

// 1. Build the frontend (TypeScript compilation check)
console.log('[check-address] Building frontend...');
execSync('npm run build', { cwd: frontendRoot, stdio: 'inherit' });

// 2. Run the actual test via tsx
console.log('[check-address] Running shouldAnalyzeAddress test...');
execSync('npx tsx src/check-address-test.ts', {
  cwd: frontendRoot,
  stdio: 'inherit',
});

console.log('[check-address] PASSED: shouldAnalyzeAddress works correctly.');
