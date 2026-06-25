// @ts-check
const { themes } = require('prism-react-renderer');

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'apiary',
  tagline: 'OpenAPI 3.1 for Go — driven by types, not comment soup.',

  url: 'https://yaop-labs.github.io',
  baseUrl: '/apiary/',
  organizationName: 'yaop-labs',
  projectName: 'apiary',
  trailingSlash: false,

  onBrokenLinks: 'throw',

  i18n: { defaultLocale: 'en', locales: ['en'] },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          routeBasePath: '/',
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl: 'https://github.com/yaop-labs/apiary/tree/main/website/',
        },
        blog: false,
        theme: { customCss: require.resolve('./src/css/custom.css') },
      }),
    ],
    [
      'redocusaurus',
      {
        // Renders the bundled example spec as a live API reference page.
        specs: [
          {
            id: 'tasks',
            spec: '../docs/tasks.yaml',
            route: '/api/',
          },
        ],
        theme: { primaryColor: '#00ADD8' },
      },
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      navbar: {
        title: 'apiary',
        items: [
          { type: 'docSidebar', sidebarId: 'docs', position: 'left', label: 'Docs' },
          { to: '/api/', label: 'API Explorer', position: 'left' },
          { href: 'https://github.com/yaop-labs/apiary', label: 'GitHub', position: 'right' },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              { label: 'Get started', to: '/installation' },
              { label: 'CLI', to: '/cli' },
              { label: 'API Explorer', to: '/api' },
            ],
          },
          {
            title: 'Project',
            items: [
              { label: 'GitHub', href: 'https://github.com/yaop-labs/apiary' },
              { label: 'Releases', href: 'https://github.com/yaop-labs/apiary/releases' },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} apiary.`,
      },
      prism: {
        theme: themes.github,
        darkTheme: themes.dracula,
        additionalLanguages: ['go', 'bash', 'yaml'],
      },
    }),
};

module.exports = config;
