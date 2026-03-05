import { NextRequest, NextResponse } from "next/server";
import { promises as fs } from "fs";
import path from "path";

const dataFile = path.join(process.cwd(), "data", "waitlist.json");

function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

export async function POST(req: NextRequest) {
  const body = await req.json().catch(() => null);
  const email: unknown = body?.email;

  if (typeof email !== "string" || !isValidEmail(email)) {
    return NextResponse.json({ error: "Invalid email address." }, { status: 400 });
  }

  // Ensure data directory exists
  await fs.mkdir(path.dirname(dataFile), { recursive: true });

  let entries: { email: string; timestamp: string }[] = [];
  try {
    const raw = await fs.readFile(dataFile, "utf-8");
    entries = JSON.parse(raw);
  } catch {
    // File doesn't exist yet — start fresh
  }

  if (entries.some((e) => e.email === email)) {
    return NextResponse.json({ error: "Already on the waitlist." }, { status: 409 });
  }

  entries.push({ email, timestamp: new Date().toISOString() });
  await fs.writeFile(dataFile, JSON.stringify(entries, null, 2));

  return NextResponse.json({ success: true });
}
