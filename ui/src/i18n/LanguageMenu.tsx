import { ActionIcon, Menu, Tooltip } from '@mantine/core';
import { Languages } from 'lucide-react';

import { useTranslation } from './context';
import { LOCALES } from './locales';

/**
 * Choosing a language.
 *
 * Each option is written in the language it selects, because somebody looking
 * for their own language is looking for the word they use for it, not the
 * English name. The chosen one is remembered, so this is a decision made once.
 */
export function LanguageMenu() {
  const { locale, setLocale, t } = useTranslation();

  return (
    <Menu shadow="md" width={200} position="bottom-end">
      <Menu.Target>
        <Tooltip label={t('common.language')} withArrow>
          <ActionIcon variant="subtle" color="gray" size="lg" aria-label={t('common.language')}>
            <Languages size={18} />
          </ActionIcon>
        </Tooltip>
      </Menu.Target>
      <Menu.Dropdown>
        <Menu.Label>{t('common.language')}</Menu.Label>
        {LOCALES.map((option) => (
          <Menu.Item
            key={option.tag}
            onClick={() => setLocale(option.tag)}
            // The current one is marked rather than hidden: a menu that drops
            // the active option makes people wonder whether it applied.
            fw={option.tag === locale ? 700 : undefined}
            lang={option.tag}
          >
            {option.endonym}
          </Menu.Item>
        ))}
      </Menu.Dropdown>
    </Menu>
  );
}
