import '../styles/index.css'
import '../lib/docs/styles/docs.css'
import '../lib/changelog/styles/changelog.css'

import React from 'react'
import Head from 'next/head'

import Header from "../lib/app/header"
import Footer from "../lib/app/footer"

import {ConsentManager} from "@teseraio/cookie-consent-manager"
import {App} from "@teseraio/tesera-oss"

import Link from 'next/link'

import Ensemble from "../assets/logo-ensemble-white.svg";

const navigation = [
  {
    title: "Use cases",
    href: "/use-cases",
  },
  {
    title: "Changelog",
    href: "/changelog"
  },
  {
    title: "Docs",
    href: "/docs"
  },
  {
    title: "Community",
    href: "/community"
  },
]

function MyApp({ Component, pageProps }) {
  return (
    <>
      <Head>
        <title>Ensemble</title>
      </Head>
      <App.Header
        logoWhite={"/logo-ensemble-white.svg"}
        logoBlack={"/logo-ensemble-black.svg"}
        navigation={navigation}
        repo={'teseraio/ensemble'}
      />
      <Component {...pageProps} />
      <ConsentManager
        segmentWriteKey={"pGBpgRG6HG5pXv6sMMJqTVzC9ww8kjNQ"}
      />
      <App.Footer Link={Link} navigation={[]} social={[]} links={footer_links} icons={footer_icons}/>
    </>
  )
}

const footer_links = [
  {
    text: "About",
    href: "/about"
  },
  {
    text: "Blog",
    href: "/blog"
  },
  {
    text: "Jobs",
    href: "/jobs"
  },
  {
    text: "Press",
    href: "/press"
  },
  {
    text: "Accessibility",
    href: "/accessibility"
  },
]

const footer_icons = [
  {
    name: "Github",
    href: ""
  },
  {
    name: "Twitter",
    href: ""
  },
  {
    name: "Facebook",
    href: ""
  }
]

export default MyApp