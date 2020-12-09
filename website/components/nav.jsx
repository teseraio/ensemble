import Link from 'next/link'
import React from 'react'
import { useRouter } from 'next/router'
import clsx from 'clsx';

import SvgLike from "../assets/logo-ensemble-white.svg";
import TeseraLogo from "../assets/logo-tesera-white.svg"
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

const TeseraLink = ({href, text}) => (
  <Link href={href}>
    <a className="text-base ml-5 mt-1 hover:font-bold">
      {text}
    </a>
  </Link>
)

export default function Nav() {
  return (
    <>
      <div className="z-30 w-full bg-black">
        <div className="container mx-auto text-white py-4 px-5 md:px-0">
          <div className="flex">
            <TeseraLogo style={{width: '100px'}} />
            <TeseraLink href={'https://tesera.io'} text={'About us'} />
          </div>
        </div>
      </div>
      <div className="z-30 bg-main w-full" style={{position: 'sticky', top: '0'}}>
          <div className="container mx-auto h-16 items-center flex p-5 md:p-0">
              <a className="relative" href="/">
                <SvgLike style={{width: '150px'}} />
              </a>
              <ul className="flex space-x-10 mr-0 ml-auto text-white">
                  <SingleItem title={'Docs'} href="/docs/get-started" />
                  <GithubWidget repo={'teseraio/ensemble'} />
              </ul>
          </div>
      </div>
    </>
  )
}
