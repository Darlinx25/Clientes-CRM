import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import esTranslations from './locales/es.json';

// Suppress i18next's promotional console message (hardcoded since v23)
const noop = () => {};
const origLog = console.log;
console.log = noop;
i18n
  .use(initReactI18next)
  .init({
    resources: {
      es: {
        translation: esTranslations
      }
    },
    lng: 'es',
    fallbackLng: 'es',
    load: 'languageOnly',
    debug: false,
    interpolation: {
      escapeValue: false
    }
  }).then(() => {
    console.log = origLog;
  });

export default i18n;
