import React from "react"
import { classSet } from "@fider/services"

import "./PoweredByFider.scss"

interface PoweredByFiderProps {
  slot: string
  className?: string
}

// ReadyPixl fork: no "Powered by Fider" branding. The source link stays because this board
// runs a modified Fider under AGPL-3.0, which requires offering its source to every visitor.
export const PoweredByFider = (props: PoweredByFiderProps) => {
  const className = classSet({
    "c-powered": true,
    [props.className || ""]: props.className,
  })

  return (
    <div className={className} data-slot={props.slot}>
      <a rel="noopener" className="text-2xs" href="https://github.com/qminati/readypixl-feedback" target="_blank">
        Source code
      </a>
    </div>
  )
}
