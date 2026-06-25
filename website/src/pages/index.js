import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import CodeBlock from '@theme/CodeBlock';
import styles from './index.module.css';

const FEATURES = [
  {
    title: 'Types, not comment soup',
    body: 'Function signatures and struct tags are the contract. No schema descriptions duplicated in comments.',
  },
  {
    title: 'Real go/types analysis',
    body: 'Cross-package, imported, and generic types resolve semantically — not guessed from bare names.',
  },
  {
    title: 'Rich schemas for free',
    body: 'validate tags become JSON-Schema constraints. Pointers become nullable. Enums are detected automatically.',
  },
  {
    title: 'Built for clients & CI',
    body: 'Stable operationIds for codegen, JSON or YAML output, and a -check mode that fails on a stale spec.',
  },
  {
    title: 'Honest output',
    body: 'Diagnostics with source positions for bad signatures, typos, unsupported methods and collisions.',
  },
  {
    title: 'Live preview',
    body: 'apiary serve renders Swagger UI that regenerates on every refresh while you edit.',
  },
];

const SAMPLE = `// CreateUser registers a new account.
// apiary:operation POST /api/v1/users
// tags: users
// errors: 400,409,500
func (h *UserHandler) CreateUser(
    ctx context.Context, req CreateUserRequest,
) (UserDTO, error) {
    // business logic — apiary never touches this
}`;

function Hero() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <header className={styles.hero}>
      <div className="container">
        <div className={styles.heroGrid}>
          <div className={styles.heroText}>
            <h1 className={styles.heroTitle}>apiary</h1>
            <p className={styles.heroTagline}>{siteConfig.tagline}</p>
            <p className={styles.heroSub}>
              Generate a complete, valid OpenAPI&nbsp;3.1 document from annotated
              Go — driven by your types.
            </p>
            <div className={styles.buttons}>
              <Link className="button button--primary button--lg" to="/installation">
                Get started →
              </Link>
              <Link className="button button--secondary button--lg" to="/api">
                API Explorer
              </Link>
              <Link
                className="button button--outline button--lg"
                href="https://github.com/yaop-labs/apiary"
              >
                GitHub
              </Link>
            </div>
          </div>
          <div className={styles.heroCode}>
            <CodeBlock language="go">{SAMPLE}</CodeBlock>
          </div>
        </div>
      </div>
    </header>
  );
}

function Features() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className={styles.featureGrid}>
          {FEATURES.map((f) => (
            <div key={f.title} className={styles.card}>
              <h3 className={styles.cardTitle}>{f.title}</h3>
              <p className={styles.cardBody}>{f.body}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

export default function Home() {
  return (
    <Layout
      title="OpenAPI 3.1 for Go"
      description="OpenAPI 3.1 generator for Go — driven by types, not comment soup."
    >
      <Hero />
      <main>
        <Features />
      </main>
    </Layout>
  );
}
