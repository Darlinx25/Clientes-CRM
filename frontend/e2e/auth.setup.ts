import { test as setup, expect } from '@playwright/test';
import { TEST_USER } from './global-setup';

export const STORAGE_STATE = 'playwright/.auth/user.json';

setup('authenticate', async ({ page }) => {
  await page.goto('/');

  await page.getByLabel(/username or email/i).fill(TEST_USER.username);
  await page.getByLabel(/password/i).fill(TEST_USER.password);
  await page.getByRole('button', { name: /login/i }).click();

  // Landing on the contacts list confirms both the cookie and cached user info are set.
  await expect(page).toHaveURL(/\/contacts/);
  await expect(page.getByRole('heading', { name: /contacts/i })).toBeVisible({
    timeout: 15000,
  });

  await page.context().storageState({ path: STORAGE_STATE });
});
