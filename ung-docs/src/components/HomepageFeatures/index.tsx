import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  Svg: React.ComponentType<React.ComponentProps<'svg'>>;
  description: JSX.Element;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'Time Tracking',
    Svg: require('@site/static/img/features/clock.svg').default,
    description: (
      <>
        Track billable hours with automatic contract rate calculation.
        Start, stop, and log time entries per client.
      </>
    ),
  },
  {
    title: 'Invoice Generation',
    Svg: require('@site/static/img/features/invoice.svg').default,
    description: (
      <>
        Generate professional PDF invoices from tracked time or manual entries.
        Email invoices directly to clients.
      </>
    ),
  },
  {
    title: 'Contract Management',
    Svg: require('@site/static/img/features/contract.svg').default,
    description: (
      <>
        Create and store contracts with hourly or fixed rates.
        Generate PDF contracts for clients.
      </>
    ),
  },
  {
    title: 'Expense Tracking',
    Svg: require('@site/static/img/features/expense.svg').default,
    description: (
      <>
        Log business expenses by category. Generate expense reports
        for tax time or client billing.
      </>
    ),
  },
  {
    title: 'Privacy First',
    Svg: require('@site/static/img/features/privacy.svg').default,
    description: (
      <>
        All data stored locally in SQLite. No cloud, no subscriptions,
        no data sharing. Your business data stays yours.
      </>
    ),
  },
  {
    title: 'Cross-Platform',
    Svg: require('@site/static/img/features/platform.svg').default,
    description: (
      <>
        Native Go binary with instant startup. Works on macOS, Linux,
        and Windows. Also available as VSCode extension.
      </>
    ),
  },
];

function Feature({title, Svg, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <Svg className={styles.featureSvg} role="img" />
      </div>
      <div className="text--center padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): JSX.Element {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
