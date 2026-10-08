import { readFile, readdir } from 'node:fs/promises';
import { dirname, relative, resolve, sep } from 'node:path';
import { fileURLToPath } from 'node:url';

export const BOUNDARY_EXCEPTIONS = Object.freeze([]);

const MODULE_API_PREFIX = Object.freeze({
  mail: 'mail',
  tickets: 'tickets',
  users: 'users',
});

function normalized(path) {
  return path.split(sep).join('/');
}

function areaFor(path, appRoot) {
  const local = normalized(relative(appRoot, path));
  const [first, second] = local.split('/');
  if (first === 'modules' && second) return { kind: 'module', name: second };
  if (first === 'core') return { kind: 'core', name: 'core' };
  if (first === 'shared') return { kind: 'shared', name: 'shared' };
  return { kind: 'app', name: 'app' };
}

function lineAt(source, index) {
  return source.slice(0, index).split('\n').length;
}

function importSpecifiers(source) {
  const found = [];
  const patterns = [
    /\b(?:import|export)\s+(?:type\s+)?(?:[\s\S]*?\s+from\s+)?(['"])([^'"\n]+)\1/g,
    /\bimport\s*\(\s*(['"])([^'"\n]+)\1\s*\)/g,
  ];
  for (const pattern of patterns) {
    for (const match of source.matchAll(pattern)) {
      found.push({ value: match[2], index: match.index ?? 0 });
    }
  }
  return found;
}

function textLiterals(source) {
  const found = [];
  let index = 0;
  while (index < source.length) {
    if (source[index] === '/' && source[index + 1] === '/') {
      index = source.indexOf('\n', index + 2);
      if (index === -1) break;
      continue;
    }
    if (source[index] === '/' && source[index + 1] === '*') {
      const end = source.indexOf('*/', index + 2);
      index = end === -1 ? source.length : end + 2;
      continue;
    }
    const quote = source[index];
    if (quote !== "'" && quote !== '"' && quote !== '`') {
      index += 1;
      continue;
    }
    const start = index;
    let value = '';
    index += 1;
    while (index < source.length) {
      const character = source[index];
      if (character === '\\') {
        value += character + (source[index + 1] ?? '');
        index += 2;
        continue;
      }
      if (character === quote) {
        index += 1;
        break;
      }
      value += character;
      index += 1;
    }
    found.push({ value, index: start });
  }
  return found;
}

function importAllowed(sourceArea, targetArea) {
  if (sourceArea.kind === 'app') return true;
  if (sourceArea.kind === 'module') {
    return (
      targetArea.kind === 'core' ||
      targetArea.kind === 'shared' ||
      (targetArea.kind === 'module' && targetArea.name === sourceArea.name)
    );
  }
  if (sourceArea.kind === 'core') return targetArea.kind === 'core' || targetArea.kind === 'shared';
  return targetArea.kind === 'shared';
}

function exceptionMatches(violation, exceptions) {
  return exceptions.some(
    (exception) =>
      exception.origin === violation.origin &&
      exception.destination === violation.destination &&
      typeof exception.reason === 'string' &&
      exception.reason.trim() !== '',
  );
}

async function typescriptFiles(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  const nested = await Promise.all(
    entries.map(async (entry) => {
      const path = resolve(directory, entry.name);
      if (entry.isDirectory()) return typescriptFiles(path);
      return entry.isFile() && entry.name.endsWith('.ts') ? [path] : [];
    }),
  );
  return nested.flat();
}

export async function checkBoundaries({ appRoot, exceptions = BOUNDARY_EXCEPTIONS } = {}) {
  const root = resolve(appRoot ?? resolve(dirname(fileURLToPath(import.meta.url)), '../src/app'));
  const files = await typescriptFiles(root);
  const violations = [];

  for (const file of files) {
    const source = await readFile(file, 'utf8');
    const sourceArea = areaFor(file, root);
    const origin = normalized(relative(root, file));

    for (const specifier of importSpecifiers(source)) {
      if (!specifier.value.startsWith('.')) continue;
      const target = resolve(dirname(file), specifier.value);
      const targetArea = areaFor(target, root);
      if (importAllowed(sourceArea, targetArea)) continue;
      const violation = {
        file: origin,
        line: lineAt(source, specifier.index),
        rule: 'import-direction',
        origin: sourceArea.name,
        destination: targetArea.name,
        correction: `mueve el contrato compartido a shared o consume ${sourceArea.name} sin importar ${targetArea.name}`,
      };
      if (!exceptionMatches(violation, exceptions)) violations.push(violation);
    }

    if (sourceArea.kind !== 'module' || !(sourceArea.name in MODULE_API_PREFIX)) continue;
    const allowedPrefix = MODULE_API_PREFIX[sourceArea.name];
    for (const literal of textLiterals(source)) {
      for (const match of literal.value.matchAll(/\/api\/([a-z][a-z0-9-]*)/g)) {
        const destination = match[1];
        if (destination === allowedPrefix) continue;
        const violation = {
          file: origin,
          line: lineAt(source, literal.index + (match.index ?? 0)),
          rule: 'api-prefix',
          origin: sourceArea.name,
          destination,
          correction: `consume /api/${allowedPrefix}/** desde este módulo o delega la operación al módulo ${destination}`,
        };
        if (!exceptionMatches(violation, exceptions)) violations.push(violation);
      }
    }
  }
  return violations.sort((left, right) =>
    left.file.localeCompare(right.file) || left.line - right.line || left.rule.localeCompare(right.rule),
  );
}

export function formatViolation(violation) {
  return `${violation.file}:${violation.line} [${violation.rule}] ${violation.origin} → ${violation.destination}: ${violation.correction}`;
}

async function main() {
  const violations = await checkBoundaries();
  if (violations.length === 0) {
    console.log('Fronteras del frontend: sin infracciones.');
    return;
  }
  console.error(`Fronteras del frontend: ${violations.length} infracción(es).`);
  for (const violation of violations) console.error(formatViolation(violation));
  process.exitCode = 1;
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  await main();
}
