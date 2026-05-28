/**
 * check-sto-note.mjs — TDD validation for buildStoNote (Task 10)
 */

import { execSync } from 'child_process';
import { fileURLToPath } from 'url';
import { dirname, resolve } from 'path';

const __dirname = dirname(fileURLToPath(import.meta.url));
const frontendRoot = resolve(__dirname, '..');

// 1. Build the frontend (TypeScript compilation check)
console.log('[check-sto-note] Building frontend...');
execSync('npm run build', { cwd: frontendRoot, stdio: 'inherit' });

// 2. Run the actual test via tsx
console.log('[check-sto-note] Running buildStoNote test...');
execSync('npx tsx src/check-sto-note-test.ts', {
  cwd: frontendRoot,
  stdio: 'inherit',
});

console.log('[check-sto-note] PASSED: buildStoNote works correctly.');
