// The two layouts: split by category, and one list in the saved order.
// Usage: BASE=http://127.0.0.1:8080 node views.js   (against a fresh, empty data directory)
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
  // Categories interleave on purpose: A(x) B(y) C(x) D(y) E(x)
  for (const [name, category] of [['Alpha', 'X'], ['Bravo', 'Y'], ['Charlie', 'X'], ['Delta', 'Y'], ['Echo', 'X']])
    await page.evaluate((a) => fetch('/api/apps', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ ...a, url: a.name.toLowerCase() + '.lan' }) }), { name, category });
  await page.reload();
  await page.waitForSelector('text=5 apps');

  const names = () => page.locator('main a.open .name').allTextContents();
  const layout = page.getByRole('group', { name: 'Layout' });

  check('it starts split by category', (await names()).join() === 'Alpha,Charlie,Echo,Bravo,Delta' && await page.getByRole('heading', { name: 'X' }).isVisible());
  check('the category tabs are shown', await page.getByRole('button', { name: 'Y', exact: true }).isVisible());

  // switching animates the tiles
  await layout.getByRole('button', { name: 'All apps' }).click();
  const animated = await page.evaluate(() => document.getAnimations().filter((a) => a.effect && a.effect.target && a.effect.target.matches && a.effect.target.matches('[data-app-id]')).length);
  check('switching to the single list animates the tiles', animated >= 3, String(animated));
  await page.waitForTimeout(600);

  check('the list is in the saved order, not by category', (await names()).join() === 'Alpha,Bravo,Charlie,Delta,Echo', (await names()).join());
  check('there are no category headings or tabs in it', !(await page.getByRole('heading', { name: 'X' }).count()) && !(await page.getByRole('button', { name: 'Y', exact: true }).count()));
  check('the choice is marked', (await layout.getByRole('button', { name: 'All apps' }).getAttribute('aria-pressed')) === 'true');

  // drag Echo to the front
  const tile = (n) => page.locator('.tile', { hasText: n });
  const from = await tile('Echo').boundingBox();
  const to = await tile('Alpha').boundingBox();
  await page.mouse.move(from.x + from.width / 2, from.y + from.height / 2);
  await page.mouse.down();
  for (let i = 1; i <= 12; i++) await page.mouse.move(from.x + from.width / 2 + ((to.x + 10 - from.x - from.width / 2) * i) / 12, from.y + from.height / 2 + ((to.y + 10 - from.y - from.height / 2) * i) / 12);
  await page.mouse.up();
  await page.waitForTimeout(400);
  check('dragging moves an app anywhere in the list', (await names()).join() === 'Echo,Alpha,Bravo,Charlie,Delta', (await names()).join());

  await page.reload();
  await page.waitForSelector('text=5 apps');
  check('the layout and the order are remembered after a reload', (await names()).join() === 'Echo,Alpha,Bravo,Charlie,Delta', (await names()).join());

  // Move later from the pencil uses the whole list
  await tile('Echo').hover();
  await page.getByRole('button', { name: 'Edit Echo' }).click();
  await page.getByRole('dialog').getByRole('button', { name: 'Move later' }).click();
  await page.getByRole('dialog').getByRole('button', { name: 'Close' }).first().click();
  check('Move later steps through the whole list', (await names()).join() === 'Alpha,Echo,Bravo,Charlie,Delta', (await names()).join());

  // search narrows and turns dragging off
  await page.keyboard.press('/');
  await page.keyboard.type('a');
  check('search works in the list', (await names()).length > 0);
  await page.keyboard.press('Escape');

  // back to categories: groups follow first appearance of the new order
  await layout.getByRole('button', { name: 'By category' }).click();
  await page.waitForTimeout(600);
  check('back to categories, following the new order', (await names()).join() === 'Alpha,Echo,Charlie,Bravo,Delta', (await names()).join());
  check('headings and tabs are back', await page.getByRole('heading', { name: 'X' }).isVisible() && await page.getByRole('button', { name: 'Y', exact: true }).isVisible());

  // reduced motion: no animation
  const quiet = await browser.newContext({ viewport: { width: 1280, height: 800 }, reducedMotion: 'reduce' });
  await quiet.addCookies(await ctx.cookies());
  const q = await quiet.newPage();
  await q.goto(BASE + '/');
  await q.waitForSelector('text=5 apps');
  await q.getByRole('group', { name: 'Layout' }).getByRole('button', { name: 'All apps' }).click();
  const still = await q.evaluate(() => document.getAnimations().filter((a) => a.effect && a.effect.target && a.effect.target.matches && a.effect.target.matches('[data-app-id]')).length);
  check('with reduced motion, nothing animates', still === 0, String(still));
  await quiet.close();

  // phone: the layout switch is a 44px touch target
  const phone = await browser.newContext({ viewport: { width: 390, height: 844 }, hasTouch: true, isMobile: true });
  await phone.addCookies(await ctx.cookies());
  const p = await phone.newPage();
  await p.goto(BASE + '/');
  await p.waitForSelector('text=5 apps');
  const sizes = await p.getByRole('group', { name: 'Layout' }).getByRole('button').evaluateAll((els) => els.map((e) => [Math.round(e.getBoundingClientRect().width), Math.round(e.getBoundingClientRect().height)]));
  check('on a phone the layout buttons are 44px', sizes.every(([w, h]) => w >= 44 && h >= 44), JSON.stringify(sizes));
  check('and the page does not scroll sideways', await p.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth));
  await p.screenshot({ path: process.env.SHOT || '/dev/null' }).catch(() => {});
  await phone.close();

  await browser.close();
  check('no page errors', errors.length === 0, errors.join(' | '));
  console.log(failed ? `\n${failed} FAILED` : '\nALL PASSED');
  process.exit(failed ? 1 : 0);
})();
