import clsx from 'clsx';
import React, { useState } from 'react';
import Link from 'next/link'

import Sidebar from './sidebar'
import hydrate from 'next-mdx-remote/hydrate'

import { useRouter } from 'next/router'

/*
export default function Docs({postData, index, sidebar}) {
    return (
        <div class="h-screen flex overflow-hidden bg-white">
          <div class="lg:hidden">
            <div class="fixed inset-0 flex z-40">
              <div class="fixed inset-0">
                <div class="absolute inset-0 bg-gray-600 opacity-75"></div>
              </div>
              <div tabindex="0" class="relative flex-1 flex flex-col max-w-xs w-full bg-white focus:outline-none">
                <div class="absolute top-0 right-0 -mr-12 pt-2">
                  <button type="button" class="ml-1 flex items-center justify-center h-10 w-10 rounded-full focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white">
                    <span class="sr-only">Close sidebar</span>
                    <svg class="h-6 w-6 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
                <div class="flex-1 h-0 pt-5 pb-4 overflow-y-auto">
                  <div class="flex-shrink-0 flex items-center px-4">
                    <img class="h-8 w-auto" src="https://tailwindui.com/img/logos/workflow-logo-indigo-600-mark-gray-900-text.svg" alt="Workflow"/>
                  </div>
                  <nav aria-label="Sidebar" class="mt-5">
                    <div class="px-2 space-y-1">
                    </div>
                  </nav>
                </div>
                <div class="flex-shrink-0 flex border-t border-gray-200 p-4">
                  dd
                </div>
              </div>
              <div class="flex-shrink-0 w-14" aria-hidden="true">
              </div>
            </div>
          </div>
        
          <div class="hidden lg:flex lg:flex-shrink-0">
            <div class="flex flex-col w-64">
              <div class="flex flex-col h-0 flex-1 border-r border-gray-200 bg-gray-100">
                <div class="flex-1 flex flex-col pt-5 pb-4 overflow-y-auto">
                    <Sidebar sidebar={sidebar.sidebar} />
                </div>
              </div>
            </div>
          </div>
          <div class="flex flex-col min-w-0 flex-1 overflow-hidden">
            <div class="lg:hidden">
              <div class="flex items-center justify-between bg-gray-50 border-b border-gray-200 px-4 py-1.5">
                <div>
                  <img class="h-8 w-auto" src="https://tailwindui.com/img/logos/workflow-mark-indigo-600.svg" alt="Workflow"/>
                </div>
                <div>
                  <button type="button" class="-mr-3 h-12 w-12 inline-flex items-center justify-center rounded-md text-gray-500 hover:text-gray-900">
                    <span class="sr-only">Open sidebar</span>
                    <svg class="h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
            <div class="flex-1 relative z-0 flex overflow-hidden">
              <main class="flex-1 relative z-0 overflow-y-auto focus:outline-none" tabindex="0">
                <div class="absolute inset-0 py-6 px-4 sm:px-6 lg:px-8">
                  <div class="h-full border-2 border-gray-200 border-dashed rounded-lg">
                      AA
                  </div>
                </div>
            </main>
            </div>
          </div>
        </div>
    )
}
*/

