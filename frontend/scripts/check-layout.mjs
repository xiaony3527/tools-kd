import { readFileSync, readdirSync, statSync } from 'fs';
import { join } from 'path';
import { execSync } from 'child_process';

// Ensure the frontend is built first
execSync('npm run build', { cwd: new URL('..', import.meta.url).pathname, stdio: 'inherit' });

// Find the most recently built JS bundle in dist/assets
const assetsDir = new URL('../dist/assets', import.meta.url).pathname;
const files = readdirSync(assetsDir)
  .filter(f => f.endsWith('.js'))
  .map(f => ({ name: f, mtime: statSync(join(assetsDir, f)).mtimeMs }))
  .sort((a, b) => b.mtime - a.mtime);

if (files.length === 0) {
  throw new Error('No JS bundle found in dist/assets');
}

const bundle = readFileSync(join(assetsDir, files[0].name), 'utf-8');

const checks = [
  ['topbar',         'topbar element'],
  ['packages-panel', 'left panel (packages-panel)'],
  ['quotes-panel',   'right panel (quotes-panel)'],
];

let pass = true;
for (const [token, label] of checks) {
  if (!bundle.includes(token)) {
    console.error(`  MISSING: ${label} (class "${token}" not found in bundle)`);
    pass = false;
  } else {
    console.log(`  FOUND: ${label}`);
  }
}

if (!pass) {
  process.exit(1);
}

console.log('\nLayout check PASSED: three-panel skeleton found in built bundle.');
