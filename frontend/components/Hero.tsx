export default function Hero() {
  return (
    <section className="min-h-[calc(100vh-64px)] flex flex-col justify-center px-6 py-24 max-w-5xl mx-auto">
      <div className="max-w-3xl">
        <p className="text-xs uppercase tracking-widest text-muted mb-8 font-medium">
          AI-powered oral assessment
        </p>
        <h1
          className="text-5xl md:text-6xl lg:text-7xl leading-[1.05] font-light mb-8"
          style={{ letterSpacing: "-0.02em" }}
        >
          Authenticate student thinking, not just writing.
        </h1>
        <p className="text-lg md:text-xl text-muted font-normal leading-relaxed mb-4 max-w-2xl">
          Can you trust the authenticity of any student work completed outside of the class?
          Microviva captures baseline writing in class, compares submitted work against it,
          and uses AI to run tailored and dynamic interviews that reveal whether the thinking behind 
          the work is genuine. 
        </p>
        <div className="flex flex-wrap items-center gap-4">
          <a
            href="#waitlist"
            className="inline-block px-6 py-3 text-sm font-medium transition-opacity hover:opacity-80"
            style={{ backgroundColor: "var(--ink)", color: "var(--bg)" }}
          >
            join the waitlist →
          </a>
          <a
            href="#learn-more"
            className="inline-block px-6 py-3 text-sm font-medium text-muted hover:text-ink transition-colors"
          >
            learn more ↓
          </a>
        </div>
      </div>
    </section>
  );
}
