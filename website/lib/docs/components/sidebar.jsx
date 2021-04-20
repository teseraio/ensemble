import React from 'react'
import clsx from 'clsx';
import PropTypes from 'prop-types';
import Link from 'next/link'

const {useState} = React;

export default function Sidebar({path, sidebar}) {
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



/*
import Link from 'next/link'
import React from 'react'
import { useRouter } from 'next/router'
import clsx from 'clsx';

import Chevron from '../assets/chevron.svg'

const {useState} = React;

export default function Sidebar({sidebar}) {
    return (
        <div className="pt-10 overflow-y-hidden sidebar">
            <ul>
                {sidebar.map((route, indx) => {
                    const {type} = route
                    if (type == "title") {
                        return <TitleSection key={indx} item={route} />
                    } else if (type == "sep") {
                        return <Separator />
                    }
                    return <Item key={indx} text={route.name} href={route.href} prefix={route.prefix} items={route.routes} />
                })}
            </ul>
        </div>
    )
}

const Separator = () => (
    <div>{'Split'}</div>
)

const TitleSection = ({item}) => {
    return (
        <h5
            className="pt-3 pb-2 font-bold text-xl"
        >
            {item.name}
        </h5>
    )
}

const Item = ({text, href, prefix, items}) => {
    const router = useRouter()
    const path = router.asPath

    // check if the route opens this sidebar
    let checkPath = href
    if (checkPath == undefined) {
        checkPath = prefix
    }
    const isActive = path.startsWith(checkPath)
    
    // check if the sidebar route is open
    const [isOpen, setIsOpen] = useState(isActive)
    const buttonHandler = () => {
        setIsOpen(current => !current)
    }

    return (
        <li
            id="nav"
            className={clsx(
                'relative',
            )}
        >
            <div
            className={clsx('pl-3',
                {
                    'text-main': isActive,
                }
            )}
            >
            {items ? 
                <span>
                    <a className="cursor-pointer block" onClick={buttonHandler}>
                        <Chevron className={clsx('absolute', {'rotate': isOpen})} style={{top: '17px', left: '5px'}}/>
                        <span>{text}</span>
                    </a>
                </span>
            : 
                <Link href={href}>
                    <a className="cursor-pointer">{text}</a>
                </Link>
            }
            </div>
            {items && isOpen &&
                <ul
                    className='pl-8 relative'
                    style={{top: '-3px'}}
                >
                    {items.map((item, indx) => (
                        <Item key={indx} text={item.name} href={item.href} items={item.routes} />
                    ))}
                </ul>
            }
        </li>
    )
}
*/