// Status checks: a badge on tiles, how often to look, and what counts as online.
// Usage: BASE=http://127.0.0.1:8080 node status.js   (against a fresh, empty data directory)
// It plays the part of the apps with small servers of its own on this machine.
const { chromium } = require('playwright');
const http = require('http');
const BASE = process.env.BASE || 'http://127.0.0.1:8080';
let failed = 0;
const check = (name, ok, extra = '') => { console.log((ok ? 'PASS ' : 'FAIL ') + name + (ok ? '' : '  ' + extra)); if (!ok) failed++; };

const listen = (port, code = 200) => new Promise((resolve) => {
  const s = http.createServer((req, res) => { res.statusCode = code; res.end('hi'); });
  s.listen(port, '127.0.0.1', () => resolve(s));
});
const close = (s) => new Promise((resolve) => { s.closeAllConnections?.(); s.close(resolve); });

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage({ viewport: { width: 1280, height: 900 } });
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));

  let good = await listen(19181);          // answers 200
  const broken = await listen(19182, 500); // answers 500

  await page.goto(BASE + '/');
  await page.getByLabel('Username').fill('demo');
  await page.getByLabel('Password', { exact: true }).fill('demo-password-1');
  await page.getByLabel('Confirm password').fill('demo-password-1');
  await page.getByRole('button', { name: 'Create account' }).click();
  await page.waitForSelector('text=No apps yet');

  // --- settings: presets, custom values, and refusals
  await page.goto(BASE + '/settings/status');
  await page.getByRole('heading', { name: 'Status checks' }).waitFor();
  check('the defaults are 1 minute and 5 seconds', await page.getByLabel('1 min', { exact: true }).isChecked() && await page.getByLabel('5 s', { exact: true }).isChecked());
  const presets = await page.locator('.segmented label').allTextContents();
  check('the choices are 1, 2, 5 minutes and custom; 1, 5, 10, 30 seconds and custom', presets.join('|') === '1 min|2 min|5 min|Custom|1 s|5 s|10 s|30 s|Custom', presets.join('|'));
  await page.locator('input[name=interval][value=custom]').check({ force: true });
  await page.getByLabel('Seconds').first().fill('5');
  await page.getByRole('button', { name: 'Save' }).click();
  await page.getByRole('alert').filter({ hasText: '10 to 86400' }).waitFor();
  check('a custom interval under 10 seconds is refused', true);
  await page.getByLabel('Seconds').first().fill('10');
  await page.locator('input[name=timeout][value=custom]').check({ force: true });
  await page.getByLabel('Seconds').last().fill('20');
  await page.getByRole('button', { name: 'Save' }).click();
  await page.getByRole('alert').filter({ hasText: 'cannot be longer' }).waitFor();
  check('giving up later than the check interval is refused', true);
  await page.getByLabel('Seconds').last().fill('2');
  await page.getByRole('button', { name: 'Save' }).click();
  await page.getByRole('status').filter({ hasText: 'Saved' }).waitFor();
  await page.reload();
  await page.getByRole('heading', { name: 'Status checks' }).waitFor();
  check('custom values are saved and come back', await page.locator('input[name=interval][value=custom]').isChecked() && (await page.getByLabel('Seconds').first().inputValue()) === '10' && (await page.getByLabel('Seconds').last().inputValue()) === '2');

  // --- add apps
  await page.goto(BASE + '/');
  const add = async (name, url, { check: on = false, address = '', strict = false } = {}) => {
    await page.getByRole('button', { name: 'Add app' }).first().click();
    await page.getByRole('dialog').getByRole('button', { name: /Custom app/ }).click();
    const d = page.getByRole('dialog');
    await d.getByLabel('Name', { exact: true }).fill(name);
    await d.getByLabel('Address', { exact: true }).fill(url);
    if (on) {
      await d.getByRole('switch', { name: /Check if this app is online/ }).click();
      if (address) await d.getByLabel('Address to check (optional)').fill(address);
      if (strict) await d.getByLabel(/It answers with success/).check();
    }
    return d;
  };

  let d = await add('Plain', 'plain.example.com');
  check('the status check is off by default', (await d.getByRole('switch', { name: /Check if this app is online/ }).getAttribute('aria-checked')) === 'false');
  check('its options are hidden until it is switched on', !(await d.getByLabel('Address to check (optional)').count()));
  await d.getByRole('button', { name: 'Add app' }).click();
  await d.waitFor({ state: 'detached' });

  d = await add('Reachable', 'reachable.example.com', { check: true, address: '127.0.0.1:19181' });
  await d.getByRole('button', { name: 'Test now' }).click();
  await d.getByRole('status').filter({ hasText: 'Online' }).waitFor();
  check('Test now says online, with the time it took', /answered in \d+ ms \(HTTP 200\)/.test(await d.getByRole('status').textContent()));
  await d.getByRole('button', { name: 'Add app' }).click();
  await d.waitFor({ state: 'detached' });

  d = await add('Fussy', 'fussy.example.com', { check: true, address: '127.0.0.1:19182', strict: true });
  await d.getByRole('button', { name: 'Test now' }).click();
  await d.getByRole('status').filter({ hasText: 'Offline' }).waitFor();
  check('Test now says why: HTTP 500 is no success', (await d.getByRole('status').textContent()).includes('Answered with HTTP 500'));
  await d.getByRole('button', { name: 'Add app' }).click();
  await d.waitFor({ state: 'detached' });

  d = await add('Lenient', 'lenient.example.com', { check: true, address: '127.0.0.1:19182' });
  await d.getByRole('button', { name: 'Add app' }).click();
  await d.waitFor({ state: 'detached' });

  d = await add('Bad', 'bad.example.com', { check: true, address: 'javascript:alert(1)' });
  await d.getByRole('button', { name: 'Add app' }).click();
  await d.getByRole('alert').filter({ hasText: 'address to check' }).waitFor();
  check('an unsafe address to check is refused', true);
  await d.getByRole('button', { name: 'Cancel' }).click();

  const tile = (n) => page.locator('.tile', { hasText: n }).first();
  const badge = (n) => tile(n).locator('.status');
  const until = async (fn, ms = 40000) => {
    const end = Date.now() + ms;
    while (Date.now() < end) {
      if (await fn()) return true;
      await page.reload();
      await page.waitForSelector('text=apps');
      await page.waitForTimeout(1500);
    }
    return false;
  };

  check('an app with no check has no badge', await page.locator('.tile', { hasText: 'Plain' }).locator('.status').count() === 0);
  check('a reachable app turns online', await until(async () => (await badge('Reachable').getAttribute('aria-label').catch(() => '')) === 'Online'));
  check('with strict on, a 500 means offline once two looks agree', await until(async () => (await badge('Fussy').getAttribute('aria-label').catch(() => '')) === 'Offline'));
  check('without strict, the same 500 counts as online', (await badge('Lenient').getAttribute('aria-label')) === 'Online');
  check('an offline tile says so in words, not only colour', (await tile('Fussy').textContent()).includes('Offline ·'));
  check('an offline tile has a cross on its badge', (await badge('Fussy').textContent()).includes('×'));

  // hover card
  const card = tile('Fussy').locator('.card');
  await tile('Fussy').locator('.name').hover();
  await page.waitForTimeout(250);
  check('hovering the tile itself does not show the card', !(await card.isVisible()));
  await badge('Fussy').hover();
  await card.waitFor({ state: 'visible' });
  const text = await card.textContent();
  check('the card is drawn above the tiles that follow it', await tile('Fussy').evaluate((el) => getComputedStyle(el).zIndex !== 'auto' && Number(getComputedStyle(el).zIndex) > 0));
  const box = await card.boundingBox();
  check('the card stays inside the page', box.x >= 0 && box.x + box.width <= 1280, JSON.stringify(box));
  check('hovering the badge shows why and what was checked', text.includes('Answered with HTTP 500') && text.includes('http://127.0.0.1:19182'), text);

  // the 24-hour bar
  await card.locator('.bar').nth(23).waitFor();
  check('the card has 24 hourly bars', (await card.locator('.bar').count()) === 24);
  await page.waitForFunction((sel) => !document.querySelector(sel + ' .bar:last-child')?.classList.contains('none'), 'main .card:not([hidden])', { timeout: 8000 }).catch(() => {});
  check('the current hour has data, and failures are filled in red', (await card.locator('.bar').last().evaluate((el) => !el.classList.contains('none') && !!el.querySelector('.fail'))));
  check('hours with no data are empty', (await card.locator('.bar.none').count()) >= 20);
  check('it says how much of the time the app was online', /\d+(\.\d)?% online/.test(await card.locator('.hist-head').textContent()), await card.locator('.hist-head').textContent());
  // the mouse can travel from the badge onto the card and point at a bar
  const lastBar = card.locator('.bar').last();
  const bb = await lastBar.boundingBox();
  const from = await badge('Fussy').boundingBox();
  await page.mouse.move(from.x + from.width / 2, from.y + from.height / 2);
  await page.mouse.move(bb.x + bb.width / 2, bb.y + bb.height / 2, { steps: 8 });
  await page.waitForTimeout(400);
  check('the card stays while the mouse moves onto it', await card.isVisible());
  const tipText = await lastBar.evaluate((el) => getComputedStyle(el, '::after').content);
  check('pointing at a bar says what happened that hour', /checks failed/.test(tipText), tipText);
  await page.mouse.move(5, 5);
  await page.waitForTimeout(500);
  check('the card goes when the mouse leaves', !(await card.isVisible()));

  // keyboard focus shows it too, since a keyboard has no hover
  await page.mouse.move(5, 5);
  await tile('Fussy').locator('a.open').focus();
  await page.waitForTimeout(250);
  check('focusing the tile with the keyboard shows the card', await card.isVisible());
  await tile('Fussy').locator('a.open').blur();

  // the offline filter
  const filter = page.getByRole('button', { name: /^Offline · 1$/ });
  check('an Offline filter appears with the count', await filter.isVisible());
  await filter.click();
  const names = [...new Set(await page.locator('main a.open .name').allTextContents())];
  check('it shows only the offline apps', names.join() === 'Fussy', names.join());
  await filter.click();

  // an app going away and coming back
  await close(good);
  check('when the app stops, it goes offline', await until(async () => (await badge('Reachable').getAttribute('aria-label').catch(() => '')) === 'Offline', 45000));
  const down = await tile('Reachable').locator('.note').textContent();
  check('and says how long', /Offline · (under a minute|\d+ min)/.test(down), down);
  good = await listen(19181);
  check('when it comes back, it is online again', await until(async () => (await badge('Reachable').getAttribute('aria-label').catch(() => '')) === 'Online', 45000));

  // turning the check off removes the badge
  await tile('Reachable').hover();
  await page.getByRole('button', { name: 'Edit Reachable' }).click();
  d = page.getByRole('dialog');
  await d.getByRole('switch', { name: /Check if this app is online/ }).click();
  await d.getByRole('button', { name: 'Save changes' }).click();
  await d.waitFor({ state: 'detached' });
  check('switching the check off removes the badge', await until(async () => (await tile('Reachable').locator('.status').count()) === 0, 20000));

  await close(good); await close(broken);
  await browser.close();
  check('no page errors', errors.length === 0, errors.join(' | '));
  console.log(failed ? `\n${failed} FAILED` : '\nALL PASSED');
  process.exit(failed ? 1 : 0);
})();
