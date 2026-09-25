"""Installer regression tests. No daemon, downloads, or production credentials needed."""
import os
import json
import pty
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]
REAL_DOCKER = shutil.which('docker')


class InstallerTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name) / 'checkout'
        self.root.mkdir()
        shutil.copy(ROOT / 'install.sh', self.root)
        shutil.copytree(ROOT / 'scripts', self.root / 'scripts')
        shutil.copytree(ROOT / 'deploy', self.root / 'deploy',
                        ignore=shutil.ignore_patterns('.env'))
        self.env_file = self.root / 'deploy/docker/.env'
        bin_dir = Path(self.tmp.name) / 'bin'
        bin_dir.mkdir()
        docker = bin_dir / 'docker'
        docker.write_text('''#!/bin/bash
printf '%s\\n' "$*" >> "$DOCKER_LOG"
if [[ "$*" == info && "${FAIL_DOCKER:-}" == 1 ]]; then exit 1; fi
if [[ "$*" == 'compose ps --format json '* ]]; then
  printf '{"Health":"%s","State":"running"}\\n' "${TEST_HEALTH:-healthy}"
fi
if [[ "$*" == 'compose exec '* ]]; then cat >/dev/null; fi
''')
        docker.chmod(0o755)
        self.env = {**os.environ, 'PATH': str(bin_dir) + ':' + os.environ['PATH'],
                    'DOCKER_LOG': str(Path(self.tmp.name) / 'docker.log'),
                    'SSH_CONNECTION': '192.0.2.1 1234 169.58.33.188 22'}

    def run_install(self, **extra):
        return subprocess.run(['sh', str(self.root / 'deploy/docker/install.sh')],
                              cwd=self.tmp.name, env={**self.env, **extra}, input='',
                              text=True, capture_output=True, timeout=20)

    def test_clean_install_and_retry_preserve_secrets(self):
        result = self.run_install()
        self.assertEqual(result.returncode, 0, result.stderr)
        initial = self.env_file.read_text()
        self.assertRegex(initial, r"SUPERSET_SECRET_KEY='[a-f0-9]{64}'")
        self.assertRegex(initial, r"SUPERSET_ADMIN_PASSWORD='[a-f0-9]{32}'")
        self.assertRegex(initial, r"SENSORFLOW_INGESTION_TOKEN='[a-f0-9]{48}'")
        self.assertEqual(self.env_file.stat().st_mode & 0o777, 0o600)
        self.assertIn('http://169.58.33.188:8088/superset/dashboard/', result.stdout)
        again = self.run_install()
        self.assertEqual(again.returncode, 0, again.stderr)
        self.assertEqual(initial, self.env_file.read_text())
        log = Path(self.env['DOCKER_LOG']).read_text()
        self.assertNotIn('up -d ingestion', log)
        self.assertIn('compose ps --format json superset', log)

    def test_template_missing_secrets_and_custom_settings(self):
        shutil.copy(ROOT / 'deploy/docker/.env.example', self.env_file)
        with self.env_file.open('a') as f:
            f.write("\nSUPERSET_PORT=18088\nCUSTOM='$(touch SHOULD_NOT_EXIST)'\n")
        result = self.run_install()
        self.assertEqual(result.returncode, 0, result.stderr)
        content = self.env_file.read_text()
        self.assertIn('SUPERSET_PORT=18088', content)
        self.assertIn('http://169.58.33.188:18088', result.stdout)
        self.assertEqual(self.run_install().returncode, 0)
        self.assertEqual(content, self.env_file.read_text())
        self.assertIn("CUSTOM='$(touch SHOULD_NOT_EXIST)'", content)
        self.assertFalse((Path(self.tmp.name) / 'SHOULD_NOT_EXIST').exists())
        self.assertRegex(content, r"SENSORFLOW_INGESTION_TOKEN='[a-f0-9]{48}'")
        self.assertIn('CLICKHOUSE_PASSWORD=', content)

    def test_unavailable_daemon_does_not_write_config(self):
        result = self.run_install(FAIL_DOCKER='1')
        self.assertNotEqual(result.returncode, 0)
        self.assertFalse(self.env_file.exists())
        self.assertIn('无法连接 Docker', result.stderr)

    def test_unhealthy_service_is_not_reported_as_success(self):
        result = self.run_install(TEST_HEALTH='unhealthy')
        self.assertNotEqual(result.returncode, 0)
        self.assertNotIn('演示环境已启动', result.stdout)
        self.assertTrue(self.env_file.exists())
        initial = self.env_file.read_text()
        retry = self.run_install()
        self.assertEqual(retry.returncode, 0, retry.stderr)
        self.assertEqual(initial, self.env_file.read_text())

    def test_terminal_credentials_are_red(self):
        master, slave = pty.openpty()
        try:
            env = dict(self.env)
            env.pop('NO_COLOR', None)
            process = subprocess.Popen(['sh', str(self.root / 'install.sh')],
                                       env=env, stdin=subprocess.DEVNULL, stdout=slave,
                                       stderr=subprocess.PIPE)
            os.close(slave)
            slave = None
            output = b''
            while True:
                try:
                    chunk = os.read(master, 65536)
                    if not chunk:
                        break
                    output += chunk
                except OSError:
                    break
            stderr = process.communicate(timeout=20)[1]
            self.assertEqual(process.returncode, 0, stderr)
            self.assertIn(b'\x1b[1;31m', output)
            self.assertIn(b'\x1b[0m', output)
        finally:
            os.close(master)
            if slave is not None:
                os.close(slave)

    def test_ipv6_public_url(self):
        result = self.run_install(SSH_CONNECTION='2001:db8::1 1234 2001:db8::2 22')
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn('http://[2001:db8::2]:8088', result.stdout)

    @unittest.skipUnless(REAL_DOCKER, 'Docker CLI unavailable')
    def test_generated_config_with_real_compose(self):
        result = self.run_install()
        self.assertEqual(result.returncode, 0, result.stderr)
        result = subprocess.run([REAL_DOCKER, 'compose', '--profile', 'edge', '--profile', 'redis-bundled',
                                 '--profile', 'clickhouse-bundled', 'config', '--format', 'json'],
                                cwd=self.env_file.parent, text=True, capture_output=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        services = json.loads(result.stdout)['services']
        self.assertEqual(services['superset']['ports'][0]['host_ip'], '0.0.0.0')
        self.assertEqual(services['redis']['ports'][0]['host_ip'], '127.0.0.1')
        self.assertIn('superset', services['caddy']['depends_on'])
        self.assertNotIn('ingestion', services)


if __name__ == '__main__':
    unittest.main()
