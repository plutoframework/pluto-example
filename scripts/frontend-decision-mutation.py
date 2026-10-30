"""Bounded frontend lifecycle mutations, compiled and tested in disposable copies."""
from pathlib import Path
import shutil
import subprocess
import tempfile

root = Path(__file__).resolve().parents[1]
mutations = [
    ('response correlation', 'src/features/purchases/flows/decision.ts',
     'event.type === "resolved" && event.requestId === current.requestId',
     'event.type === "resolved"'),
    ('retry key retention', 'src/features/purchases/flows/decision.ts',
     'keys.set(effect.verdict, key);', '/* Retention removed by mutation. */'),
    ('scope cleanup', 'src/features/purchases/mutations.ts',
     'return () => flow.dispose();', 'return () => {};'),
]
for name, filename, before, after in mutations:
    with tempfile.TemporaryDirectory(prefix='pluto-decision-mutant-') as temporary:
        checkout = Path(temporary) / 'web'
        shutil.copy2(root / 'pluto.toml', Path(temporary) / 'pluto.toml')
        shutil.copytree(root / 'web', checkout, ignore=shutil.ignore_patterns('node_modules', 'dist', '.cert', 'test-results', 'playwright-report'))
        (checkout / 'node_modules').symlink_to(root / 'web/node_modules', target_is_directory=True)
        target = checkout / filename
        source = target.read_text()
        assert source.count(before) == 1, f'Mutation anchor drift: {name}'
        target.write_text(source.replace(before, after, 1))
        build = subprocess.run(['npm', 'exec', '--', 'tsc', '--noEmit'], cwd=checkout, capture_output=True, text=True, timeout=60)
        if build.returncode:
            raise SystemExit(f'INVALID: {name}\n{build.stdout}{build.stderr}')
        tests = subprocess.run(['npm', 'exec', '--', 'vitest', 'run', 'src/features/purchases/flows/decision.test.ts', 'src/features/purchases/mutations.test.tsx'], cwd=checkout, capture_output=True, text=True, timeout=30)
        if not tests.returncode or 'FAIL' not in tests.stdout + tests.stderr:
            raise SystemExit(f'NOT KILLED: {name}\n{tests.stdout}{tests.stderr}')
        print(f'KILLED: {name}', flush=True)
print('PASS: three mutations killed; zero invalid builds, survivors or timeouts')