export default function Docs({postData, index, sidebar}) {
    let content = undefined;
    if (index != undefined) {
        content = index
    } else {
        content = hydrate(postData.contentHtml, { })
    }

    return (
        <>
        <div className="relative flex flex-col">
            <div className="flex-grow w-full max-w-7xl mx-auto xl:px-8 lg:flex">
                <div className="flex-1 min-w-0 bg-white xl:flex">
                  <div className="border-b border-gray-200 xl:border-b-0 xl:flex-shrink-0 xl:w-64 xl:border-r xl:border-gray-200 bg-white">
                      <div className="h-full pl-4 pr-6 py-6 sm:pl-6 lg:pl-8 xl:pl-0">
                      <div className="h-full relative" style={{minHeight: '12rem'}}>
                          <Sidebar sidebar={sidebar.sidebar} />
                      </div>
                      </div>
                  </div>
    
                  <div className="bg-white lg:min-w-0 lg:flex-1">
                      <div className="py-6 px-4 sm:px-6 lg:px-8">
                      <div className="relative" style={{minHeight: '36rem'}}>
                          <div className="inset-0">
                              <h1>
                                {postData.title}
                              </h1>
                              {content}
                          </div>
                      </div>
                      </div>
                  </div>

                  {/*
                  <div className="border-l border-gray-200 xl:border-l-0 xl:flex-shrink-0 xl:w-64 xl:border-l xl:border-gray-200 bg-white">
                      <div className="h-full pl-4 pr-6 py-6 sm:pl-6 lg:pl-8 xl:pl-0">
                        <div className="h-full relative">
                            AA
                        </div>
                      </div>
                  </div>
                  */}

                </div>
            </div>
        </div>
        </>
      )
}


/*
export default function Docs({postData, index, sidebar}) {
    const [isNavOpen, setOpenNav] = useState(false)

    const router = useRouter()

    let content = undefined;
    if (index != undefined) {
        content = index
    } else {
        content = hydrate(postData.contentHtml, { })
    }

    return (
        <div className="text-xl">
            <NavButton isNavOpen={isNavOpen} setOpenNav={setOpenNav} />
            <div className="z-40 w-full max-w-screen-xl mx-auto px-6">
                <div className="lg:flex -mx-6">
                    <div id="sidebar" className={clsx(
                            "fixed z-20 bg-white inset-0 w-3/5 h-full bg-white z-90 -mb-16 lg:-mb-0 lg:static lg:h-auto lg:overflow-y-visible lg:border-b-0 lg:pt-0 lg:w-1/4 lg:block lg:border-0 xl:w-1/5",
                            {
                                'hidden': !isNavOpen,
                                'z-50 pl-10 bg-tgrey-black': isNavOpen
                            }
                        )}>
                        <div id="navWrapper" className="h-full overflow-y-auto scrolling-touch lg:h-auto lg:block lg:relative lg:sticky lg:bg-transparent overflow-hidden lg:top-16">
                            <Sidebar Link={Link} path={router.query} sidebar={sidebar.sidebar} />
                        </div>
                    </div>
                    <div id="content-wrapper" className={clsx(
                        "min-h-screen w-full lg:static lg:max-h-full lg:overflow-visible lg:w-3/4 xl:w-4/5",
                        {
                            'z-10 overflow-hidden fixed': isNavOpen,
                        }
                    )}>
                        <div>
                            <div id="app" className="flex">
                                <div className="pb-16 w-full pt-5 md:pt-10">
                                    <div className="mb-3 px-6 mx-auto lg:ml-0 lg:mr-auto xl:mx-0 xl:px-12">
                                        <div className="flex items-center">
                                            <h1 className="font-bold">{postData.title}</h1>
                                        </div>
                                        {postData.subtitle &&
                                        <p className="mb-4 text-gray-600">
                                            {postData.subtitle}
                                        </p>
                                        }
                                    </div>
                                    <div className="flex">
                                        <div className="content px-6 xl:px-12 w-full mx-auto lg:ml-0 lg:mr-auto xl:mx-0">
                                            {content}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}
*/

/*
const NavButton = ({isNavOpen, setOpenNav}) => (
  <button
    type="button"
    className={"z-50 fixed block right-0 bottom-0 mr-10 mb-10 w-16 h-16 rounded-full bg-main lg:hidden"}
    onClick={() => {setOpenNav(!isNavOpen)}}
  >
        <svg
          width="24"
          height="24"
          fill="none"
          style={{"marginLeft": "1.25rem"}}
          className={clsx(
            'absolute -mt-3 transition duration-300 transform',
            {
                '-rotate-45': !isNavOpen
            }
          )}
        >
          <path
            d="M6 18L18 6M6 6l12 12"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
  </button>
)
*/
