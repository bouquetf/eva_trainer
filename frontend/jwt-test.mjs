import pg from 'pg';
import crypto from 'crypto';
import http from 'http';

const pool = new pg.Pool({
  host: 'localhost',
  port: 5432,
  database: 'eva_trainer',
  user: 'eva_user',
  password: 'eva_pass',
});

const res = await pool.query(
  `INSERT INTO users (discord_id, display_name, is_admin) VALUES ($1, $2, $3) ON CONFLICT (discord_id) DO UPDATE SET display_name = $2 RETURNING id`,
  ['jwt-test-' + Date.now(), 'JWT Test User', false]
);
const userId = res.rows[0].id;
console.log('Created user with ID:', userId);

const jwtSecret = 'test-secret-key-for-manual-validation-32chars';
const now = Math.floor(Date.now() / 1000);
const header = { alg: 'HS256', typ: 'JWT' };
const payload = { sub: userId, exp: now + 3600, iat: now };

const b64url = (obj) =>
  Buffer.from(JSON.stringify(obj)).toString('base64url').replace(/=+$/, '');

const b64header = b64url(header);
const b64payload = b64url(payload);
const signature = crypto
  .createHmac('sha256', jwtSecret)
  .update(b64header + '.' + b64payload)
  .digest('base64url')
  .replace(/=+$/, '');

const token = b64header + '.' + b64payload + '.' + signature;
console.log('Token:', token.substring(0, 50) + '...');

await new Promise((resolve, reject) => {
  const options = {
    hostname: 'localhost',
    port: 8080,
    path: '/api/v1/auth/me',
    method: 'GET',
    headers: { 'Authorization': 'Bearer ' + token }
  };
  const req = http.request(options, (res) => {
    let data = '';
    res.on('data', (chunk) => data += chunk);
    res.on('end', () => {
      console.log('Status:', res.statusCode);
      console.log('Response:', data.substring(0, 500));
      resolve();
    });
  });
  req.on('error', (e) => { console.error('Error:', e.message); reject(e); });
  req.end();
});

await pool.end();
