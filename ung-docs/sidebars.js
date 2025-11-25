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
        {
          type: 'category',
          label: 'company',
          items: [
            'cli/ung_company',
            'cli/ung_company_add',
            'cli/ung_company_edit',
            'cli/ung_company_ls',
          ],
        },
        {
          type: 'category',
          label: 'client',
          items: [
            'cli/ung_client',
            'cli/ung_client_add',
            'cli/ung_client_edit',
            'cli/ung_client_ls',
          ],
        },
        {
          type: 'category',
          label: 'contract',
          items: [
            'cli/ung_contract',
            'cli/ung_contract_add',
            'cli/ung_contract_edit',
            'cli/ung_contract_ls',
            'cli/ung_contract_pdf',
            'cli/ung_contract_email',
          ],
        },
        {
          type: 'category',
          label: 'invoice',
          items: [
            'cli/ung_invoice',
            'cli/ung_invoice_new',
            'cli/ung_invoice_ls',
            'cli/ung_invoice_pdf',
            'cli/ung_invoice_email',
            'cli/ung_invoice_batch-email',
            'cli/ung_invoice_from-time',
          ],
        },
        {
          type: 'category',
          label: 'track',
          items: [
            'cli/ung_track',
            'cli/ung_track_start',
            'cli/ung_track_stop',
            'cli/ung_track_now',
            'cli/ung_track_ls',
            'cli/ung_track_log',
          ],
        },
        {
          type: 'category',
          label: 'expense',
          items: [
            'cli/ung_expense',
            'cli/ung_expense_add',
            'cli/ung_expense_ls',
            'cli/ung_expense_report',
          ],
        },
        'cli/ung_dashboard',
        'cli/ung_create',
        {
          type: 'category',
          label: 'config',
          items: [
            'cli/ung_config',
            'cli/ung_config_init',
            'cli/ung_config_show',
            'cli/ung_config_path',
          ],
        },
        {
          type: 'category',
          label: 'version',
          items: [
            'cli/ung_version',
            'cli/ung_version_bump',
          ],
        },
        'cli/ung_upgrade',
        'cli/ung_doctor',
        {
          type: 'category',
          label: 'completion',
          items: [
            'cli/ung_completion',
            'cli/ung_completion_bash',
            'cli/ung_completion_zsh',
            'cli/ung_completion_fish',
            'cli/ung_completion_powershell',
          ],
        },
      ],
    },
  ],
};

module.exports = sidebars;
