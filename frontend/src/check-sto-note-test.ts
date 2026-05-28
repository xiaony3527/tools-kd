/**
 * check-sto-note-test.ts — Tests buildStoNote
 */

import { buildStoNote } from './App';

// Overweight + unavailable
const note1 = buildStoNote(51, false);
if (note1 !== '⚠ >50kg 仅发百世') {
  throw new Error(`overweight note mismatch: got "${note1}"`);
}

// Underweight + available
const note2 = buildStoNote(30, true);
if (note2 !== '') {
  throw new Error(`normal weight available should be empty, got "${note2}"`);
}

// Underweight + unavailable (should still be empty — only weight>50 triggers)
const note3 = buildStoNote(30, false);
if (note3 !== '') {
  throw new Error(`normal weight unavailable should be empty, got "${note3}"`);
}

// Overweight + available (edge case — weight>50 but available)
const note4 = buildStoNote(51, true);
if (note4 !== '') {
  throw new Error(`overweight but available should be empty, got "${note4}"`);
}

// Exactly 50 + unavailable (should still be empty — needs >50)
const note5 = buildStoNote(50, false);
if (note5 !== '') {
  throw new Error(`weight exactly 50 unavailable should be empty, got "${note5}"`);
}

console.log('  buildStoNote results:', { note1, note2, note3, note4, note5 });
