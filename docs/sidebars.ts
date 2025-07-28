// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

/**
 * Create as many sidebars as you want.
 */
const sidebars = {
  tutorialSidebar: [
    "introduction",
    {
      type: "category",
      label: "Usage",
      items: [
        "usage/quickstart",
        "usage/terminology",
        {
          type: "category",
          label: "Configuration",
          items: [
            "usage/configuration/server",
            {
              type: "category",
              label: "Sources",
              items: [
                "usage/configuration/sources/prometheus",
                "usage/configuration/sources/site",
                "usage/configuration/sources/http",
              ],
            },
          ],
        },
        {
          type: "category",
          label: "Installation",
          items: [
            "usage/installation/docker-compose",
            "usage/installation/helm-chart",
          ],
        },
        "usage/troubleshooting",
      ],
    },
    {
      type: "category",
      label: "Development",
      items: [
        "development/architecture",
        "development/setup",
        "development/frontend",
        {
          type: "category",
          label: "Backend",
          items: ["development/backend/sources"],
        },
        "development/testing",
        "development/contributing",
      ],
    },
    {
      type: "category",
      label: "API",
      items: ["api/overview", "api/endpoints"],
    },
    "metrics",
  ],
};

export default sidebars;
