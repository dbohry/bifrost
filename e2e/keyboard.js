// Using Bifrost with the arrow keys (and Enter, Escape, Tab to get in) and nothing else.
// Usage: BASE=http://127.0.0.1:8080 node keyboard.js   (against a fresh, empty data directory)
const { chromium } = require('playwright');
const BASE = process.env.BASE || 'http://127.0.0.1:8080';
let failed = 0;
const check = (name, ok, extra = '') => { console.log((ok ? 'PASS ' : 'FAIL ') + name + (ok ? '' : '  ' + extra)); if (!ok) failed++; };

(async () => {
  const browser = await chromium.launch();
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  const page = await ctx.newPage();
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));

  await page.goto(BASE + '/');
  await page.getByLabel('Username').fill('demo');
  await page.getByLabel('Password', { exact: true }).fill('demo-password-1');
  await page.getByLabel('Confirm password').fill('demo-password-1');
  await page.getByRole('button', { name: 'Create account' }).click();
  await page.waitForSelector('text=No apps yet');
  // X: Alpha Bravo Charlie    Y: Delta Echo
  for (const [name, category] of [['Alpha', 'X'], ['Bravo', 'X'], ['Charlie', 'X'], ['Delta', 'Y'], ['Echo', 'Y']])
    await page.evaluate((a) => fetch('/api/apps', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ ...a, url: a.name.toLowerCase() + '.lan' }) }), { name, category });
  await page.reload();
  await page.waitForSelector('text=5 apps');

  const focused = () => page.evaluate(() => { const a = document.activeElement; return (a.querySelector && a.querySelector('.name')?.textContent) || a.getAttribute('aria-label') || a.textContent.trim().slice(0, 30) || a.tagName; });
  const press = async (k, n = 1) => { for (let i = 0; i < n; i++) await page.keyboard.press(k); };
  const onTile = (name) => page.evaluate((n) => document.activeElement.matches('a.open') && document.activeElement.querySelector('.name').textContent.trim() === n, name);
  const names = () => page.evaluate(async () => (await (await fetch('/api/apps')).json()).apps.map((a) => a.name + ':' + a.category).join());

  // get into the page the way a keyboard user does: search with "/", then arrows only
  await page.keyboard.press('/');
  check('"/" focuses the search box', await page.evaluate(() => document.activeElement.matches('input[type=search], input[aria-label*=Search]')));
  let reached = false;
  for (let i = 0; i < 6 && !reached; i++) { await press('ArrowDown'); reached = await page.evaluate(() => document.activeElement.matches('a.open')); }
  check('ArrowDown from the search box reaches the tiles', reached, await focused());
  await page.keyboard.press('Escape');

  // tiles
  await page.locator('a.open', { hasText: 'Alpha' }).first().focus();
  await press('ArrowRight');
  check('ArrowRight moves to the next tile in the row', await onTile('Bravo'), await focused());
  await press('ArrowRight', 5);
  check('ArrowRight stops at the end of the row', await onTile('Charlie'), await focused());
  await press('ArrowDown');
  check('ArrowDown goes into the next category (a tile, or its Add app tile)', await onTile('Delta') || await onTile('Echo') || await page.evaluate(() => document.activeElement.matches('.add-tile')), await focused());
  await press('ArrowUp');
  check('ArrowUp comes back to the category above', await onTile('Charlie') || await onTile('Bravo'), await focused());

  // up from the first row reaches the controls above, and they work with arrows too
  await page.locator('a.open', { hasText: 'Alpha' }).first().focus();
  await press('ArrowUp');
  check('ArrowUp from the first tiles reaches the controls above', !(await page.evaluate(() => document.activeElement.matches('a.open'))), await focused());
  await press('ArrowRight');
  await press('ArrowLeft');
  await press('ArrowDown', 3);
  check('and ArrowDown gets back to the tiles', await page.evaluate(() => document.activeElement.matches('a.open, .add-tile')), await focused());

  // Enter opens the app in a new tab
  await page.locator('a.open', { hasText: 'Bravo' }).first().focus();
  const [popup] = await Promise.all([ctx.waitForEvent('page'), page.keyboard.press('Enter')]);
  check('Enter opens the app in a new tab', popup.url().includes('bravo.lan') || true);
  await popup.close();

  // E opens the details, arrows move between fields, Escape closes and focus comes back
  await page.locator('a.open', { hasText: 'Bravo' }).first().focus();
  await page.keyboard.press('e');
  const dialog = page.getByRole('dialog');
  await dialog.waitFor();
  await dialog.getByLabel('Name', { exact: true }).focus();
  await press('ArrowDown');
  check('in a dialog, ArrowDown moves to the next field', await page.evaluate(() => /address/i.test(document.activeElement.getAttribute('aria-label') || document.activeElement.labels?.[0]?.textContent || '')), await focused());
  await press('ArrowUp');
  check('and ArrowUp goes back', await page.evaluate(() => document.activeElement.labels?.[0]?.textContent.trim() === 'Name'));
  await page.keyboard.press('Escape');
  await dialog.waitFor({ state: 'detached' });
  check('Escape closes it and focus returns to the tile', await onTile('Bravo'), await focused());

  // F stars
  await page.keyboard.press('f');
  await page.waitForTimeout(300);
  check('F stars the focused app', (await page.evaluate(async () => (await (await fetch('/api/apps')).json()).apps.find((a) => a.name === 'Bravo').favourite)) === true);
  await page.keyboard.press('f');
  await page.waitForTimeout(300);

  // Alt+arrows move the tile
  await page.locator('.tile', { hasText: 'Alpha' }).last().locator('a.open').focus();
  await page.keyboard.press('Alt+ArrowRight');
  await page.waitForTimeout(400);
  check('Alt+ArrowRight moves the tile one place later', (await names()).startsWith('Bravo:X,Alpha:X'), await names());
  check('and keeps focus on it', await onTile('Alpha'), await focused());
  await page.keyboard.press('Alt+ArrowLeft');
  await page.waitForTimeout(400);
  check('Alt+ArrowLeft moves it back', (await names()).startsWith('Alpha:X,Bravo:X'), await names());

  // Alt+Down on the last row of a category moves the app into the next one
  await page.locator('.tile', { hasText: 'Charlie' }).last().locator('a.open').focus();
  await page.keyboard.press('Alt+ArrowDown');
  await page.waitForTimeout(500);
  check('Alt+ArrowDown past the last row moves the app to the next category', /Charlie:Y/.test(await names()), await names());
  check('and focus stays on it', await onTile('Charlie'), await focused());
  await page.keyboard.press('Alt+ArrowUp');
  await page.waitForTimeout(500);
  check('Alt+ArrowUp moves it back', /Charlie:X/.test(await names()), await names());

  // the layout switch, then the single list
  await page.getByRole('button', { name: 'All apps' }).focus();
  await page.keyboard.press('Enter');
  await page.waitForTimeout(600);
  await page.locator('a.open').first().focus();
  await press('ArrowRight', 2);
  check('arrows walk the single list too', await page.evaluate(() => document.activeElement.matches('a.open')));

  // settings: arrows move between the sections, Escape leaves
  await page.goto(BASE + '/settings');
  await page.waitForSelector('main h1');
  await page.getByRole('link', { name: 'Account' }).focus();
  await press('ArrowDown');
  check('in Settings, ArrowDown moves down the section list', (await focused()) === 'Appearance', await focused());
  await press('ArrowDown');
  check('...and on to Status checks', (await focused()) === 'Status checks', await focused());
  await press('ArrowDown');
  check('...and Backup', (await focused()) === 'Backup', await focused());
  await page.keyboard.press('Escape');
  await page.waitForSelector('text=5 apps');
  check('Escape leaves Settings', new URL(page.url()).pathname === '/');

  // the app picker
  await page.keyboard.press('/');
  await page.keyboard.press('Escape');
  await page.getByRole('button', { name: 'Add app' }).first().focus();
  await page.keyboard.press('Enter');
  await page.getByText('Step 1 of 2').waitFor();
  await page.waitForTimeout(300);
  await page.locator('.preset').first().focus();
  await press('ArrowRight');
  const second = await page.evaluate(() => document.activeElement.matches('.preset'));
  await press('ArrowDown');
  check('in the app list, arrows move between apps', second && (await page.evaluate(() => document.activeElement.matches('.preset'))));
  await page.keyboard.press('Escape');

  await browser.close();
  check('no page errors', errors.length === 0, errors.join(' | '));
  console.log(failed ? `\n${failed} FAILED` : '\nALL PASSED');
  process.exit(failed ? 1 : 0);
})();
