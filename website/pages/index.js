import Nav from '../components/nav'
import Button from '../components/button'
import clsx from 'clsx';
import fs from 'fs'
//import Home from "../components/home"
//import Cta from "../lib/landing/cta"
//import Hero from "../lib/landing/hero"
import {Landing} from "@teseraio/tesera-oss"

export default function IndexPage() {
  return (
    <div className="text-xl">
     <Landing.Hero />
     <Landing.Cta />
    </div>
  )
}

const Section = ({className, children, padding=true}) => (
  <div className={clsx(className, {"py-10 md:py-24": padding})}>
    <div className={clsx("container mx-auto px-6 md:px-0")}>
      {children}
    </div>
  </div>
)

/*
const Hero = () => (
  <Section className="bg-main text-white" padding={false}>
    <div className="py-16 md:py-24">
      <div className="text-5xl font-bold w-full md:w-3/5 leading-none">{'Run production grade databases on Kubernetes at ease'}</div>
      <div className="pt-4 w-full md:w-3/6">{'A simple and modular Kubernetes Operator to manage the lifecycle of your databases: resource provisioning, routine maintenance, monitoring or encryption among others.'}</div>
      <div className="pt-10 space-x-3">
        <Button className="border text-white hover:text-black" href={'/docs/get-started'}>
          {'Get started'}
        </Button>
        <Button className="bg-black hover:bg-white hover:text-black" href={'https://github.com/teseraio/ensemble/releases'}>
          {'Download'}
        </Button>
      </div>
    </div>
  </Section>
)
*/

import FHighAvailability from "../assets/features/Ensemble-high-availability.svg"
import FKubernetes from "../assets/features/Ensemble-kubernetes.svg"
import FMonitoring from "../assets/features/Ensemble-monitoring.svg"
import FSecure from "../assets/features/Ensemble-secure.svg"
import FSimple from "../assets/features/Ensemble-simple.svg"
import FUpdates from "../assets/features/Ensemble-updates.svg"

const featuresList = [
  {
    title: 'High availability',
    text: 'Reliable deployments with failure recovery and zero downtime.',
    Image: FHighAvailability,
  },
  {
    title: 'Kubernetes native',
    text: 'Kubernetes control-plane support and integration with its native resources.',
    Image: FKubernetes,
  },
  {
    title: 'Monitoring',
    text: 'Database metrics exposed in Prometheus format and custom metrics alerts.',
    Image: FMonitoring
  },
  {
    title: 'Security',
    text: 'Out of the box at rest and in transit encryption with certificate rotation.',
    Image: FSecure,
  },
  {
    title: 'Simple',
    text: 'Easy setup and interaction with CLI and API.',
    Image: FSimple,
  },
  {
    title: 'Upgrades',
    text: 'Patch and version upgrades.',
    Image: FUpdates
  }
]

const Features2 = () => (
  <Section>
    <h1 className="text-4xl font-bold text-center mb-8">
      {'Features'}
    </h1>
    <div className="grid grid-cols-1 gap-8 md:grid-cols-2">
      {featuresList.map((item, indx) => (
        <div className="p-8 bg-white2 flex" key={indx}>
          <div className="mr-4">
            <item.Image style={{width: '72px', height: '72px'}}/>
          </div>
          <div>
            <div className="text-black font-bold">{item.title}</div>
            <div className="text-tgrey-white">{item.text}</div>
          </div>
        </div>          
      ))}
    </div>
  </Section>
)

import FDatabases from "../assets/reasons/ensemble-databases.svg"
import FInterface from "../assets/reasons/ensemble-interface.svg"
import FArrows from "../assets/reasons/ensemble-arrows.svg"

const WhyEnsemble = () => (
  <Section className="bg-black text-white why">
    <div className="space-y-10 md:space-y-20">
      <Reason showImg={false} title="Full lifecycle" Image={FArrows}>
        <WLifecycle />
      </Reason>
      <Reason addSep={true} title="Single interface" Image={FInterface} reverse={true}>
        <WInterface />
      </Reason>
      <Reason addSep={true} title="One operator to rule them all" Image={FDatabases}>
        <WDatabases />
      </Reason>
    </div>
  </Section>
)

