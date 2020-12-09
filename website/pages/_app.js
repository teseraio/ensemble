import '../styles/index.css'

import React from 'react'
import Head from 'next/head'

function MyApp({ Component, pageProps }) {
  return (
    <>
      <Head>
        <title>Ensemble</title>
      </Head>
      <Component {...pageProps} />
    </>
  )
}

export default MyApp