import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const mainLayoutPath = resolve(process.cwd(), 'ui', 'src', 'containers', 'MainLayout.tsx');
const source = readFileSync(mainLayoutPath, 'utf8');

const hasOrganizationHook = source.includes('useOrganizations');
const hasUserOrganizationsGate = source.includes('user?.organizations && user.organizations.length > 0 && (');

if (!hasOrganizationHook || hasUserOrganizationsGate) {
  console.error('Organization selector is still tied to auth user organizations and can be hidden in header.');
  process.exit(1);
}

console.log('Organization selector source check passed.');