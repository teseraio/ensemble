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

export default function Example() {
  return (

          <div className="bg-black text-white flex justify-between items-center px-4 py-4 sm:px-6 md:justify-start md:space-x-10">
            <div>
              <a href="/" className="flex">
                <span className="sr-only">Workflow</span>
                <img
                  className="h-7 w-auto"
                  src="/logo-tesera-white-cont.svg"
                  alt=""
                />
              </a>
            </div>
            <div className="-mr-2 -my-2 md:hidden">
            </div>
            <div className="hidden md:flex-1 md:flex md:items-center md:justify-between">
              <div as="nav" className="flex space-x-10">
                <a href="https://tesera.io" className="text-base">
                  About us
                </a>
              </div>
            </div>
          </div>

  )
}
