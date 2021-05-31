
import Hero2, {GetStarted, Download} from "@teseraio/oss-react-landing/lib/hero"
import Cta from "@teseraio/oss-react-landing/lib/cta"
import Details, {LeftCard, RightCard, DetailsList, Sep, Paragraph} from "@teseraio/oss-react-landing/lib/details"
import Features from "@teseraio/oss-react-landing/lib/features"

export default function Home() {
    return (
        <>
          <Hero />
          <EnsembleDetails />
          <Features 
            title={'Features'}
            abstract={'cdef'}
            features={features}
          />
          <Cta />
        </>
    )
}

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
        href={'#'}
        text={'Get started'}
      />,
      <Download
        href={'#'}
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
                    img={'/reasons/ensemble-arrows.svg'}
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
                    img={'/reasons/ensemble-interface.svg'}
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
                    img={'/reasons/ensemble-databases.svg'}
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

import FHighAvailability from "../assets/features/Ensemble-high-availability.svg"
import FKubernetes from "../assets/features/Ensemble-kubernetes.svg"
import FMonitoring from "../assets/features/Ensemble-monitoring.svg"
import FSecure from "../assets/features/Ensemble-secure.svg"
import FSimple from "../assets/features/Ensemble-simple.svg"
import FUpdates from "../assets/features/Ensemble-updates.svg"

const features = [
    {
      name: 'High availability',
      description:
        'Automated configuration, provision and recovery. Reliable deployments with failure recovery and zero downtime.',
      icon: FHighAvailability,
    },
    {
      name: 'Monitoring',
      description:
        'Export and analyze metrics, insights and workload analysis from any database in real time.',
      icon: FMonitoring,
    },
    {
      name: 'Simple',
      description:
        'Deploy the application as a single Kubernetes service. Simple to operate with minimal operational overhead.',
      icon: FSimple,
    },
    {
      name: 'Kubernetes native',
      description:
        'Define any database deployment using declarative Yaml and integrate with other Kubernetes services.',
      icon: FKubernetes,
    },
    {
      name: 'Security',
      description:
        'Use a consistent and secure workflow to protect the data. Enabled by default, both inflight and stored data are secured.',
      icon: FSecure,
    },
    {
      name: 'Upgrades',
      description:
        'Move between minor versions without downtime and stay up to date with security patches and improvements.',
      icon: FUpdates,
    },
  ]

Home.getInitialProps = ({ req }) => {
  return {
    title: "Home",
    onBrand: false
  }
}
