import type { Metadata } from "next";
import localFont from "next/font/local";
import { Inter, JetBrains_Mono } from "next/font/google";
import { getLocale } from "next-intl/server";
import "./globals.css";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  weight: ["400", "500", "600", "700", "800"],
  display: "swap",
});

const jetBrainsMono = JetBrains_Mono({
  variable: "--font-jetbrains-mono",
  subsets: ["latin"],
  weight: ["400", "500", "600"],
  display: "swap",
});

const notoColorEmoji = localFont({
  src: [
    {
      path: "../public/fonts/Noto_Color_Emoji/NotoColorEmoji-Flags.woff2",
      weight: "400",
      style: "normal",
    },
  ],
  variable: "--font-noto-emoji",
  display: "swap",
  preload: true,
  fallback: ["Apple Color Emoji", "Segoe UI Emoji", "Noto Color Emoji", "sans-serif"],
});

export const metadata: Metadata = {
  icons: {
    icon: [{ url: "/favicons/favicon.svg", type: "image/svg+xml" }],
  },
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const locale = await getLocale();

  return (
    <html
      lang={locale}
      className={`${inter.variable} ${jetBrainsMono.variable} ${notoColorEmoji.variable} h-full antialiased`}
      suppressHydrationWarning
    >
      <body className="min-h-full flex flex-col" suppressHydrationWarning>
        {children}
      </body>
    </html>
  );
}
