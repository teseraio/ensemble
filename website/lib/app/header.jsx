/* This example requires Tailwind CSS v2.0+ */
import { Fragment } from 'react'
import { Popover, Transition } from '@headlessui/react'
import {
  ChartBarIcon,
  CursorClickIcon,
  DocumentReportIcon,
  MenuIcon,
  RefreshIcon,
  ShieldCheckIcon,
  ViewGridIcon,
  XIcon,
} from '@heroicons/react/outline'
import { ChevronDownIcon } from '@heroicons/react/solid'
import GithubWidget from "./github"

import Tesera from "./tesera"

/*
const solutions = [
  {
    name: 'Analytics',
    description: 'Get a better understanding of where your traffic is coming from.',
    href: '#',
    icon: ChartBarIcon,
  },
  {
    name: 'Engagement',
    description: 'Speak directly to your customers in a more meaningful way.',
    href: '#',
    icon: CursorClickIcon,
  },
  { name: 'Security', description: "Your customers' data will be safe and secure.", href: '#', icon: ShieldCheckIcon },
  {
    name: 'Integrations',
    description: "Connect with third-party tools that you're already using.",
    href: '#',
    icon: ViewGridIcon,
  },
  {
    name: 'Automations',
    description: 'Build strategic funnels that will drive your customers to convert',
    href: '#',
    icon: RefreshIcon,
  },
  {
    name: 'Reports',
    description: 'Get detailed reports that will help you make more informed decisions ',
    href: '#',
    icon: DocumentReportIcon,
  },
]
*/

/*
const resources = [
  {
    name: 'Help Center',
    description: 'Get all of your questions answered in our forums or contact support.',
    href: '#',
  },
  { name: 'Guides', description: 'Learn how to maximize our platform to get the most out of it.', href: '#' },
  { name: 'Events', description: 'See what meet-ups and other events we might be planning near you.', href: '#' },
  { name: 'Security', description: 'Understand how we take your privacy seriously.', href: '#' },
]
*/

function classNames(...classes) {
  return classes.filter(Boolean).join(' ')
}

const resources = [
  {
    name: "Changelog",
    href: "/changelog"
  },
  {
    name: "Community",
    href: "/community"
  },
  {
    name: "Docs",
    href: "/docs"
  },
  {
    name: "Enterprise",
    href: "/enterprise"
  }
]

export default function Example() {
  
  var resources2 = [...resources];

  // append enterprise
  resources2.push({
    name: "Enterprise",
    href: "https://tesera.io"
  })

  // append github (always)
  resources2.push({
    name: "Github",
    href: ""
  })

  // append download
  resources2.push({
    name: "Download",
    href: ""
  })

  return (
    <>
      <Tesera />
    <Popover className="bg-main sticky w-full top-0 z-10 text-white">
      {({ open }) => (
        <>
          <div className="flex justify-between items-center px-4 py-4 sm:px-6 lg:justify-start lg:space-x-10">
            <div>
              <a href="/" className="flex">
                <span className="sr-only">Workflow</span>
                <img
                  className="h-7 w-auto"
                  src="/logo-ensemble-white.svg"
                  alt=""
                />
              </a>
            </div>
            <div className="-mr-2 -my-2 lg:hidden">
              <Popover.Button className="bg-ensemble p-2 inline-flex items-center justify-center focus:outline-none focus:ring-2 focus:ring-inset focus:ring-indigo-500">
                <span className="sr-only">Open menu</span>
                <MenuIcon className="h-6 w-6" aria-hidden="true" />
              </Popover.Button>
            </div>
            <div className="hidden lg:flex-1 lg:flex lg:items-center lg:justify-between">
              <Popover.Group as="nav" className="flex space-x-10">
                {resources.map((item, indx) => (
                  <a href={item.href} key={indx} className="text-base font-medium">
                    {item.name}
                  </a>
                ))}
              </Popover.Group>
              <div className="flex items-center lg:ml-10">
                <GithubWidget repo="teseraio/ensemble" />
                <a
                  href="#"
                  className="ml-5 inline-flex items-center justify-center px-4 py-2 border border-transparent shadow-sm text-base font-medium text-white hover:text-black bg-black hover:bg-white"
                >
                  Download
                </a>
              </div>
            </div>
          </div>

          <Transition
            show={open}
            as={Fragment}
            enter="duration-200 ease-out"
            enterFrom="opacity-0 scale-95"
            enterTo="opacity-100 scale-100"
            leave="duration-100 ease-in"
            leaveFrom="opacity-100 scale-100"
            leaveTo="opacity-0 scale-95"
          >
            <Popover.Panel
              focus
              static
              className="absolute top-0 inset-x-0 p-2 transition transform origin-top-right lg:hidden"
            >
              <div className="ring-1 ring-black ring-opacity-5 bg-white divide-y-2 divide-gray-50">
                <div className="pt-5 pb-6 px-5">
                  <div className="flex items-center justify-between">
                    <div>
                      <img
                        className="h-8 w-auto"
                        src="/logo-ensemble-black.svg"
                        alt="Ensemble"
                      />
                    </div>
                    <div className="-mr-2">
                      <Popover.Button className="bg-black p-2 inline-flex items-center justify-center hover:bg-white hover:text-black focus:outline-none">
                        <span className="sr-only">Close menu</span>
                        <XIcon className="h-6 w-6" aria-hidden="true" />
                      </Popover.Button>
                    </div>
                  </div>
                </div>
                <div className="py-6 px-5">
                  <div className="grid grid-cols-1 gap-4">
                    {resources2.map((item) => (
                      <a
                        key={item.name}
                        href={item.href}
                        className="text-base font-medium text-gray-900 hover:text-gray-700"
                      >
                        {item.name}
                      </a>
                    ))}
                  </div>
                </div>
              </div>
            </Popover.Panel>
          </Transition>
        </>
      )}
    </Popover>
    </>
  )
}

