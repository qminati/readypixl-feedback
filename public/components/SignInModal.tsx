import { useEffect } from "react"

interface SignInModalProps {
  isOpen: boolean
  onClose: () => void
}

// ReadyPixl fork: every "sign in" goes to /signin, the same sign-in card as readypixl.com.
export const SignInModal = (props: SignInModalProps) => {
  useEffect(() => {
    if (!props.isOpen) return
    const redirect = window.location.pathname + window.location.search
    window.location.href = `/signin?redirect=${encodeURIComponent(redirect)}`
  }, [props.isOpen])

  return null
}
