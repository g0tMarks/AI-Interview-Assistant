"use client";

import { useState, FormEvent } from "react";

export default function WaitlistForm() {
  const [email, setEmail] = useState("");
  const [status, setStatus] = useState<"idle" | "loading" | "success" | "error">("idle");
  const [message, setMessage] = useState("");

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();

    const trimmed = email.trim();
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmed)) {
      setStatus("error");
      setMessage("Please enter a valid email address.");
      return;
    }

    setStatus("loading");
    try {
      const res = await fetch("/api/waitlist", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: trimmed }),
      });
      const data = await res.json();
      if (res.ok) {
        setStatus("success");
      } else {
        setStatus("error");
        setMessage(data.error ?? "Something went wrong. Please try again.");
      }
    } catch {
      setStatus("error");
      setMessage("Network error. Please try again.");
    }
  }

  if (status === "success") {
    return (
      <p className="text-base font-medium" style={{ color: "var(--ink)" }}>
        You&apos;re on the list. We&apos;ll be in touch.
      </p>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col sm:flex-row gap-3 justify-center">
      <input
        type="email"
        value={email}
        onChange={(e) => {
          setEmail(e.target.value);
          if (status === "error") { setStatus("idle"); setMessage(""); }
        }}
        placeholder="your@email.com"
        required
        className="flex-1 sm:max-w-xs px-4 py-3 text-sm bg-transparent outline-none"
        style={{
          border: `1px solid ${status === "error" ? "#c0392b" : "var(--border)"}`,
          color: "var(--ink)",
        }}
      />
      <button
        type="submit"
        disabled={status === "loading"}
        className="px-6 py-3 text-sm font-medium transition-opacity hover:opacity-80 disabled:opacity-50 whitespace-nowrap"
        style={{ backgroundColor: "var(--ink)", color: "var(--bg)" }}
      >
        {status === "loading" ? "joining…" : "join the waitlist →"}
      </button>
      {status === "error" && (
        <p className="w-full text-xs mt-1" style={{ color: "#c0392b" }}>
          {message}
        </p>
      )}
    </form>
  );
}
