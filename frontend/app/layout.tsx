import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Microviva — AI-powered oral viva assessment",
  description:
    "Authenticate student thinking with AI-powered oral assessment. Combine in-class writing baselines, student submissions, and structured interviews to verify understanding and validate authorship.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
