import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { GlobalHeader } from "./components/GlobalHeader";
import { AppShell } from "./components/AppShell";
import { SessionProvider } from "./components/session-provider";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "UrbanMemory | Civic Intelligence Layer",
  description: "Split access civic mapping and governance workflow for Mumbai",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <SessionProvider>
          <GlobalHeader />
          <AppShell>{children}</AppShell>
        </SessionProvider>
      </body>
    </html>
  );
}
