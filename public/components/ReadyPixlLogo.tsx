import React from "react"

import { LogoText } from "./LogoText"
import "./ReadyPixlLogo.scss"

interface ReadyPixlLogoProps {
  label?: string
}

// The readypixl.com header brand, copied from apps/web/components/Header.tsx in
// Readypixl/readypixl-webapp: the 36px gradient-ring "RP" mark (Inter bold) with its
// soft glow, then the official ReadyPixl wordmark (LogoText). Light and dark themes
// follow readypixl.com: white panel + #232323 ink, or #121212 panel + white ink.
export const ReadyPixlLogo = ({ label }: ReadyPixlLogoProps) => (
  <span className="c-rp-brand">
    <span className="c-rp-logo" aria-hidden="true">
      <span className="c-rp-logo__glow" />
      <span className="c-rp-logo__ring">
        <span className="c-rp-logo__panel">RP</span>
      </span>
    </span>
    <span className="c-rp-brand__wordmark">
      <LogoText height={19} variant="current" className="c-rp-brand__svg" />
    </span>
    {label && <span className="c-rp-brand__label">{label}</span>}
  </span>
)

// LogoIcon from @readypixl/brand (used on readypixl.com's sign-in card): the same mark,
// sized proportionally, without the header glow.
export const ReadyPixlIcon = ({ size = 48 }: { size?: number }) => (
  <span
    className="c-rp-logo__ring"
    aria-hidden="true"
    style={{ width: size, height: size, borderRadius: size * 0.25, padding: Math.max(1.5, size * 0.055), filter: "none" }}
  >
    <span className="c-rp-logo__panel" style={{ borderRadius: size * 0.19, fontSize: size * 0.44 }}>
      RP
    </span>
  </span>
)
