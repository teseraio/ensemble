import Details, {LeftCard, RightCard, DetailsList, Sep, Paragraph} from "../../lib/landing/details"
import Features from "../../lib/landing/features"
import UseCases from "../../lib/landing/use-cases"

import { AnnotationIcon, GlobeAltIcon, LightningBoltIcon, MailIcon, ScaleIcon } from '@heroicons/react/outline'

const transferFeatures = [
    {
      id: 1,
      name: 'Competitive exchange rates',
      description:
        'Lorem ipsum, dolor sit amet consectetur adipisicing elit. Maiores impedit perferendis suscipit eaque, iste dolor cupiditate blanditiis ratione.',
      icon: GlobeAltIcon,
    },
    {
      id: 2,
      name: 'No hidden fees',
      description:
        'Lorem ipsum, dolor sit amet consectetur adipisicing elit. Maiores impedit perferendis suscipit eaque, iste dolor cupiditate blanditiis ratione.',
      icon: ScaleIcon,
    },
    {
      id: 3,
      name: 'Transfers are instant',
      description:
        'Lorem ipsum, dolor sit amet consectetur adipisicing elit. Maiores impedit perferendis suscipit eaque, iste dolor cupiditate blanditiis ratione.',
      icon: LightningBoltIcon,
    },
]

import FHighAvailability from "../../assets/features/Ensemble-high-availability.svg"
import FKubernetes from "../../assets/features/Ensemble-kubernetes.svg"
import FMonitoring from "../../assets/features/Ensemble-monitoring.svg"
import FSecure from "../../assets/features/Ensemble-secure.svg"
import FSimple from "../../assets/features/Ensemble-simple.svg"
import FUpdates from "../../assets/features/Ensemble-updates.svg"

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

export default function Home() {
    return (
        <>
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
            <Features 
                title={'Features'}
                abstract={'cdef'}
                features={features}
            />
            <UseCases
              features={useCases}
            />
        </>
    )
}
