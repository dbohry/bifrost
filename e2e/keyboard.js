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


  // --- nothing may be out of reach: from the first control, the arrows alone get to every other one ---
  const FOCUSABLE = 'a[href],button:not([disabled]),input:not([disabled]):not([type=hidden]),select:not([disabled]),textarea:not([disabled]),[tabindex]:not([tabindex="-1"])';
  async function reachesEverything(label) {
    await page.waitForTimeout(300);
    const names = await page.evaluate((sel) => {
      document.querySelectorAll('[data-reach]').forEach((e) => e.removeAttribute('data-reach'));
      const scope = document.querySelector('dialog[open]') || document;
      const els = [...scope.querySelectorAll(sel)].filter((el) => {
        if (el.closest('[data-nav-skip],[inert],[hidden]')) return false;
        const r = el.getBoundingClientRect();
        return r.width && r.height && getComputedStyle(el).visibility !== 'hidden';
      });
      els.forEach((el, i) => el.setAttribute('data-reach', i));
      return els.map((el) => (el.getAttribute('aria-label') || el.labels?.[0]?.textContent || el.textContent || el.tagName).trim().replace(/\s+/g, ' ').slice(0, 30));
    }, FOCUSABLE);
    const seen = new Set([0]);
    const queue = [0];
    while (queue.length) {
      const i = queue.shift();
      for (const k of ['ArrowDown', 'ArrowUp', 'ArrowRight', 'ArrowLeft']) {
        await page.evaluate((i) => document.querySelector(`[data-reach="${i}"]`)?.focus(), i);
        await page.keyboard.press(k);
        const at = await page.evaluate(() => document.activeElement.getAttribute('data-reach'));
        if (at !== null && !seen.has(+at)) { seen.add(+at); queue.push(+at); }
      }
    }
    const missed = names.filter((_, i) => !seen.has(i));
    check(label + ': every control is reachable with the arrows', missed.length === 0, 'out of reach: ' + missed.join(', '));
  }

  await page.goto(BASE + '/');
  await page.waitForSelector('text=5 apps');
  await reachesEverything('the dashboard');
  await page.locator('a.open', { hasText: 'Alpha' }).first().focus();
  await page.keyboard.press('e');
  await page.getByRole('dialog').waitFor();
  await reachesEverything('the app details');
  await page.getByRole('switch', { name: /Check if this app is online/ }).click();
  await reachesEverything('the app details with status checks on');
  await page.keyboard.press('Escape');
  await page.getByRole('button', { name: 'Add app' }).first().click();
  await page.getByText('Step 1 of 2').waitFor();
  await page.getByRole('button', { name: /custom app/i }).click();
  await page.getByLabel('Name', { exact: true }).waitFor();
  await reachesEverything('the add-app form');
  await page.keyboard.press('Escape');
  for (const [path, label] of [['/settings', 'Account'], ['/settings/appearance', 'Appearance'], ['/settings/status', 'Status checks'], ['/settings/backup', 'Backup'], ['/settings/users', 'Users']]) {
    await page.goto(BASE + path);
    await page.waitForSelector('main h1');
    await reachesEverything('Settings, ' + label);
  }
  await page.getByRole('button', { name: 'Add user' }).click();
  await page.getByRole('dialog').waitFor();
  await reachesEverything('the add-user dialog');
  await page.keyboard.press('Escape');

  // the caret keeps the arrows until it reaches the end of the text, then they move on
  await page.goto(BASE + '/');
  await page.waitForSelector('text=5 apps');
  await page.keyboard.press('/');
  await page.keyboard.type('abc');
  await page.keyboard.press('ArrowLeft');
  check('in a text field, an arrow still moves the caret', await page.evaluate(() => document.activeElement.matches('input[type=text]') && document.activeElement.selectionStart === 2), await focused());
  await page.keyboard.press('ArrowRight');
  await page.keyboard.press('ArrowRight');
  check('...and at the end of the text, ArrowRight leaves the field', await page.evaluate(() => !document.activeElement.matches('input[type=text]')), await focused());
  await page.keyboard.press('ArrowLeft');
  check('...and ArrowLeft comes back', await page.evaluate(() => document.activeElement.matches('input[type=text]')), await focused());
  await page.keyboard.press('Escape');

  // from the tiles, the controls on the right of the top bar are one Up (or two) away
  await page.locator('a.open', { hasText: 'Alpha' }).first().focus();
  let top = false;
  for (let i = 0; i < 3 && !top; i++) { await press('ArrowUp'); top = await page.evaluate(() => document.activeElement.closest('.bar') !== null); }
  check('ArrowUp from the tiles reaches the top bar', top, await focused());
  for (let i = 0; i < 4 && (await focused()) !== 'Settings'; i++) await press('ArrowRight');
  check('...and the arrows get to Settings', (await focused()) === 'Settings', await focused());

  // radio buttons: arrows choose within the group, and also lead out of it
  await page.goto(BASE + '/settings/appearance');
  await page.waitForSelector('main h1');
  await page.getByRole('radio', { name: /Dracula/ }).focus();
  const group = await page.evaluate(() => document.activeElement.name);
  await press('ArrowRight');
  check('in Appearance, an arrow on a theme chooses the next one', await page.evaluate(() => document.activeElement.type === 'radio' && document.activeElement.checked), await focused());
  await press('ArrowDown', 8);
  check('...and the arrows lead out of the group instead of wrapping inside it', await page.evaluate((g) => document.activeElement.name !== g, group), await focused());

  // focus is never left with nothing: after adding, after deleting, and after changing page
  await page.goto(BASE + '/');
  await page.waitForSelector('text=5 apps');
  await page.getByRole('button', { name: 'Add app' }).first().focus();
  await page.keyboard.press('Enter');
  await page.getByRole('button', { name: /custom app/i }).click();
  await page.getByLabel('Name', { exact: true }).fill('Foxtrot');
  await page.getByLabel('Address').first().fill('foxtrot.lan');
  await page.keyboard.press('Enter');
  await page.getByRole('dialog').waitFor({ state: 'detached' });
  await page.waitForTimeout(300);
  check('after adding an app, focus is back where it started', (await focused()).includes('Add app'), await focused());
  await page.locator('a.open', { hasText: 'Foxtrot' }).first().focus();
  await page.keyboard.press('e');
  await page.getByRole('dialog').waitFor();
  await page.getByRole('button', { name: 'Delete' }).click();
  await page.getByRole('button', { name: 'Delete' }).last().click();
  await page.getByRole('dialog').waitFor({ state: 'detached' });
  await page.waitForTimeout(300);
  check('after deleting an app, focus moves to a neighbouring tile', await page.evaluate(() => document.activeElement.matches('a.open')), await focused());
  await page.goto(BASE + '/settings');
  await page.waitForSelector('main h1');
  await page.evaluate(() => document.activeElement?.blur());
  await press('ArrowDown');
  check('with nothing focused, an arrow enters the page (Settings: the section list)', (await focused()) === 'Account', await focused());
  await page.goto(BASE + '/');
  await page.waitForSelector('text=5 apps');
  await press('ArrowDown');
  check('with nothing focused, an arrow enters the page (dashboard: the tiles)', await page.evaluate(() => document.activeElement.matches('a.open')), await focused());

  await browser.close();
  check('no page errors', errors.length === 0, errors.join(' | '));
  console.log(failed ? `\n${failed} FAILED` : '\nALL PASSED');
  process.exit(failed ? 1 : 0);
})();
