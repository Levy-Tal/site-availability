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
    "intro",
    {
      type: "category",
      label: "Getting Started",
      items: [
        "getting-started/installation",
        "getting-started/quick-start",
        "getting-started/docker",
      ],
    },
    {
      type: "category",
      label: "Configuration",
      items: [
        "configuration/overview",
        "configuration/backend",
        "configuration/frontend",
        "configuration/http-source",
        "configuration/prometheus",
      ],
    },
    {
      type: "category",
      label: "Deployment",
      items: [
        "deployment/docker-compose",
        "deployment/kubernetes",
        "deployment/helm",
        "deployment/production",
      ],
    },
    {
      type: "category",
      label: "Development",
      items: [
        "development/setup",
        "development/architecture",
        "development/contributing",
        "development/testing",
      ],
    },
    {
      type: "category",
      label: "API Reference",
      items: [
        "api/overview",
        "api/endpoints",
        "api/authentication",
        "api/metrics",
      ],
    },
    {
      type: "category",
      label: "Frontend",
      items: [
        "frontend/overview",
        "frontend/components",
        "frontend/configuration",
      ],
    },
    "troubleshooting",
  ],
};

export default sidebars;
