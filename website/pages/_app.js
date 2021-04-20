import '../styles/index.css'

import React from 'react'
import Head from 'next/head'
import Nav from "../lib/app/nav"
import Footer from "../lib/app/footer"
import Link from 'next/link'

import Ensemble from "../assets/logo-ensemble-white.svg";

const links = [
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
      <Nav Logo={Ensemble} links={links} />
      <Component {...pageProps} />
      <Footer Link={Link} links={footer_links} icons={footer_icons}/>
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