import React, {useState} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HomepageFeatures from '@site/src/components/HomepageFeatures';

import styles from './index.module.css';

const INSTALL_COMMAND = 'brew install andriiklymiuk/homebrew-ung/ung';

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      // Try modern clipboard API first
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(INSTALL_COMMAND);
      } else {
        // Fallback for non-secure contexts
        const textArea = document.createElement('textarea');
        textArea.value = INSTALL_COMMAND;
        textArea.style.position = 'fixed';
        textArea.style.left = '-999999px';
        textArea.style.top = '-999999px';
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();
        document.execCommand('copy');
        textArea.remove();
      }
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };

  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <h1 className={styles.heroTitle}>{siteConfig.title}</h1>
        <p className={styles.heroSubtitle}>{siteConfig.tagline}</p>
        <p className={styles.heroTagline}>
          A fast CLI tool for freelancers and small businesses to track time,
          generate invoices, manage contracts, and more. All data stays local.
        </p>
        <div className={styles.installCommand}>
          <pre>
            <code>{INSTALL_COMMAND}</code>
          </pre>
          <button
            className={clsx(styles.copyButton, copied && styles.copyButtonCopied)}
            onClick={handleCopy}
            title="Copy to clipboard"
            aria-label="Copy install command to clipboard"
          >
            {copied ? '✓ Copied!' : 'Copy'}
          </button>
        </div>
        <div className={styles.buttons}>
          <Link
            className="button button--secondary button--lg"
            to="/docs/intro">
            Get Started
          </Link>
          <Link
            className="button button--outline button--lg"
            href="https://marketplace.visualstudio.com/items?itemName=andriiklymiuk.ung">
            VS Code Extension
          </Link>
        </div>
      </div>
    </header>
  );
}

export default function Home(): JSX.Element {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title} - Billing & Tracking CLI`}
      description="Universal Next-Gen CLI for freelancers: time tracking, invoice generation, contract management, expense tracking. Privacy-first, all data stored locally.">
      <HomepageHeader />
      <main>
        <HomepageFeatures />
      </main>
    </Layout>
  );
}
