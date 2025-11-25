/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  docs: [
    'intro',
    'installation',
    'quickstart',
    'configuration',
    {
      type: 'category',
      label: 'CLI Reference',
      link: {
        type: 'doc',
        id: 'cli/ung',
      },
      items: [
        'cli/ung_company',
        'cli/ung_client',
        'cli/ung_contract',
        'cli/ung_invoice',
        'cli/ung_track',
        'cli/ung_expense',
        'cli/ung_dashboard',
        'cli/ung_config',
        'cli/ung_create',
        'cli/ung_version',
        'cli/ung_upgrade',
      ],
    },
  ],
};

module.exports = sidebars;
