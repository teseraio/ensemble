import Toc from './toc'
import Nav from './nav'
import clsx from 'clsx';
import React, { useState } from 'react';

import Sidebar from './sidebar'
import hydrate from 'next-mdx-remote/hydrate'

import { useRouter } from 'next/router'

const useToc = false;

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
            <Nav/>
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
                            <Sidebar route={router.query} sidebar={sidebar.sidebar} />
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
                                        {(postData.toc && useToc) &&
                                            <div className="hidden xl:text-sm xl:block xl:w-1/4 xl:px-6">
                                                <div className="flex flex-col justify-between overflow-y-auto sticky max-h-(screen-16) pt-12 pb-4 -mt-12 top-16">
                                                    <Toc data={postData.toc} />
                                                </div>
                                            </div>    
                                        }
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
