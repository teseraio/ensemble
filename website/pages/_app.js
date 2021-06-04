import '../styles/index.css'
import "@teseraio/oss-react-changelog/lib/styles/changelog.css"
import "@teseraio/oss-react-docs/lib/styles/docs.css"
import "@teseraio/oss-react-docs/lib/styles/prysm.css"

import React from 'react'
import Head from 'next/head'
import { useRouter } from 'next/router'

import {ConsentManager, openConsentManager} from "@teseraio/cookie-consent-manager"
import App from "@teseraio/oss-react-app"

const navigation = [
  {
    name: "Docs",
    href: "/docs"
  },
  {
    name: "Changelog",
    href: "/changelog"
  },
  {
    name: "Community",
    href: "/community"
  }
]

function MyApp({ Component, pageProps }) {
  const router = useRouter()

  const {
    onBrand,
    title
  } = pageProps;
  
  return (
    <>
      {title &&
        <Head>
          <title>{title} | Ensemble - Data plane for database orchestration</title>
        </Head>
      }
      <App.Header
        onBrand={onBrand != undefined ? onBrand : true}
        current={router.asPath}
        navigation={navigation}
        repo={'teseraio/ensemble'}
      />
      <Component {...pageProps} />
      <ConsentManager
        segmentWriteKey={"pGBpgRG6HG5pXv6sMMJqTVzC9ww8kjNQ"}
      />
      <App.Footer navigation={footer_links} />
    </>
  )
}

const footer_links = [
  {
    name: "Changelog",
    href: "/changelog"
  },
  {
    name: "Docs",
    href: "/docs"
  },
  {
    name: "Community",
    href: "/community"
  },
  {
    name: "Tesera",
    href: "https://tesera.io"
  },
  {
    name: "Github",
    href: "https://github.com/teseraio/ensemble"
  },
  {
    name: "Cookie manager",
    onClick: openConsentManager,
  },
]

export default MyApp