import { FullConfig } from '@playwright/test';
import * as path from 'path';

async function globalSetup(config: FullConfig) {
  if (process.env.CI) {
    console.log('CI detected: Docker Compose should already be running');
    return;
  }

  const { execSync } = await import('child_process');
  const projectRoot = path.resolve(process.cwd(), '..');

  try {
    console.log('Cleaning up previous Docker state...');
    execSync('docker compose down -v', { stdio: 'inherit', cwd: projectRoot });

    console.log('Starting Docker Compose...');
    execSync('docker compose up -d', { stdio: 'inherit', cwd: projectRoot });

    console.log('Waiting for app to be healthy...');
    for (let i = 0; i < 90; i++) {
      try {
        const res = await fetch('http://localhost:8080/health');
        if (res.ok) {
          console.log('App is healthy after ~' + (i * 2) + 's');
          return;
        }
      } catch {
        // Not ready yet
      }
      await new Promise(r => setTimeout(r, 2000));
    }
    throw new Error('App did not become healthy within 3 minutes. Check docker logs: docker compose logs app');
  } catch (err) {
    console.error('Failed to start Docker Compose:', err);
    throw err;
  }
}

export default globalSetup;
