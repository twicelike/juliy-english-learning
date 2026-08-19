import type { Metadata } from "next";
import { Geist } from "next/font/google";
import "./globals.css";
import "../themes/sketch/tokens.css";
import "../themes/sketch/components.css";
import Navbar from "@/components/layout/Navbar";
import Footer from "@/components/layout/Footer";

const geist = Geist({ subsets: ["latin"], variable: "--font-geist" });

export const metadata: Metadata = {
  title: "Juliy — English Learning",
  description:
    "Learn English smarter with modern vocabulary tools, spaced repetition, and personalized practice.",
};

import { AuthProvider } from "@/features/auth";
import { ThemeProvider } from "@/components/providers/ThemeProvider";
export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className={geist.variable}>
      <body className="min-h-screen bg-[#0a0a12] text-white antialiased flex flex-col transition-colors duration-300">
        <ThemeProvider>
          <AuthProvider>
            <Navbar />
            <main className="flex-1 pt-16">{children}</main>
            <Footer />
          </AuthProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
