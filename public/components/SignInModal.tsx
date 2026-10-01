import React from "react"
import { Modal, LegalFooter, TenantLogo, Button } from "@fider/components"
import { CloseIcon } from "./common"
import { Trans } from "@lingui/react/macro"
import { HStack, VStack } from "./layout"

interface SignInModalProps {
  isOpen: boolean
  onClose: () => void
}

// ReadyPixl fork: the board has no login of its own; people sign in with their ReadyPixl account.
export const SignInModal: React.FC<SignInModalProps> = (props) => {
  const redirect = typeof window !== "undefined" ? window.location.pathname + window.location.search : "/"
  const signInURL = `/sso/readypixl/start?redirect=${encodeURIComponent(redirect)}`

  return (
    <Modal.Window isOpen={props.isOpen} onClose={props.onClose}>
      <Modal.Header>
        <VStack spacing={8}>
          <HStack justify="between">
            <TenantLogo size={24} useFiderIfEmpty={true} />
            <CloseIcon closeModal={props.onClose} />
          </HStack>
          <p>
            <Trans id="modal.signin.header">Join the conversation</Trans>
          </p>
        </VStack>
      </Modal.Header>
      <Modal.Content>
        <VStack spacing={4}>
          <Button variant="primary" href={signInURL}>
            <Trans id="modal.signin.readypixl">Continue with ReadyPixl</Trans>
          </Button>
          <p className="text-muted text-sm">
            <Trans id="modal.signin.readypixl.hint">Use the same account you use in the ReadyPixl app.</Trans>
          </p>
        </VStack>
      </Modal.Content>
      <LegalFooter />
    </Modal.Window>
  )
}
