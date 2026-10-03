import React from "react"

import "./SiteFooter.scss"

// ReadyPixl fork: small legal footer on every page. Privacy/Terms are linked from the
// homepage (CalOPPA / GDPR), and "Source code" is the prominent AGPL-3.0 section 13 offer
// for this modified Fider.
export const SiteFooter = () => (
  <footer className="c-site-footer">
    <div className="container c-site-footer__row">
      <span>© {new Date().getFullYear()} ReadyPixl</span>
      <nav className="c-site-footer__links">
        <a href="/privacy">Privacy Policy</a>
        <a href="/terms">Terms of Service</a>
        <a href="https://helpcenter.readypixl.com">Help Center</a>
        <a href="https://github.com/qminati/readypixl-feedback" target="_blank" rel="noopener">
          Source code
        </a>
      </nav>
    </div>
  </footer>
)
