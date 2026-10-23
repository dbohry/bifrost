// Runs axe-core over every screen in every bundled theme, on a phone and a desktop.
// Usage: BASE=http://127.0.0.1:8080 node a11y.js   (against a fresh, empty data directory)
const { chromium } = require('playwright');
const axeSource = require('fs').readFileSync(require.resolve('axe-core/axe.min.js'), 'utf8');

const BASE = process.env.BASE || 'http://127.0.0.1:8080';
const themes = ['bifrost-light', 'solarized-light', 'bifrost-dark', 'dracula', 'nord', 'gruvbox', 'solarized-dark'];
const screens = [
  { name: 'phone', viewport: { width: 390, height: 844 }, touch: true },
  { name: 'desktop', viewport: { width: 1280, height: 800 }, touch: false },
];
let failed = 0;

async function audit(page, label) {
  await page.evaluate(axeSource);
  const result = await page.evaluate(() => axe.run(document, { runOnly: ['wcag2a', 'wcag2aa', 'wcag21aa', 'best-practice'] }));
  const bad = result.violations;
  console.log((bad.length ? 'FAIL ' : 'PASS ') + label);
  for (const v of bad) {
    failed++;
    console.log(`     ${v.id} (${v.impact}): ${v.help}`);
    for (const n of v.nodes.slice(0, 3)) console.log(`       ${n.target.join(' ')}  ${n.failureSummary.split('\n')[1] || ''}`);
  }
}

(async () => {
  const browser = await chromium.launch();
  const ctx0 = await browser.newContext();
  const p0 = await ctx0.newPage();
  await p0.goto(BASE + '/');
  await audit(p0, 'setup screen');
  await p0.getByLabel('Username').fill('demo');
  await p0.getByLabel('Password', { exact: true }).fill('demo-password-1');
  await p0.getByLabel('Confirm password').fill('demo-password-1');
  await p0.getByRole('button', { name: 'Create account' }).click();
  await p0.waitForSelector('text=No apps yet');
  const post = (body) => p0.evaluate((b) => fetch('/api/apps', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(b) }), body);
  await post({ name: 'Jellyfin', url: 'http://jellyfin.lan', category: 'Media', favourite: true, icon: 'jellyfin' });
  await post({ name: 'Sonarr', url: 'http://sonarr.lan', category: 'Media', favourite: false, icon: '' });
  await post({ name: 'Pi-hole', url: 'http://pihole.lan', category: 'Network', favourite: false, icon: 'pi-hole' });
  const cookies = await ctx0.cookies();
  await ctx0.close();

  // sign-in screen, signed out
  const out = await browser.newContext();
  const po = await out.newPage();
  await po.goto(BASE + '/');
  await audit(po, 'sign-in screen');
  await out.close();

  for (const sc of screens) {
    for (const theme of themes) {
      const dark = ['bifrost-dark', 'dracula', 'nord', 'gruvbox', 'solarized-dark'].includes(theme);
      const ctx = await browser.newContext({ viewport: sc.viewport, hasTouch: sc.touch, isMobile: sc.touch });
      await ctx.addCookies(cookies);
      const page = await ctx.newPage();
      await page.goto(BASE + '/');
      await page.evaluate((t) => fetch('/api/preferences', { method: 'PUT', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ mode: t.dark ? 'dark' : 'light', lightTheme: t.dark ? 'bifrost-light' : t.theme, darkTheme: t.dark ? t.theme : 'bifrost-dark', logo: 'app' }) }), { theme, dark });
      await page.goto(BASE + '/');
      await page.waitForSelector('text=3 apps');
      const tag = `${sc.name}, ${theme}`;
      await audit(page, `${tag}: dashboard`);
      if (theme === 'bifrost-light' || theme === 'dracula') {
        await page.goto(BASE + '/settings/appearance'); await page.waitForSelector('main h1');
        await audit(page, `${tag}: appearance`);
        await page.goto(BASE + '/settings/status'); await page.waitForSelector('text=Check every');
        await audit(page, `${tag}: status checks`);
        await page.goto(BASE + '/settings/users'); await page.waitForSelector('text=demo (you)');
        await audit(page, `${tag}: users`);
        await page.goto(BASE + '/settings'); await page.waitForSelector('main h1');
        await audit(page, `${tag}: account`);
        await page.goto(BASE + '/');
        await page.waitForSelector('text=3 apps');
        await page.getByRole('button', { name: 'Add app' }).last().click();
        await page.getByRole('dialog').getByText('Step 1 of 2').waitFor();
        await page.waitForTimeout(300);
        await audit(page, `${tag}: add app, step 1`);
        await page.getByRole('dialog').getByRole('button', { name: /Custom app/ }).click();
        await page.waitForTimeout(300);
        await audit(page, `${tag}: add app, step 2`);
        await page.getByRole('dialog').getByRole('switch', { name: /Check if this app is online/ }).click();
        await page.waitForTimeout(200);
        await audit(page, `${tag}: add app, status check options`);
      }
      await ctx.close();
    }
  }
  await browser.close();
  console.log(failed ? `\n${failed} violations` : '\nNo violations');
  process.exit(failed ? 1 : 0);
})();
