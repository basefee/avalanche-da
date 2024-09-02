import React, { useState, useEffect } from 'react'
import { morpheusClient } from '../MorpheusClient'

interface FaucetProps {
    balance: bigint
    minBalance: bigint
    children: React.ReactNode
    myAddr: string
}

export default function Faucet({ children, minBalance, myAddr }: FaucetProps) {
    const [loading, setLoading] = useState(0)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        if (myAddr === "") {
            return
        }

        async function performFaucetRequest() {
            setLoading(l => l + 1)
            try {
                const initialBalance = await morpheusClient.getBalance(myAddr)
                if (initialBalance <= minBalance) {
                    await morpheusClient.requestFaucetTransfer(myAddr)
                    for (let i = 0; i < 100; i++) {
                        const balance = await morpheusClient.getBalance(myAddr)
                        if (balance !== minBalance) {
                            console.log(`Balance is ${balance}, changed from ${minBalance}`)
                            break
                        }
                        await new Promise(resolve => setTimeout(resolve, 100))
                    }
                }
            } catch (e) {
                setError((e instanceof Error && e.message) ? e.message : String(e))
            } finally {
                setLoading(l => l - 1)
            }
        }

        performFaucetRequest()
    }, [myAddr, minBalance])

    if (loading) {
        return <div>Requesting faucet funds...</div>
    }

    if (error) {
        return <div>Error: {error}</div>
    }

    return <>{children}</>
}