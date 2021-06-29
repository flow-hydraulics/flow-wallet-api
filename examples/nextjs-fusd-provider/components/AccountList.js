import { useState } from 'react'
import {
  Table,
  Button,
  Spacer,
  Tooltip,
  Badge,
} from "@geist-ui/react"
import { UserPlus } from '@geist-ui/react-icons'
import { createAccount } from '../lib/actions'

const tooltipMessage = "This is the account that supplies FUSD for purchase."

export default function AccountList({ accounts, onCreate }) {

  const [isLoading, setIsLoading] = useState(false)

  const create = async event => {
    event.preventDefault()

    setIsLoading(true)

    const account = await createAccount()

    onCreate(account)

    setIsLoading(false)
  }

  return (
    <>
      <Table data={formatAccounts(accounts)}>
        <Table.Column prop="address" label="address" />
        <Table.Column prop="balance" label="FUSD balance" />
      </Table>
      
      <Spacer y={.5} />

      <Button
        auto
        size="small"
        icon={<UserPlus />}
        loading={isLoading}
        onClick={create}>
        Create test account
      </Button>
    </>
  )
}

function formatAccounts(accounts) {
  return accounts.map(account => {
    if (account.isAdmin) {
      return formatAdminAccount(account)
    }

    return account
  })
}

function formatAdminAccount(account) {
  return {
    address: (
      <>
        {account.address}
        <Tooltip text={tooltipMessage}>
          <Badge 
            style={{ marginLeft: "0.5rem" }}
            type="success">
            Supplier
          </Badge>
        </Tooltip>
      </>
    ),
    balance: account.balance,
  }
}