const solutions2 = [
  {
    name: 'Analytics',
    description: 'Get a better understanding of where your traffic is coming from.',
    href: '#',
    icon: ChartBarIcon,
  },
  {
    name: 'Engagement',
    description: 'Speak directly to your customers in a more meaningful way.',
    href: '#',
    icon: CursorClickIcon,
  },
  { name: 'Security', description: "Your customers' data will be safe and secure.", href: '#', icon: ShieldCheckIcon },
  {
    name: 'Integrations',
    description: "Connect with third-party tools that you're already using.",
    href: '#',
    icon: ViewGridIcon,
  },
  {
    name: 'Automations',
    description: 'Build strategic funnels that will drive your customers to convert',
    href: '#',
    icon: RefreshIcon,
  },
]
const callsToAction = [
  { name: 'Watch Demo', href: '#', icon: PlayIcon },
  { name: 'Contact Sales', href: '#', icon: PhoneIcon },
]

import {
  BookmarkAltIcon,
  CalendarIcon,
  PhoneIcon,
  PlayIcon,
  SupportIcon,
} from '@heroicons/react/outline'

const UseCases = () => (
  <Popover className="relative">
                  {({ open }) => (
                    <>
                      <Popover.Button
                        className={classNames(
                          'group bg-main rounded-md inline-flex items-center text-base font-medium hover:text-gray-900 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500'
                        )}
                      >
                        <span>Use cases</span>
                        <ChevronDownIcon
                          className={classNames(
                            open ? 'text-gray-600' : 'text-gray-400',
                            'ml-2 h-5 w-5 group-hover:text-gray-500'
                          )}
                          aria-hidden="true"
                        />
                      </Popover.Button>

                      <Transition
                        show={open}
                        as={Fragment}
                        enter="transition ease-out duration-200"
                        enterFrom="opacity-0 translate-y-1"
                        enterTo="opacity-100 translate-y-0"
                        leave="transition ease-in duration-150"
                        leaveFrom="opacity-100 translate-y-0"
                        leaveTo="opacity-0 translate-y-1"
                      >
                        <Popover.Panel
                          static
                          className="absolute z-10 -ml-4 mt-3 transform px-2 w-screen max-w-md sm:px-0 lg:ml-0 lg:left-1/2 lg:-translate-x-1/2"
                        >
                          <div className="rounded-lg shadow-lg ring-1 ring-black ring-opacity-5 overflow-hidden">
                            <div className="relative grid gap-6 bg-white px-5 py-6 sm:gap-8 sm:p-8">
                              {solutions2.map((item) => (
                                <a
                                  key={item.name}
                                  href={item.href}
                                  className="-m-3 p-3 flex items-start rounded-lg hover:bg-gray-50"
                                >
                                  <item.icon className="flex-shrink-0 h-6 w-6 text-indigo-600" aria-hidden="true" />
                                  <div className="ml-4">
                                    <p className="text-base font-medium text-gray-900">{item.name}</p>
                                    <p className="mt-1 text-sm text-gray-500">{item.description}</p>
                                  </div>
                                </a>
                              ))}
                            </div>
                            <div className="p-5 bg-gray-50 sm:p-8">
                            <a href="#" className="-m-3 p-3 flow-root rounded-md hover:bg-gray-100">
                              <div className="flex items-center">
                                <div className="text-base font-medium text-gray-900">Enterprise</div>
                                <span className="ml-3 inline-flex items-center px-3 py-0.5 rounded-full text-xs font-medium leading-5 bg-indigo-100 text-indigo-800">
                                  New
                                </span>
                              </div>
                              <p className="mt-1 text-sm text-gray-500">
                                Empower your entire team with even more advanced tools.
                              </p>
                            </a>
                          </div>
                          </div>
                        </Popover.Panel>
                      </Transition>
                    </>
                  )}
                </Popover>
)
