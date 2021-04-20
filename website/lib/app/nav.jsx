import Link from 'next/link'
import React, { useState } from 'react'
import { useRouter } from 'next/router'
import clsx from 'clsx';
import GithubWidget from "./github"

const SingleItem = ({title, href}) => {
  const router = useRouter()
  const path = router.asPath

  let active = path.startsWith(href)

  return (
    <li>
        <Link href={href}>
          <a className={clsx("leading-10 hover:text-black", {
            'border-b border-b-white': active,
          })}>
              {title}
          </a>
        </Link>
    </li>
  )
}


export default function Nav({Logo, links=[]}) {
  const [isOpen, setOpen] = useState(false)
  
  return (
<nav class="z-50 bg-gray-800" style={{position: 'sticky', top: '0'}}>
  <div class="max-w-7xl mx-auto px-2 sm:px-4 lg:px-8">
    <div class="relative flex items-center justify-between h-16">
      <div class="flex items-center px-2 lg:px-0">
        <div class="flex-shrink-0">
          <img class="block lg:hidden h-8 w-auto" src="https://tailwindui.com/img/logos/workflow-mark-indigo-500.svg" alt="Workflow"/>
          <img class="hidden lg:block h-8 w-auto" src="https://tailwindui.com/img/logos/workflow-logo-indigo-500-mark-white-text.svg" alt="Workflow"/>
        </div>
        <div class="hidden lg:block lg:ml-6">
          <div class="flex space-x-4">
            {links.map((item, indx) => (
              <a key={indx} href={item.href} class="bg-gray-900 text-white px-3 py-2 rounded-md text-sm font-medium">{item.title}</a>
            ))}
          </div>
        </div>
      </div>

      {/*
      <div class="flex-1 flex justify-center px-2 lg:ml-6 lg:justify-end">
        <div class="max-w-lg w-full lg:max-w-xs">
          <label for="search" class="sr-only">Search</label>
          <div class="relative">
            <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
              <svg class="h-5 w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd" />
              </svg>
            </div>
            <input id="search" name="search" class="block w-full pl-10 pr-3 py-2 border border-transparent rounded-md leading-5 bg-gray-700 text-gray-300 placeholder-gray-400 focus:outline-none focus:bg-white focus:border-white focus:ring-white focus:text-gray-900 sm:text-sm" placeholder="Search" type="search"/>
          </div>
        </div>
      </div>
      */}

      <div class="flex lg:hidden">
        <button type="button" onClick={() => {setOpen(!isOpen)}} class="inline-flex items-center justify-center p-2 rounded-md text-gray-400 hover:text-white hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white" aria-controls="mobile-menu" aria-expanded="false">
          <span class="sr-only">Open main menu</span>

          <svg class="block h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>

          <svg class="hidden h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      
      <div class="hidden lg:block lg:ml-4">
        <div class="flex items-center">

          <GithubWidget repo={'teseraio/ensemble'} />

        </div>
      </div>
    </div>
  </div>

  {isOpen &&
    <div class="lg:hidden" id="mobile-menu">
      <div class="px-2 pt-2 pb-3 space-y-1">
        {links.map((item, indx) => (
          <a key={indx} href={item.href} class="bg-gray-900 text-white block px-3 py-2 rounded-md text-base font-medium">{item.title}</a>
        ))}
      </div>
      <div class="pt-4 pb-2 border-t border-gray-700">
        <div class="px-2 space-y-1">
          <a href="#" class="block px-3 py-2 rounded-md text-base font-medium text-gray-400 hover:text-white hover:bg-gray-700">Your Profile</a>
          <a href="#" class="block px-3 py-2 rounded-md text-base font-medium text-gray-400 hover:text-white hover:bg-gray-700">Settings</a>
          <a href="#" class="block px-3 py-2 rounded-md text-base font-medium text-gray-400 hover:text-white hover:bg-gray-700">Sign out</a>
        </div>
      </div>
    </div>
  }
  
</nav>

  )
}

/*
export default function Nav({Logo, links=[]}) {
  return (
    <>
      <div className="py-1 z-30 bg-main w-full" style={{position: 'sticky', top: '0'}}>
          <div className="mx-10 h-16 items-center flex p-5 md:p-0">
              <a className="relative" href="/">
                <Logo style={{width: '150px'}} />
              </a>
              <ul className="flex space-x-10 mr-0 ml-auto text-white">
                {links.map((item, indx) => (
                  <SingleItem key={indx} title={item.title} href={item.href} />
                ))}
                <GithubWidget repo={'teseraio/ensemble'} />
              </ul>
          </div>
      </div>
    </>
  )
}
*/