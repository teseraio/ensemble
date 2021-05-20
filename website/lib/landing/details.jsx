
/* This example requires Tailwind CSS v2.0+ */
import { AnnotationIcon, GlobeAltIcon, LightningBoltIcon, MailIcon, ScaleIcon } from '@heroicons/react/outline'
import clsx from 'clsx';

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
const communicationFeatures = [
  {
    id: 1,
    name: 'Mobile notifications',
    description:
      'Lorem ipsum, dolor sit amet consectetur adipisicing elit. Maiores impedit perferendis suscipit eaque, iste dolor cupiditate blanditiis ratione.',
    icon: AnnotationIcon,
  },
  {
    id: 2,
    name: 'Reminder emails',
    description:
      'Lorem ipsum, dolor sit amet consectetur adipisicing elit. Maiores impedit perferendis suscipit eaque, iste dolor cupiditate blanditiis ratione.',
    icon: MailIcon,
  },
]

export default function Example({title, abstract, children}) {
  return (
    <div className="text-white bg-black overflow-hidden pb-24">
      <div className="relative max-w-xl mx-auto px-4 sm:px-6 lg:px-8 lg:max-w-7xl">
        {children}
      </div>
    </div>
  )
}

export const RightCard = ({children, img, title, link, hideImg}) => (
    <div className="md:ml-16 relative mt-12 lg:mt-24 lg:grid lg:grid-cols-2 lg:gap-8 lg:items-center">
      <div className="relative">
        <TextSection
          title={title}
          link={link}
        >
          {children}
        </TextSection>
      </div>

      <div className={clsx("mt-16 lg:mt-10 -mx-4 relative lg:mt-0", {'sm:hidden md:block': hideImg})} aria-hidden="true">
        <img
          className="relative mx-auto"
          width={350}
          src={img}
          alt=""
        />
      </div>
  </div>
)

export const Sep = ()=> (
  <div className="border-t border-tgrey-white mt-12 lg:mt-24">
    {}
  </div>
)

export const LeftCard = ({children, img, title, link, hideImg}) => (
    <div className="relative mt-12 sm:mt-16 lg:mt-24">
        <div className="lg:grid lg:grid-flow-row-dense lg:grid-cols-2 lg:gap-8 lg:items-center">
            <div className="lg:col-start-2">
              <TextSection
                title={title}
                link={link}
              >
                  {children}
                </TextSection>
            </div>

            <div className="mt-16 lg:mt-10 -mx-4 relative lg:mt-0 lg:col-start-1">
              <img
                className="relative mx-auto"
                width={350}
                src={img}
                alt=""
              />
            </div>
          </div>
        </div>
)

/*
    <p className="mt-3 text-lg">
      {description}
    </p>
*/

const TextSection = ({children, title, link}) => (
    <>
    <h3 className="text-2xl mb-7 font-extrabold tracking-tight sm:text-4xl">
      {title}
    </h3>

    {children}

    {link &&
      <div className="mt-6">
        <a href={link.href} className="border table text-white hover:text-main p-4">
          {link.text}
        </a>
      </div>
    }
  </>
)

import { CheckIcon } from '@heroicons/react/outline'

export const DetailsList = ({description, list}) => (
  <>
    <p className="mb-3 text-lg">
      {description}
    </p>
    <dl className="mt-5 mb-8 space-y-7">
    {list.map((item, indx) => (
      <dt key={indx}>
        <CheckIcon className="absolute h-6 w-6 text-green-500" aria-hidden="true" />
        <p className="ml-9 text-lg leading-6">{item}</p>
      </dt>
    ))}
  </dl>
  </>
)

export const Paragraph = ({lines}) => (
  <div>
    {lines.map((line, indx) => (
        <p key={indx} className="mb-7 text-lg">
          {line}
        </p>
    ))}
  </div>
)