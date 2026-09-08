import { createElement } from 'react';
import type { LucideProps } from 'lucide-react';
import { connectorIcon } from './connectorIcons';

/**
 * Renders a connector's glyph by the name the catalogue stores. A component so
 * the icon is not built during another component's render, which reads as a new
 * component type each pass. createElement rather than JSX so the linter does not
 * read a dynamic component as one freshly declared on every render.
 */
export function ConnectorGlyph({ name, ...props }: { name?: string } & LucideProps) {
  return createElement(connectorIcon(name), props);
}
