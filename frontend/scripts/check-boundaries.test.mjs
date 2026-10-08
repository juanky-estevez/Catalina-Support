import assert from 'node:assert/strict';
import { mkdtemp, mkdir, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import test from 'node:test';

import { checkBoundaries, formatViolation } from './check-boundaries.mjs';

async function fixture(files) {
  const root = await mkdtemp(join(tmpdir(), 'catalina-boundaries-'));
  for (const [path, content] of Object.entries(files)) {
    const destination = join(root, path);
    await mkdir(dirname(destination), { recursive: true });
    await writeFile(destination, content);
  }
  return root;
}

test('acepta la dirección de capas y el prefijo propio', async (context) => {
  const root = await fixture({
    'app.routes.ts': "import('./modules/tickets/page');",
    'core/service.ts': "import { shared } from '../shared/value';",
    'shared/value.ts': 'export const shared = true;',
    'modules/tickets/page.ts':
      "import { shared } from '../../shared/value';\nconst url = `/api/tickets/${shared}`;",
  });
  context.after(() => rm(root, { recursive: true, force: true }));
  assert.deepEqual(await checkBoundaries({ appRoot: root }), []);
});

test('informa todas las importaciones que cruzan módulos o invierten capas', async (context) => {
  const root = await fixture({
    'modules/tickets/page.ts': "import { user } from '../users/user';",
    'modules/users/user.ts': 'export const user = true;',
    'core/service.ts': "export { ticket } from '../modules/tickets/ticket';",
    'modules/tickets/ticket.ts': 'export const ticket = true;',
    'shared/value.spec.ts': "import { core } from '../core/core';",
    'core/core.ts': 'export const core = true;',
  });
  context.after(() => rm(root, { recursive: true, force: true }));
  const violations = await checkBoundaries({ appRoot: root });
  assert.equal(violations.length, 3);
  assert.deepEqual(
    violations.map(({ rule, origin, destination }) => ({ rule, origin, destination })),
    [
      { rule: 'import-direction', origin: 'core', destination: 'tickets' },
      { rule: 'import-direction', origin: 'tickets', destination: 'users' },
      { rule: 'import-direction', origin: 'shared', destination: 'core' },
    ],
  );
});

test('detecta prefijos ajenos en literales y plantillas interpoladas, también en pruebas', async (context) => {
  const root = await fixture({
    'modules/mail/mail.ts': "const one = '/api/users';\nconst two = `/api/tickets/${id}`;",
    'modules/users/users.spec.ts': "expect('/api/mail/templates').toBeTruthy();",
  });
  context.after(() => rm(root, { recursive: true, force: true }));
  const violations = await checkBoundaries({ appRoot: root });
  assert.deepEqual(
    violations.map(({ rule, origin, destination }) => ({ rule, origin, destination })),
    [
      { rule: 'api-prefix', origin: 'mail', destination: 'users' },
      { rule: 'api-prefix', origin: 'mail', destination: 'tickets' },
      { rule: 'api-prefix', origin: 'users', destination: 'mail' },
    ],
  );
  assert.match(formatViolation(violations[0]), /\.ts:1 \[api-prefix\].*correction|\.ts:1 \[api-prefix\].*consume/);
});

test('una excepción exige origen, destino y motivo', async (context) => {
  const root = await fixture({
    'modules/tickets/page.ts': "import { user } from '../users/user';",
    'modules/users/user.ts': 'export const user = true;',
  });
  context.after(() => rm(root, { recursive: true, force: true }));
  assert.equal(
    (await checkBoundaries({
      appRoot: root,
      exceptions: [{ origin: 'tickets', destination: 'users', reason: 'contrato aprobado' }],
    })).length,
    0,
  );
  assert.equal(
    (await checkBoundaries({
      appRoot: root,
      exceptions: [{ origin: 'tickets', destination: 'users', reason: '' }],
    })).length,
    1,
  );
});
