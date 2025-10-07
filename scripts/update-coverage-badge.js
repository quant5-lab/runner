#!/usr/bin/env node
import { readFileSync, writeFileSync, existsSync } from 'fs';

/* Extract coverage % from coverage-summary.json */
function extractCoverage() {
  const coverageFile = './coverage/coverage-summary.json';

  /* Coverage file must exist - script should be run after vitest coverage */
  if (!existsSync(coverageFile)) {
    console.error('❌ Coverage file not found. Run "pnpm coverage" first.');
    return null;
  }

  /* Read coverage-summary.json */
  try {
    const summary = JSON.parse(readFileSync(coverageFile, 'utf8'));
    const totalCoverage = summary.total;

    if (!totalCoverage || !totalCoverage.lines) {
      console.error('❌ No coverage data found');
      return null;
    }

    /* Return line coverage percentage */
    return Math.round(totalCoverage.lines.pct * 10) / 10;
  } catch (error) {
    console.error('❌ Failed to read coverage file:', error.message);
    return null;
  }
}

/* Update README.md with coverage badge */
function updateReadme(coverage) {
  const readmePath = './README.md';

  if (!existsSync(readmePath)) {
    console.error('❌ README.md not found');
    return false;
  }

  let readme = readFileSync(readmePath, 'utf8');

  /* Generate badge markdown based on coverage threshold */
  const color = coverage >= 80 ? 'brightgreen' : coverage >= 60 ? 'yellow' : 'red';
  const badge = `![Coverage](https://img.shields.io/badge/coverage-${coverage}%25-${color})`;

  /* Replace existing badge or add new one - matches both valid numbers and NaN */
  const badgeRegex = /!\[Coverage\]\(https:\/\/img\.shields\.io\/badge\/coverage-([\d.]+|NaN)%25-\w+\)/;

  if (badgeRegex.test(readme)) {
    readme = readme.replace(badgeRegex, badge);
    console.log(`✅ Updated existing coverage badge: ${coverage}%`);
  } else {
    /* Add badge after first heading */
    const headingRegex = /(^#\s+.+$)/m;
    if (headingRegex.test(readme)) {
      readme = readme.replace(headingRegex, `$1\n\n${badge}`);
      console.log(`✅ Added new coverage badge: ${coverage}%`);
    } else {
      /* Prepend if no heading found */
      readme = `${badge}\n\n${readme}`;
      console.log(`✅ Prepended coverage badge: ${coverage}%`);
    }
  }

  writeFileSync(readmePath, readme, 'utf8');
  return true;
}

/* Main execution */
console.log('📊 Extracting test coverage...');
const coverage = extractCoverage();

if (coverage !== null) {
  console.log(`📈 Coverage: ${coverage}%`);
  if (updateReadme(coverage)) {
    console.log('✅ README.md updated successfully');
    process.exit(0);
  }
}

console.error('❌ Failed to update coverage badge');
process.exit(1);
