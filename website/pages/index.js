
import Hero2, {GetStarted, Download} from "@teseraio/oss-react-landing/lib/hero"
import Cta from "@teseraio/oss-react-landing/lib/cta"
import Details, {LeftCard, RightCard, DetailsList, Sep, Paragraph} from "@teseraio/oss-react-landing/lib/details"
import Features from "@teseraio/oss-react-landing/lib/features"
import UseCases from "@teseraio/oss-react-landing/lib/use-cases"

export default function Home() {
    return (
        <>
          <Hero />
          <EnsembleDetails />
          <Features 
            title={'Features'}
            features={features}
          />
          <UseCases
            useCases={useCases}
          />
          <Cta
            link={{
              href: "https://tesera.io",
              text: "Learn more"
            }}
            content={(<>
              <span className="block">Ready to dive in?</span>
              <span className="tracking-normal block text-xl font-medium">Use Kubernetes to deploy databases on any cloud or on-premise.</span>
            </>)}
          />
        </>
    )
}

const useCases = [
  {
    name: 'Data-driven applications',
    text: 'Run a long lived database (e.g. Postgresql or Redis) alongside your application.'
  },
  {
      name: 'Data pipelines',
    text: 'Define a complete data-layer (e.g. Kafka and Clickhouse) to support data processing at scale.'
  },
  {
      name: 'Ad-hoc clusters',
    text: 'Create ephemeral deployments (e.g. Spark) for specific analytical jobs.'
  }
]

const Hero = () => (
  <Hero2
    title={(
      <>
        <span className="block xl:inline">Data to enrich your</span>{' '}
        <span className="block xl:inline">online business</span>
      </>
    )}
    img={'/ensemble-hero.png'}
    subtitle={'A simple and modular Kubernetes Operator to manage the lifecycle of your databases: resource provisioning, routine maintenance, monitoring or encryption among others.'}
    buttons={[
      <GetStarted
        href={'/docs/get-started'}
        text={'Get started'}
      />,
      <Download
        href={'https://github.com/teseraio/ensemble/releases'}
        text={'Download'}
      />
    ]}
  />
)

const EnsembleDetails = () => (
    <Details
                title={'Benefits'}
                abstract={'ccdd'}
            >
                <RightCard
                    title={'Full lifecycle'}
                    img={'/reasons/ensemble-arrows.png'}
                    link={{
                      text: 'Check our roadmap to learn more',
                      href: '/docs/roadmap'
                    }}
                    hideImg={true}
                >
                    <DetailsList 
                      description={"Simplify the complete operational complexity of running production databases on Kubernetes."}
                      list={[
                        'Resource provision and dynamic scaling.',
                        'Maintenance and failover tasks.',
                        'Monitoring and insights.',
                        'Security best practices for data and traffic encryption.'
                      ]}
                    />
                </RightCard>
                <Sep />
                <LeftCard
                    title={'Single interface'}
                    img={'/reasons/ensemble-interface.png'}
                    link={{
                      text: 'Read more about the specification',
                      href: '/docs/specification'
                    }}
                >
                  <Paragraph
                    lines={[
                      'A consistent Yaml specification to provision any type of cluster and their internal resources.',
                      'Each cluster supports a fine grained configuration of their settings and includes native integrations with other Kubernetes services (i.e. Secrets).',
                      'Ensemble ensures that the configuration is always up to date. If necessary, it performs a rolling update or scales transparent to the user to reach the desired configuration.',
                    ]}
                  />
                </LeftCard>
                <Sep />
                <RightCard
                    title={'One operator to rule them all'}
                    img={'/reasons/ensemble-databases.png'}
                    link={{
                      text: 'Learn more about our supported databases',
                      href: '/docs'
                    }}
                >
                    <Paragraph
                      lines={[
                        'A common interface based on the Operator pattern to provision, operate and manage a variety of databases on Kubernetes.',
                        'Use a single service to model and automate a complete data pipeline solution: databases, queues, schedulers or olap analytical warehouses.',
                        'Reduce the complexity of running databases on Kubernetes and ensure high availability and security compliance along your data-layer.'
                      ]}
                    />
                </RightCard>
            </Details>
)

const features = [
    {
      name: 'High availability',
      description:
        'Automated configuration, provision and recovery. Reliable deployments with failure recovery and zero downtime.',
      src: "/features/high-availability.png"
    },
    {
      name: 'Monitoring',
      description:
        'Export and analyze metrics, insights and workload analysis from any database in real time.',
      src: "/features/monitoring.png",
    },
    {
      name: 'Simple',
      description:
        'Deploy the application as a single Kubernetes service. Simple to operate with minimal operational overhead.',
      src: "/features/simple.png",
    },
    {
      name: 'Kubernetes native',
      description:
        'Define any database deployment using declarative Yaml and integrate with other Kubernetes services.',
      src: "/features/kubernetes-native.png",
    },
    {
      name: 'Security',
      description:
        'Use a consistent and secure workflow to protect the data. Enabled by default, both inflight and stored data are secured.',
      src: "/features/secure.png",
    },
    {
      name: 'Upgrades',
      description:
        'Move between minor versions without downtime and stay up to date with security patches and improvements.',
      src: "/features/updates.png",
    },
  ]

Home.getInitialProps = ({ req }) => {
  return {
    title: "Home",
    onBrand: false
  }
}
