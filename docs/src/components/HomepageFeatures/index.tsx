import type { ReactNode } from "react";
import clsx from "clsx";
import Heading from "@theme/Heading";
import styles from "./styles.module.css";

type FeatureItem = {
  title: string;
  imageSrc: string;
  description: ReactNode;
};

const FeatureList: FeatureItem[] = [
  {
    title: "Real-time Monitoring",
    imageSrc: "/site-availability/img/pic1.png",
    description: (
      <>
        Monitor your applications and services in real-time with comprehensive
        availability tracking across multiple locations and sources.
      </>
    ),
  },
  {
    title: "Multi-Source Support",
    imageSrc: "/site-availability/img/pic2.png",
    description: (
      <>
        Support for multiple monitoring sources including Prometheus, custom
        APIs, and more. Easily integrate with your existing monitoring
        infrastructure.
      </>
    ),
  },
  {
    title: "Geographic Distribution",
    imageSrc: "/site-availability/img/pic3.png",
    description: (
      <>
        Monitor your services from multiple geographic locations to ensure
        global availability and performance tracking.
      </>
    ),
  },
];

function Feature({ title, imageSrc, description }: FeatureItem) {
  return (
    <div className={clsx("col col--4")}>
      <div className="text--center">
        <img src={imageSrc} alt={title} className={styles.featureImage} />
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): ReactNode {
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
