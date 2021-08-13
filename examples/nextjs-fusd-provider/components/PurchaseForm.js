import { useState } from 'react'
import { Input, Button, Spacer, useToasts } from "@geist-ui/react"
import { ShoppingCart } from '@geist-ui/react-icons'
import { purchaseFusd } from '../lib/actions'

export default function PurchaseForm({ onPurchase }) {

  const [isLoading, setIsLoading] = useState(false)
  
  const [, setToast] = useToasts()

  const purchase = async event => {
    event.preventDefault()

    setIsLoading(true)

    const { recipient, amount } = await purchaseFusd(
      event.target.address.value,
      event.target.amount.value,
    )

    await onPurchase()

    setToast({
      text: `${amount} FUSD delivered to ${recipient}.`,
      type: "success",
      delay: 5000,
    })

    setIsLoading(false)
  }

  return (
    <form onSubmit={purchase}>
      <div>
        <Input 
          label="Address"
          name="address"
          placeholder="0xabc..."
          width="100%" />

        <Spacer y={.5} />

        <Input label="Amount" name="amount" width="100%" />

        <Spacer y={.5} />
      </div>

      <Button
        loading={isLoading}
        icon={<ShoppingCart />}
        type="success"
        htmlType="submit">
        Purchase
      </Button>

    </form>
  )
}
