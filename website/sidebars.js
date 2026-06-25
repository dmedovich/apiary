// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  docs: [
    'intro',
    'installation',
    {
      type: 'category',
      label: 'Authoring',
      collapsed: false,
      items: ['annotations', 'struct-tags', 'validation', 'security', 'frameworks'],
    },
    'cli',
    'stability',
    'migrating-from-swaggo',
  ],
};

module.exports = sidebars;
