import React from 'react'
import NextApp from 'next/app'
import { GeistProvider, CssBaseline } from '@geist-ui/react'

export default class App extends NextApp {
  render() {
    const { Component, pageProps } = this.props
    return (
      <GeistProvider>
        <CssBaseline />
        <Component {...pageProps} />
      </GeistProvider>
    )
  }
}
