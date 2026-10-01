import React from "react"

import "./ReadyPixlLogo.scss"

interface ReadyPixlLogoProps {
  size?: number
}

// ReadyPixl square brand mark, drawn like LogoIcon in @readypixl/brand: a gradient rounded
// square framing a flat panel with a bold "RP" (white panel in light theme, #121212 in dark).
export const ReadyPixlLogo = ({ size = 32 }: ReadyPixlLogoProps) => (
  <span
    className="c-rp-logo"
    aria-hidden="true"
    style={{ width: size, height: size, borderRadius: size * 0.25, padding: Math.max(1.5, size * 0.055) }}
  >
    <span className="c-rp-logo__panel" style={{ borderRadius: size * 0.19, fontSize: size * 0.44 }}>
      RP
    </span>
  </span>
)
