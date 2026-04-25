import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";

const inter = Inter({ subsets: ["latin"], variable: "--font-inter" });

export const metadata: Metadata = {
  title: "CoffeeOracle",
  description:
    "Upload your latte art and let the Coffee Oracle decode delightful fortunes.",
  metadataBase: new URL("https://coffee-oracle.local"),
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className={`${inter.variable} antialiased`}>
      <body className="bg-coffee-night text-coffee-foam">
        <div className="flex min-h-screen flex-col bg-coffee-radial">
          <main className="mx-auto flex w-full max-w-6xl flex-1 flex-col px-1 py-3 sm:px-3 sm:py-7 lg:px-6 lg:py-12">
            {children}
          </main>

          <footer className="border-t border-white/5 bg-coffee-night/80">
            <div className="mx-auto flex w-full max-w-6xl items-center justify-center px-4 py-5 text-sm text-coffee-foam/60">
              <span>CoffeeOracle Labs</span>
            </div>
          </footer>
        </div>
      </body>
    </html>
  );
}
