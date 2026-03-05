import Nav from "@/components/Nav";
import Hero from "@/components/Hero";
import InsightStrip from "@/components/InsightStrip";
import ValueProps from "@/components/ValueProps";
import WaitlistForm from "@/components/WaitlistForm";

export default function Page() {
  return (
    <>
      <Nav />
      <main>
        <Hero />
        <InsightStrip />
        <ValueProps />
        <section
          id="waitlist"
          className="px-6 py-32 max-w-2xl mx-auto text-center"
        >
          <p className="text-xs uppercase tracking-widest text-muted mb-6 font-medium">
            early access
          </p>
          <h2 className="text-5xl font-light leading-tight mb-6">
            be first in line
          </h2>
          <p className="text-base text-muted mb-10 leading-relaxed">
            Microviva is coming soon. Join the waitlist and we&apos;ll let you
            know when it&apos;s ready.
          </p>
          <WaitlistForm />
        </section>
      </main>
      <footer className="border-t px-6 py-8 text-center text-xs text-muted"
        style={{ borderColor: "var(--border)" }}>
        © {new Date().getFullYear()} Microviva. All rights reserved.
      </footer>
    </>
  );
}
