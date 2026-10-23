// The main journeys, start to finish: set up, sign out, sign in, add apps, reorder, persist.
// Usage: BASE=http://127.0.0.1:8080 node flows.js   (against a fresh, empty data directory)
const { chromium } = require('playwright');
const BASE = process.env.BASE || 'http://127.0.0.1:8080';
let failed = 0;
const check = (name, ok, extra = '') => { console.log((ok ? 'PASS ' : 'FAIL ') + name + (ok ? '' : '  ' + extra)); if (!ok) failed++; };

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage({ viewport: { width: 1280, height: 800 } });
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));

  await page.goto(BASE + '/');
  await page.getByLabel('Username').fill('demo');
  await page.getByLabel('Password', { exact: true }).fill('demo-password-1');
  await page.getByLabel('Confirm password').fill('demo-password-1');
  await page.getByRole('button', { name: 'Create account' }).click();
  await page.waitForSelector('text=No apps yet');
  check('first run creates the account and signs in', true);

  await page.getByRole('button', { name: /Sign out/ }).click();
  await page.getByText('Sign in to Bifrost').waitFor();
  await page.getByLabel('Username').fill('demo');
  await page.getByLabel('Password', { exact: true }).fill('wrong-password');
  await page.getByRole('button', { name: 'Sign in' }).click();
  await page.getByRole('alert').waitFor();
  check('a wrong password is refused', true);
  await page.getByLabel('Password', { exact: true }).fill('demo-password-1');
  await page.getByRole('button', { name: 'Sign in' }).click();
  await page.waitForSelector('text=No apps yet');
  check('the right password signs in', true);

  for (const [name, url] of [['Alpha', 'alpha.lan'], ['Beta', 'beta.lan'], ['Gamma', 'gamma.lan']]) {
    await page.getByRole('button', { name: 'Add app' }).last().click();
    await page.getByRole('dialog').getByRole('button', { name: /Custom app/ }).click();
    await page.getByRole('dialog').getByLabel('Name', { exact: true }).fill(name);
    await page.getByRole('dialog').getByLabel(/Address/).fill(url);
    await page.getByRole('dialog').getByRole('button', { name: 'Add app' }).click();
    await page.getByRole('dialog').waitFor({ state: 'detached' });
  }
  const names = () => page.locator('a.open .name').allTextContents();
  check('three apps added, in order', (await names()).join() === 'Alpha,Beta,Gamma', (await names()).join());

  await page.getByRole('button', { name: 'Edit Alpha' }).click({ force: true });
  await page.getByRole('dialog').getByRole('button', { name: 'Move later' }).click();
  await page.getByRole('dialog').getByRole('button', { name: 'Close' }).first().click();
  check('Alpha moves after Beta', (await names()).join() === 'Beta,Alpha,Gamma', (await names()).join());

  await page.reload();
  await page.waitForSelector('text=3 apps');
  check('the order survives a reload', (await names()).join() === 'Beta,Alpha,Gamma', (await names()).join());

  await page.keyboard.press('/');
  await page.keyboard.type('gam');
  check('search narrows to Gamma', (await names()).join() === 'Gamma', (await names()).join());

  await browser.close();
  check('no page errors', errors.length === 0, errors.join(' | '));
  console.log(failed ? `\n${failed} FAILED` : '\nALL PASSED');
  process.exit(failed ? 1 : 0);
})();
