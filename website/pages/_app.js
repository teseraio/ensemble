import '../styles/index.css'
import "@teseraio/oss-react-changelog/lib/styles/changelog.css"
import "@teseraio/oss-react-docs/lib/styles/docs.css"
import "@teseraio/oss-react-docs/lib/styles/prysm.css"

import React from 'react'
import Head from 'next/head'

import {ConsentManager, openConsentManager} from "@teseraio/cookie-consent-manager"
import App from "@teseraio/oss-react-app"

const navigation = [
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

function MyApp({ Component, pageProps }) {
  const {
    onBrand,
    title
  } = pageProps;

  console.log("-- on brand --")
  console.log(onBrand)

  return (
    <>
      {title &&
        <Head>
          <title>{title} | Ensemble - Data plane for database orchestration</title>
        </Head>
      }
      <App.Header
        onBrand={onBrand != undefined ? onBrand : true}
        current={''}
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
    name: "Enterprise",
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