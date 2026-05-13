import { FullConfig } from '@playwright/test';

async function globalTeardown(config: FullConfig) {
  if (process.env.CI) {
    console.log('CI detected: Docker Compose cleanup handled by GitHub Actions');
    return;
  }

  const { execSync } = await import('child_process');

  try {
    console.log('Stopping Docker Compose...');
    execSync('docker compose down', { stdio: 'inherit', cwd: process.cwd() + '/..' });
    console.log('Docker Compose stopped');
  } catch (err) {
    console.error('Failed to stop Docker Compose:', err);
  }
}

export default globalTeardown;
