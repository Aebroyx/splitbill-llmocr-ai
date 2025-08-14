import type { Metadata } from "next";
import localFont from "next/font/local";
import "./globals.css";
import ClientProvider from "./ClientProvider";

const geistSans = localFont({
  src: "./fonts/GeistVF.woff",
  variable: "--font-geist-sans",
  weight: "100 900",
});
const geistMono = localFont({
  src: "./fonts/GeistMonoVF.woff",
  variable: "--font-geist-mono",
  weight: "100 900",
});

export const metadata: Metadata = {
  title: "SplitBill - Split Bills with AI LLM-OCR",
  description: "Easily split bills and expenses with friends using AI-powered LLM-OCR technology. Upload receipts, extract items, and divide costs automatically.",
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
        <ClientProvider>
          {children}
        </ClientProvider>
      </body>
    </html>
  );
}