const Reason = ({className, showImg=true, addSep=false, children, title, Image, reverse=false}) => (
  <div className={clsx("flex flex-wrap px-3 md:px-10", className, {'border-t border-tgrey-white pt-10 md:pt-20': addSep})}>
    <div className="w-full md:self-center md:w-1/2 reason">
      <h3 className="text-4xl font-bold mb-5">{title}</h3>
      {children}
    </div>
    <div className={clsx("w-full pt-16 md:pt-0 md:block md:self-center md:w-1/2 reason-image", {"order-none md:order-first": reverse}, {"hidden" : !showImg})}>
      <Image style={{maxHeight: "325px"}}/>
    </div>
  </div>
)

const Strong = ({children}) => (
  <strong className="text-main">
    {children}
  </strong>
)

import Link from 'next/link'

const WLink = ({href, text}) => (
  <div className="mt-6">
    <Link href={href}>
      <a className="border table text-white hover:text-main p-4">
        {text}
      </a>
    </Link>
  </div>
)

const WDatabases = () => (
  <>
    <p>
      {'A common interface based on the Operator pattern to provision, operate and manage a variety of databases on Kubernetes.'}
    </p>
    <p>
      {'Use a single service to model and automate a complete data pipeline solution: databases, queues, schedulers or olap analytical warehouses.'}
    </p>
    <p>
      {'Reduce the complexity of running databases on Kubernetes and ensure high availability and security compliance along your data-layer.'}
    </p>
    <WLink href="/docs" text={'Learn more about our supported databases'} />
  </>
)

const WInterface = () => (
  <>
    <p>
      {'A consistent Yaml specification to provision any type of cluster and their internal resources.'}
    </p>
    <p>
      {'Each cluster supports a fine grained configuration of their settings and includes native integrations with other Kubernetes services (i.e. Secrets).'}
    </p>
    <p>
      {'Ensemble ensures that the configuration is always up to date. If necessary, it performs a rolling update or scales transparent to the user to reach the desired configuration.'}
    </p>
    <WLink href="/docs/specification" text={'Read more about the specification'} />
  </>
)

const WLifecycle = () => (
  <>
    <p>
      {'Simplify the complete operational complexity of running production databases on Kubernetes.'}
    </p>
    <ul className="space-y-3">
      <li>{'Resource provision and dynamic scaling.'}</li>
      <li>{'Maintenance and failover tasks.'}</li>
      <li>{'Monitoring and insights.'}</li>
      <li>{'Security best practices for data and traffic encryption.'}</li>
    </ul>
    <WLink href="/docs/roadmap" text={'Check our roadmap to learn more'} />
  </>
)

const CTA = () => (
  <Section className="bg-black text-white">
    <div className="border p-10">
      <h1 className="text-4xl font-bold">{'Ready for more?'}</h1>
      <div className="mt-4 sm:w-full md:w-3/5">{'Tesera offers an Enterprise package for Ensemble which includes VM execution, periodic backups, restores or autoscaling among other things.'}</div>
      <div className="mt-10">
        <Button className="bg-main text-white hover:bg-white" href={'https://tesera.io'}>
          {'Learn more'}
        </Button>
      </div>
    </div>
  </Section>
)

const useCases = [
  {
    title: 'Data-driven applications',
    text: 'Run a long lived database (e.g. Postgresql or Redis) alongside your application.'
  },
  {
    title: 'Data pipelines',
    text: 'Define a complete data-layer (e.g. Kafka and Clickhouse) to support data processing at scale.'
  },
  {
    title: 'Ad-hoc clusters',
    text: 'Create ephemeral deployments (e.g. Spark) for specific analytical jobs.'
  }
]

const UseCases = () => (
  <Section className="use-cases bg-white2" border={false}>
    <div className="text-4xl font-bold text-center mb-8">
      <h2>{'Use cases'}</h2>
    </div>
    <div className="grid grid-cols-1 gap-8 md:grid-cols-3">
      {useCases.map((item, indx) => (
        <div key={indx} className="bg-white use-case-shadow border border-black p-10">
          <div className="font-bold text-black mb-3">{item.title}</div>
          <div className="text-tgrey-white">{item.text}</div>
        </div>
      ))}
    </div>
  </Section>
)
