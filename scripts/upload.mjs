import { Client } from 'ssh2';
import { readFileSync } from 'fs';
import { join, basename } from 'path';
import { readdirSync, statSync, existsSync } from 'fs';
import { createInterface } from 'readline';

// 凭据从环境变量读取，绝不硬编码（历史中泄露过的密码必须已在服务器上更换）
const HOST = process.env.DEPLOY_HOST;
const PORT = Number(process.env.DEPLOY_PORT || 22);
const USERNAME = process.env.DEPLOY_USER;
const PASSWORD = process.env.DEPLOY_PASSWORD;
const KEY_PATH = process.env.DEPLOY_KEY_PATH || '';
const REMOTE_PATH = process.env.DEPLOY_PATH || '/opt/blog';

if (!HOST || !USERNAME || (!PASSWORD && !KEY_PATH)) {
  console.error('缺少部署凭据。用法示例：');
  console.error('  DEPLOY_HOST=1.2.3.4 DEPLOY_USER=root DEPLOY_KEY_PATH=~/.ssh/id_rsa node scripts/upload.mjs');
  console.error('  DEPLOY_HOST=1.2.3.4 DEPLOY_USER=root DEPLOY_PASSWORD=xxx node scripts/upload.mjs');
  process.exit(1);
}

// Recursive directory walk
function* walkDir(dir, base = '') {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const rel = join(base, entry.name);
    const full = join(dir, entry.name);
    if (entry.name === '.git' || entry.name === 'node_modules' || entry.name === 'dist' ||
        entry.name.endsWith('.exe') || entry.name.endsWith('.exe~') || entry.name.endsWith('.db') ||
        entry.name.endsWith('.db-shm') || entry.name.endsWith('.db-wal') || entry.name === 'upload.mjs') continue;
    if (entry.isDirectory()) {
      yield* walkDir(full, rel);
    } else {
      yield { localPath: full, remoteRel: rel };
    }
  }
}

function execCommand(conn, cmd) {
  return new Promise((resolve, reject) => {
    conn.exec(cmd, (err, stream) => {
      if (err) return reject(err);
      let out = '', errOut = '';
      stream.on('data', d => { out += d.toString(); process.stdout.write(d); });
      stream.stderr.on('data', d => { errOut += d.toString(); process.stderr.write(d); });
      stream.on('close', code => {
        if (code !== 0) reject(new Error(`Exit ${code}: ${errOut}`));
        else resolve(out);
      });
    });
  });
}

const files = [...walkDir('.')];
console.log(`Found ${files.length} files to upload`);

const conn = new Client();
conn.on('ready', async () => {
  console.log('SSH connected');

  // Create remote directory
  await execCommand(conn, `mkdir -p ${REMOTE_PATH}`);

  // Use SFTP for file upload
  conn.sftp(async (err, sftp) => {
    if (err) { console.error('SFTP error:', err); conn.end(); return; }

    // Ensure remote directory structure
    const dirs = new Set();
    for (const f of files) {
      const parts = f.remoteRel.replace(/\\/g, '/').split('/');
      for (let i = 1; i < parts.length; i++) {
        dirs.add(parts.slice(0, i).join('/'));
      }
    }

    for (const dir of [...dirs].sort()) {
      await new Promise((res, rej) => {
        sftp.mkdir(join(REMOTE_PATH, dir), { mode: 0o755 }, err => {
          if (err && err.code !== 4) console.log(`mkdir ${dir}: ${err.message}`);
          res();
        });
      });
    }

    // Upload files
    let uploaded = 0;
    for (const f of files) {
      const remoteFile = join(REMOTE_PATH, f.remoteRel).replace(/\\/g, '/');
      await new Promise((res, rej) => {
        sftp.fastPut(f.localPath, remoteFile, err => {
          if (err) { console.error(`Upload ${f.remoteRel} failed:`, err.message); rej(err); }
          else {
            uploaded++;
            process.stdout.write(`\rUploaded: ${uploaded}/${files.length} - ${f.remoteRel.padEnd(50)}`);
            res();
          }
        });
      });
    }

    console.log(`\n\nAll ${uploaded} files uploaded!`);

    // Now build and deploy
    console.log('\n=== Building and deploying ===');
    await execCommand(conn, `cd ${REMOTE_PATH} && docker compose down 2>/dev/null; docker compose up -d --build`);
    await execCommand(conn, `cd ${REMOTE_PATH} && docker compose ps`);

    console.log('\n=== Done! ===');
    sftp.end();
    conn.end();
  });
});

conn.on('error', err => { console.error('Connection error:', err); process.exit(1); });

conn.connect(KEY_PATH
  ? { host: HOST, port: PORT, username: USERNAME, privateKey: readFileSync(KEY_PATH) }
  : { host: HOST, port: PORT, username: USERNAME, password: PASSWORD });
