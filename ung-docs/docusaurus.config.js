// @ts-check

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'UNG',
  tagline: 'Universal Next-Gen Billing & Tracking CLI',
  url: 'https://andriiklymiuk.github.io',
  baseUrl: '/ung/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon.ico',

  organizationName: 'Andriiklymiuk',
  projectName: 'ung',
  deploymentBranch: 'gh-pages',
  trailingSlash: false,

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl: 'https://github.com/Andriiklymiuk/ung/tree/main/ung-docs/',
        },
        blog: false,
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      navbar: {
        title: 'UNG',
        logo: {
          alt: 'UNG Logo',
          src: 'img/logo.svg',
        },
        items: [
          {
            type: 'doc',
            docId: 'intro',
            position: 'left',
            label: 'Docs',
          },
          {
            type: 'doc',
            docId: 'cli/ung',
            position: 'left',
            label: 'CLI Reference',
          },
          {
            href: 'https://marketplace.visualstudio.com/items?itemName=andriiklymiuk.ung',
            label: 'VS Code Extension',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              {
                label: 'Introduction',
                to: '/docs/intro',
              },
              {
                label: 'CLI Reference',
                to: '/docs/cli/ung',
              },
            ],
          },
          {
            title: 'More',
            items: [
              {
                label: 'VS Code Extension',
                href: 'https://marketplace.visualstudio.com/items?itemName=andriiklymiuk.ung',
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} UNG. Built with Docusaurus.`,
      },
      colorMode: {
        defaultMode: 'dark',
        disableSwitch: false,
        respectPrefersColorScheme: false,
      },
      prism: {
        theme: require('prism-react-renderer').themes.github,
        darkTheme: require('prism-react-renderer').themes.dracula,
        additionalLanguages: ['bash', 'go'],
      },
    }),
};

module.exports = config;
