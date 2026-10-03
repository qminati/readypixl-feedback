import React, { useEffect, useMemo, useState } from "react"
import type { SupabaseClient } from "@supabase/supabase-js"
import type { Auth as AuthWidget } from "@supabase/auth-ui-react"
import type { ThemeSupa as ThemeSupaTheme } from "@supabase/auth-ui-shared"

import { ReadyPixlIcon } from "../../components/ReadyPixlLogo"
import "./ReadyPixlSignIn.page.scss"

// The board's sign-in is the readypixl.com sign-in: same card, same Supabase Auth UI widget,
// same appearance settings (copied from apps/web/app/sign-in in Readypixl/readypixl-webapp)
// and the same ReadyPixl accounts. The board then checks the session with ReadyPixl.

interface ReadyPixlSignInPageProps {
  supabaseUrl: string
  supabaseAnonKey: string
}

interface AuthModules {
  client: SupabaseClient
  Auth: typeof AuthWidget
  ThemeSupa: typeof ThemeSupaTheme
}

const radii = { borderRadiusButton: "0.5rem", inputBorderRadius: "0.5rem" }

const appearanceVariables = {
  default: {
    colors: {
      brand: "#c50c71",
      brandAccent: "#a30a5e",
      inputBackground: "white",
      inputBorder: "#d4d4d8",
      inputBorderFocus: "#c50c71",
      inputBorderHover: "#a1a1aa",
    },
    radii,
  },
  dark: {
    colors: {
      brand: "#c50c71",
      brandAccent: "#a30a5e",
      brandButtonText: "white",
      defaultButtonBackground: "#27272a",
      defaultButtonBackgroundHover: "#3f3f46",
      defaultButtonBorder: "#3f3f46",
      defaultButtonText: "#fafafa",
      dividerBackground: "#3f3f46",
      inputBackground: "#27272a",
      inputBorder: "#3f3f46",
      inputBorderFocus: "#c50c71",
      inputBorderHover: "#52525b",
      inputText: "white",
      inputLabelText: "#a1a1aa",
      inputPlaceholder: "#71717a",
      anchorTextColor: "#a1a1aa",
      anchorTextHoverColor: "#d4d4d8",
    },
    radii,
  },
}

const isDarkTheme = (): boolean => (document.body.getAttribute("data-theme") || document.documentElement.getAttribute("data-theme")) === "dark"

const safeRedirect = (): string => {
  const redirect = new URLSearchParams(window.location.search).get("redirect")
  return redirect && redirect.startsWith("/") && !redirect.startsWith("//") ? redirect : "/"
}

const returnedWithError = (): boolean => {
  const query = new URLSearchParams(window.location.search)
  const hash = new URLSearchParams(window.location.hash.replace(/^#/, ""))
  return query.has("error") || hash.has("error")
}

const ReadyPixlSignInPage = (props: ReadyPixlSignInPageProps) => {
  const redirect = useMemo(safeRedirect, [])
  const [isDark, setIsDark] = useState(isDarkTheme)
  const [modules, setModules] = useState<AuthModules | null>(null)
  const [authState, setAuthState] = useState<"checking" | "widget" | "redirecting">("checking")
  const [failed, setFailed] = useState(returnedWithError)

  // Follow the board's light/dark theme, as readypixl.com follows its own.
  useEffect(() => {
    const observer = new MutationObserver(() => setIsDark(isDarkTheme()))
    observer.observe(document.body, { attributes: true, attributeFilter: ["data-theme"] })
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ["data-theme"] })
    return () => observer.disconnect()
  }, [])

  // The Supabase widget loads only on this page.
  useEffect(() => {
    let cancelled = false
    Promise.all([import("@supabase/supabase-js"), import("@supabase/auth-ui-react"), import("@supabase/auth-ui-shared")])
      .then(([supabase, authUI, authShared]) => {
        if (cancelled) return
        const client = supabase.createClient(props.supabaseUrl, props.supabaseAnonKey, {
          auth: { flowType: "pkce", detectSessionInUrl: true, persistSession: true, storageKey: "readypixl-feedback-auth" },
        })
        setModules({ client, Auth: authUI.Auth, ThemeSupa: authShared.ThemeSupa })
      })
      .catch(() => {
        if (cancelled) return
        setFailed(true)
        setAuthState("widget")
      })
    return () => {
      cancelled = true
    }
  }, [props.supabaseUrl, props.supabaseAnonKey])

  // Signed in to ReadyPixl here: hand the session to the board, which confirms it with ReadyPixl.
  useEffect(() => {
    if (!modules) return
    const client = modules.client
    let cancelled = false
    let initialCheckSettled = false

    const handOff = async (accessToken: string) => {
      setAuthState("redirecting")
      const response = await fetch("/_api/readypixl/session", {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ accessToken }),
      }).catch(() => undefined)
      // The board keeps its own session from here; drop the copy this tab holds.
      await client.auth.signOut({ scope: "local" }).catch(() => undefined)
      if (response && response.ok) {
        window.location.href = redirect
        return
      }
      if (!cancelled) {
        setFailed(true)
        setAuthState("widget")
      }
    }

    client.auth.getSession().then(async ({ data: { session } }) => {
      if (cancelled) return
      if (!session) {
        initialCheckSettled = true
        setAuthState("widget")
        return
      }
      const { data, error } = await client.auth.getUser()
      if (cancelled) return
      initialCheckSettled = true
      if (!error && data.user) {
        await handOff(session.access_token)
        return
      }
      setAuthState("widget")
    })

    const {
      data: { subscription },
    } = client.auth.onAuthStateChange((event, session) => {
      if (event !== "SIGNED_IN" || !initialCheckSettled || !session) return
      void handOff(session.access_token)
    })

    return () => {
      cancelled = true
      subscription.unsubscribe()
    }
  }, [modules, redirect])

  const AuthUI = modules?.Auth

  return (
    <div className="rp-signin" data-auth-state={authState}>
      <div className="rp-signin__card">
        <div className="rp-signin__head">
          <ReadyPixlIcon size={48} />
          <h1>Sign in to ReadyPixl</h1>
          <p>Continue with your account</p>
        </div>

        {failed && <div className="rp-signin__error">Sign in failed. Please try again.</div>}

        <div className="auth-widget">
          {modules && AuthUI && authState === "widget" && (
            <AuthUI
              supabaseClient={modules.client}
              appearance={{ theme: modules.ThemeSupa, variables: appearanceVariables }}
              theme={isDark ? "dark" : "default"}
              providers={["google", "apple"]}
              redirectTo={`${window.location.origin}/signin?redirect=${encodeURIComponent(redirect)}`}
              localization={{
                variables: {
                  sign_in: { social_provider_text: "Continue with {{provider}}" },
                  sign_up: { social_provider_text: "Continue with {{provider}}" },
                },
              }}
            />
          )}
        </div>

        <p className="rp-signin__legal">
          By continuing, you agree to our <a href="https://www.readypixl.com/terms">Terms of Service</a> and{" "}
          <a href="https://www.readypixl.com/privacy">Privacy Policy</a>
        </p>
      </div>
    </div>
  )
}

export default ReadyPixlSignInPage
