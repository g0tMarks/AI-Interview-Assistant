type Section = {
  label: string;
  subheading: string;
  body: string | JSX.Element;
  align: "left" | "right";
};

const sections: Section[] = [
  {
    label: "know your students",
    subheading: "Build a writing profile for every student.",
    body: "Microviva creates a baseline writing profile from in-class work so you can identify authorship inconsistencies when new submissions appear.",
    align: "left" as const,
  },
  {
    label: "beyond the submission",
    subheading: "Probe the thinking behind the work.",
    body: "Microviva generates follow-up questions directly from a student's submission, helping you test understanding and clarify how the work was produced.",
    align: "right" as const,
  },
  {
    label: "evidence, not instinct",
    subheading: "Defensible evidence of understanding.",
    body: (
      <>
        Every interview produces a transcript, summary of responses, and
        authorship signals that support {" "}
        <strong>not replace</strong> academic judgement.
      </>
    ),
    align: "left" as const,
  },
];

export default function ValueProps() {
  return (
    <section className="max-w-4xl mx-auto px-6">
      {sections.map((s, i) => (
        <div key={s.label}>
          {i > 0 && (
            <hr style={{ borderColor: "var(--border)" }} />
          )}
          <div
            className={`py-24 flex flex-col ${
              s.align === "right" ? "items-end text-right" : "items-start text-left"
            }`}
          >
            <p
              className="text-xs uppercase tracking-widest font-medium mb-5"
              style={{ color: "var(--muted)" }}
            >
              {s.label}
            </p>
            <h3
              className="text-3xl md:text-4xl font-semibold mb-5 max-w-xl leading-tight"
              style={{ letterSpacing: "-0.02em" }}
            >
              {s.subheading}
            </h3>
            <p
              className="text-base leading-relaxed max-w-lg"
              style={{ color: "var(--muted)" }}
            >
              {s.body}
            </p>
          </div>
        </div>
      ))}
    </section>
  );
}
