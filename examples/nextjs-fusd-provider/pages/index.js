import Head from 'next/head'
import { useEffect, useState } from 'react'
import { Text, Divider } from "@geist-ui/react"
import PurchaseForm from '../components/PurchaseForm'
import AccountList from '../components/AccountList'
import { listAccounts } from '../lib/actions'

export default function Home() {
  const [accounts, setAccounts] = useState([])

  const syncAccounts = async () => {
    const accounts = await listAccounts()
    setAccounts(accounts)
  }

  const addAccount = account => {
    setAccounts([
      account,
      ...accounts,
    ])
  }

  useEffect(() => { syncAccounts() }, [])

  return (
    <div style={{
      maxWidth: 600, 
      padding: "1rem", 
      marginLeft: "auto", 
      marginRight: "auto"
    }}>
      <Head>
        <title>FUSD Wallet API Demo</title>
      </Head>

      <header>
        <Text h3>Purchase FUSD</Text>
        <Text p type="secondary">
          This example application demonstrates how to use the Flow Wallet API
          to distribute the FUSD stablecoin.
        </Text>
      </header>

      <PurchaseForm onPurchase={syncAccounts} />

      <Divider />

      <AccountList 
        accounts={accounts} 
        onCreate={addAccount} />
    </div>
  )
}
