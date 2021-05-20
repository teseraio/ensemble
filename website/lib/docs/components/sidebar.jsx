import React from 'react'
import clsx from 'clsx';
import PropTypes from 'prop-types';
import Link from 'next/link'

const {useState} = React;

function Sidebar({path, sidebar}) {
    path = ""

    return (
        <div class="flex flex-col flex-grow border-gray-200 pb-4 bg-white overflow-y-auto">
          <div class="mt-2 flex-grow flex flex-col">
            <nav class="flex-1 px-2 space-y-1 bg-white" aria-label="Sidebar">
              {sidebar.map((item, indx) => {
                if (item.title != undefined) {
                  return <Title key={indx} title={item.title} />
                }
                return (
                  <div key={indx} class="space-y-1">
                    <Item item={item} path={path} />
                  </div>
                )
              })}
            </nav>
          </div>
        </div>
    )
}

const Title = ({title}) => (
  <h3 className="pt-3 pb-1 font-bold text-base">
    {title}
  </h3>
)

const Item = ({item, path}) => {
  const isActive = path.startsWith(item.href != undefined ? item.href : item.prefix)

  // check if the sidebar route is open
  const [isOpen, setIsOpen] = useState(isActive)
  const buttonHandler = () => {
    setIsOpen(current => !current)
  }

  if (!item.routes) {
    return (
        <a href={item.href} class="group w-full flex items-center pr-2 pl-2 py-2 text-sm font-medium text-gray-600 rounded-md hover:text-gray-900 hover:bg-gray-50">
          {item.name}
        </a>
    )
  }
  return (
    <>
      <button onClick={buttonHandler} class={clsx("group w-full flex items-center pl-2 pr-1 py-2 text-sm font-medium rounded-md bg-white text-gray-600 hover:text-gray-900 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500")}>
        {item.name}
        <svg class={clsx("ml-auto h-5 w-5 transform group-hover:text-gray-400 transition-colors ease-in-out duration-150", {"text-gray-400 rotate-90": isOpen})} viewBox="0 0 20 20" aria-hidden="true">
          <path d="M6 6L14 10L6 14V6Z" fill="currentColor" />
        </svg>
      </button>
      {isOpen &&
        <div class="space-y-1 pl-6">
          {item.routes.map((item, indx) => (
            <Item indx={indx} item={item} path={path} />
          ))}
        </div>
      }
    </>
  )
}

/* This example requires Tailwind CSS v2.0+ */
import { Disclosure } from '@headlessui/react'

function classNames(...classes) {
  return classes.filter(Boolean).join(' ')
}

export default function Example({sidebar}) {
  return (
    <div className="flex flex-col flex-grow pb-4 overflow-y-auto">
      <div className="mt-4 mx-2 flex-grow flex flex-col">
        <nav className="flex-1 px-2 space-y-1" aria-label="Sidebar">
          {sidebar.map((item) =>
            <Item22 item={item} />
          )}
        </nav>
      </div>
    </div>
  )
}


function Item22({item}) {
  if (!item.children) {
    return (
      <div key={item.name}>
      <a
        href={item.href}
        className={classNames(
          item.current
            ? 'bg-gray-100 text-gray-900'
            : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900',
          'group w-full flex items-center pl-2 py-2 text-sm font-medium rounded-md'
        )}
      >
        {item.name}
      </a>
    </div>
    )
  }

  return (
    <Disclosure as="div" key={item.name} className="space-y-1">
      {({ open }) => (
        <>
          <Disclosure.Button
            className={classNames(
              item.current
                ? 'bg-gray-100 text-gray-900'
                : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900',
              'group w-full flex items-center pl-2 pr-1 py-2 text-sm font-medium rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500'
            )}
          >
            {item.name}
            <svg
              className={classNames(
                open ? 'text-gray-400 rotate-90' : 'text-gray-300',
                'ml-auto h-5 w-5 transform group-hover:text-gray-400 transition-colors ease-in-out duration-150'
              )}
              viewBox="0 0 20 20"
              aria-hidden="true"
            >
              <path d="M6 6L14 10L6 14V6Z" fill="currentColor" />
            </svg>
          </Disclosure.Button>
          <Disclosure.Panel className="space-y-1">
            <ul className="experiences">
              {item.children.map((subItem) => (
                <li className="ml-10">
                  <Item22 item={subItem} />
                </li>
              ))}
            </ul>
          </Disclosure.Panel>
        </>
      )}
    </Disclosure>
  )
}
