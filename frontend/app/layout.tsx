import { Analytics } from '@vercel/analytics/next'
import type { Metadata } from 'next'
import { Manrope, Fraunces, Geist_Mono } from 'next/font/google'
import './globals.css'
import { LangProvider } from '@/components/lang-provider'
import { ThemeProvider } from '@/components/theme-provider'
import { UserProvider } from '@/components/user-provider'

const manrope = Manrope({
  variable: '--font-sans',
  subsets: ['latin', 'cyrillic'],
})
const fraunces = Fraunces({
  variable: '--font-heading',
  subsets: ['latin'],
  weight: ['500', '600', '700'],
})
const geistMono = Geist_Mono({
  variable: '--font-geist-mono',
  subsets: ['latin'],
})

export const metadata: Metadata = {
  title: 'Via — Find your path',
  description:
    'Honest answers and real opportunities for your future in Bulgaria. Personalized education and career guidance.',
  generator: 'v0.app',
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html
      lang="bg"
      className={`${manrope.variable} ${fraunces.variable} ${geistMono.variable} bg-background`}
    >
      <head>
        <script
          dangerouslySetInnerHTML={{
            __html: `(function(){try{var t=localStorage.getItem('via-theme');if(t==='dark'||(!t&&window.matchMedia('(prefers-color-scheme: dark)').matches)){document.documentElement.classList.add('dark')}}catch(e){}})()`,
          }}
        />
      </head>
      <body className="font-sans antialiased">
        <ThemeProvider>
          <LangProvider>
            <UserProvider>{children}</UserProvider>
          </LangProvider>
        </ThemeProvider>
        {process.env.NODE_ENV === 'production' && <Analytics />}
      </body>
    </html>
  )
}
