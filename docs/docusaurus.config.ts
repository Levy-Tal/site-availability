import { themes as prismThemes } from "prism-react-renderer";
import type { Config } from "@docusaurus/types";

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

const config = {
  title: "Site Availability",
  tagline:
    "Monitor the availability of applications and services across multiple locations",
  favicon: "img/favicon.ico",

  // Future flags, see https://docusaurus.io/docs/api/docusaurus-config#future
  future: {
    v4: true, // Improve compatibility with the upcoming Docusaurus v4
  },

  // Set the production url of your site here
  url: "https://levy-tal.github.io",
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: "/site-availability/",

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: "Levy-Tal", // Usually your GitHub org/user name.
  projectName: "site-availability", // Usually your repo name.

  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },

  presets: [
    [
      "@docusaurus/preset-classic",
      {
        docs: {
          sidebarPath: require.resolve("./sidebars.js"),
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            "https://github.com/Levy-Tal/site-availability/tree/main/docs/",
        },
        blog: {
          showReadingTime: true,
          feedOptions: {
            type: ["rss", "atom"],
            xslt: true,
          },
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            "https://github.com/Levy-Tal/site-availability/tree/main/docs/",
          // Useful options to enforce blogging best practices
          onInlineTags: "warn",
          onInlineAuthors: "warn",
          onUntruncatedBlogPosts: "warn",
        },
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
      },
    ],
  ],

  themes: [
    [
      require.resolve("@easyops-cn/docusaurus-search-local"),
      {
        hashed: true,
        language: ["en"],
        highlightSearchTermsOnTargetPage: true,
        explicitSearchResultPath: true,
      },
    ],
  ],

  themeConfig: {
    // Replace with your project's social card
    image: "img/docusaurus-social-card.jpg",
    navbar: {
      title: "Site Availability",
      logo: {
        alt: "Site Availability Logo",
        src: "img/logo.png",
      },
      items: [
        {
          type: "doc",
          docId: "introduction",
          activeBaseRegex: "docs/(?!api|development)",
          position: "left",
          label: "Documentation",
        },
        { to: "/blog", label: "Blog", position: "left" },
        //        {
        //          label: 'More',
        //          position: 'left',
        //          items: [
        //            {
        //              to: '/docs/api/overview',
        //              label: 'API Reference',
        //            },
        //            {
        //              to: '/docs/development/contributing',
        //              label: 'Contributing',
        //            },
        //
        //          ],
        //        },
        {
          href: "https://github.com/Levy-Tal/site-availability",
          label: "GitHub",
          position: "right",
        },
      ],
    },
    footer: {
      style: "dark",
      links: [
        {
          title: "Docs",
          items: [
            {
              label: "Getting Started",
              to: "/docs/introduction",
            },
          ],
        },
        {
          title: "Community",
          items: [
            {
              label: "Issues",
              href: "https://github.com/Levy-Tal/site-availability/issues",
            },
            {
              label: "Discussions",
              href: "https://github.com/Levy-Tal/site-availability/discussions",
            },
          ],
        },
        {
          title: "More",
          items: [
            {
              label: "Blog",
              to: "/blog",
            },
            {
              label: "GitHub",
              href: "https://github.com/Levy-Tal/site-availability",
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Site Availability. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: [
        "diff",
        "json",
        "docker",
        "javascript",
        "css",
        "bash",
        "nginx",
        "ini",
        "yaml",
        "go",
      ],
    },
    tableOfContents: {
      minHeadingLevel: 2,
      maxHeadingLevel: 4,
    },
    colorMode: {
      respectPrefersColorScheme: false,
    },
  } as const,

  markdown: {
    format: "detect",
  },
} satisfies Config;

export default config;
