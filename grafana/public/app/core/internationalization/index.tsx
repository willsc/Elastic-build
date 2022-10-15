import i18n, { BackendModule, ResourceKey } from 'i18next';
import React from 'react';
import { Trans as I18NextTrans, initReactI18next } from 'react-i18next'; // eslint-disable-line no-restricted-imports

import {
  DEFAULT_LOCALE,
  ENGLISH_US,
  FRENCH_FRANCE,
  SPANISH_SPAIN,
  PSEUDO_LOCALE,
  VALID_LOCALES,
  CHINESE_SIMPLIFIED,
} from './constants';

const messageLoaders: Record<string, () => Promise<ResourceKey>> = {
  // English phrases are the default fallback string in the source, so we don't need to load the catalogue
  [ENGLISH_US]: () => Promise.resolve({}),
  [FRENCH_FRANCE]: () => import('../../../locales/fr-FR/grafana.json'),
  [SPANISH_SPAIN]: () => import('../../../locales/es-ES/grafana.json'),
  [CHINESE_SIMPLIFIED]: () => import('../../../locales/zh-Hans/grafana.json'),
  [PSEUDO_LOCALE]: () => import('../../../locales/pseudo-LOCALE/grafana.json'),
};

const loadTranslations: BackendModule = {
  type: 'backend',
  init() {},
  async read(language, namespace, callback) {
    const loader = messageLoaders[language];
    if (!loader) {
      return callback(new Error('No message loader available for ' + language), null);
    }

    const messages = await loader();
    callback(null, messages);
  },
};

export function initializeI18n(locale: string) {
  const validLocale = VALID_LOCALES.includes(locale) ? locale : DEFAULT_LOCALE;

  i18n
    .use(loadTranslations)
    .use(initReactI18next) // passes i18n down to react-i18next
    .init({
      lng: validLocale,

      // We don't bundle any translations, we load them async
      partialBundledLanguages: true,
      resources: {},

      // If translations are empty strings (no translation), fall back to the default value in source code
      returnEmptyString: false,

      pluralSeparator: '__',
    });

  // This is a placeholder so we can put a 'comment' in the message json files.
  // Starts with an underscore so it's sorted to the top of the file
  t(
    '_comment',
    'Do not manually edit this file, or update these source phrases in Crowdin. The source of truth for English strings are in the code source'
  );
}

export function changeLanguage(locale: string) {
  const validLocale = VALID_LOCALES.includes(locale) ? locale : DEFAULT_LOCALE;
  return i18n.changeLanguage(validLocale);
}

export const Trans: typeof I18NextTrans = (props) => {
  return <I18NextTrans {...props} />;
};

export const t = (id: string, defaultMessage: string, values?: Record<string, unknown>) => {
  return i18n.t(id, defaultMessage, values);
};

export const i18nDate = (value: number | Date | string, format: Intl.DateTimeFormatOptions = {}): string => {
  if (typeof value === 'string') {
    return i18nDate(new Date(value), format);
  }
  const locale = i18n.options.lng ?? DEFAULT_LOCALE;

  const dateFormatter = new Intl.DateTimeFormat(locale, format);
  return dateFormatter.format(value);
};
