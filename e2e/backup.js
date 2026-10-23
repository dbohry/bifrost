// Export and import, through the Settings > Backup screen.
// Usage: BASE=http://127.0.0.1:8080 node backup.js   (against a fresh, empty data directory)
const { chromium } = require('playwright');
const fs = require('fs');
const os = require('os');
const path = require('path');
const BASE = process.env.BASE || 'http://127.0.0.1:8080';
let failed = 0;
const check = (name, ok, extra = '') => { console.log((ok ? 'PASS ' : 'FAIL ') + name + (ok ? '' : '  ' + extra)); if (!ok) failed++; };

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage({ viewport: { width: 1280, height: 800 } });
  const errors = [];
  page.on('pageerror', (e) => errors.push(e.message));
  const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'bifrost-backup-'));

  await page.goto(BASE + '/');
  await page.getByLabel('Username').fill('demo');
  await page.getByLabel('Password', { exact: true }).fill('demo-password-1');
  await page.getByLabel('Confirm password').fill('demo-password-1');
  await page.getByRole('button', { name: 'Create account' }).click();
  await page.waitForSelector('text=No apps yet');
  const post = (body) => page.evaluate((b) => fetch('/api/apps', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(b) }), body);
  await post({ name: 'Jellyfin', url: 'jf.lan', category: 'Media', favourite: true, icon: 'jellyfin' });
  await post({ name: 'Wiki', url: 'wiki.lan', category: 'Docs' });
  await page.getByRole('link', { name: 'Settings' }).first().click();
  await page.getByRole('link', { name: 'Backup' }).click();
  await page.getByRole('heading', { name: 'Backup' }).waitFor();

  const [download] = await Promise.all([page.waitForEvent('download'), page.getByRole('link', { name: 'Download export' }).click()]);
  const file = path.join(tmp, download.suggestedFilename());
  await download.saveAs(file);
  const exported = JSON.parse(fs.readFileSync(file, 'utf8'));
  check('the export downloads as bifrost-export.json', download.suggestedFilename() === 'bifrost-export.json');
  check('it holds both apps in order', exported.apps.map((a) => a.name).join() === 'Jellyfin,Wiki', JSON.stringify(exported.apps));

  // lose an app, then merge the file back
  await page.evaluate(async () => {
    const { apps } = await (await fetch('/api/apps')).json();
    await fetch('/api/apps/' + apps[1].id, { method: 'DELETE' });
  });
  await page.setInputFiles('input[type=file]', file);
  await page.getByText('holds 2 apps').waitFor();
  check('choosing a file says what is in it', true);
  await page.getByRole('button', { name: 'Import', exact: true }).click();
  await page.getByRole('status').filter({ hasText: 'Imported' }).waitFor();
  const msg = await page.getByRole('status').filter({ hasText: 'Imported' }).textContent();
  check('merge adds the missing app and skips the one already there', /Imported 1 app, skipped 1 already there/.test(msg), msg);
  await page.goto(BASE + '/');
  await page.waitForSelector('text=2 apps');
  check('both apps are back', [...new Set(await page.locator('a.open .name').allTextContents())].sort().join() === 'Jellyfin,Wiki');

  // replace needs a confirmation
  await page.goto(BASE + '/settings/backup');
  await page.setInputFiles('input[type=file]', file);
  await page.getByText('holds 2 apps').waitFor();
  await page.getByLabel(/Replace my dashboard/).check();
  await page.getByRole('button', { name: 'Import', exact: true }).click();
  const dialog = page.getByRole('dialog');
  await dialog.waitFor();
  check('replace asks first', (await dialog.textContent()).includes('cannot be undone'));
  await dialog.getByRole('button', { name: 'Cancel' }).click();
  check('cancelling changes nothing', !(await page.getByRole('status').filter({ hasText: 'Imported' }).count()));
  await page.getByRole('button', { name: 'Import', exact: true }).click();
  await dialog.getByRole('button', { name: 'Replace my apps' }).click();
  await page.getByRole('status').filter({ hasText: 'Imported' }).waitFor();
  check('replace removes the old apps and adds the file\'s', /Imported 2 apps, removed 2 apps/.test(await page.getByRole('status').filter({ hasText: 'Imported' }).textContent()));

  // wrong files
  const bad = path.join(tmp, 'notes.json');
  fs.writeFileSync(bad, JSON.stringify({ hello: 'world' }));
  await page.setInputFiles('input[type=file]', bad);
  await page.getByRole('alert').filter({ hasText: 'not a Bifrost export' }).waitFor();
  check('a file that is not an export is refused with a message', true);
  const evil = path.join(tmp, 'evil.json');
  fs.writeFileSync(evil, JSON.stringify({ format: 'bifrost-export', version: 1, apps: [{ name: 'Evil', url: 'javascript:alert(1)' }] }));
  await page.setInputFiles('input[type=file]', evil);
  await page.getByText('holds 1 app').waitFor();
  await page.getByLabel(/Add to my dashboard/).check();
  await page.getByRole('button', { name: 'Import', exact: true }).click();
  await page.getByRole('alert').filter({ hasText: 'App 1' }).waitFor();
  check('an unsafe address is refused by the server, naming the app', true);

  await browser.close();
  check('no page errors', errors.length === 0, errors.join(' | '));
  console.log(failed ? `\n${failed} FAILED` : '\nALL PASSED');
  process.exit(failed ? 1 : 0);
})();
