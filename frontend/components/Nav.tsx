"use client";

import { useEffect, useState } from "react";

export default function Nav() {
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8);
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <nav
      className="sticky top-0 z-50 flex items-center justify-between px-6 py-5 transition-all"
      style={{
        backgroundColor: "var(--bg)",
        borderBottom: scrolled ? "1px solid var(--border)" : "1px solid transparent",
      }}
    >
      <span className="text-base font-medium tracking-tight">microviva</span>
      <a
        href="#waitlist"
        className="text-sm font-medium px-4 py-2 rounded-sm transition-opacity hover:opacity-70"
        style={{
          backgroundColor: "var(--ink)",
          color: "var(--bg)",
        }}
      >
        sign up
      </a>
    </nav>
  );
}
