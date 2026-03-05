import { NextRequest, NextResponse } from "next/server";
import { Redis } from "@upstash/redis";

const redis = Redis.fromEnv();

function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

export async function POST(req: NextRequest) {
  const body = await req.json().catch(() => null);
  const email: unknown = body?.email;

  if (typeof email !== "string" || !isValidEmail(email)) {
    return NextResponse.json({ error: "Invalid email address." }, { status: 400 });
  }

  const added = await redis.sadd("waitlist", email);
  if (added === 0) {
    return NextResponse.json({ error: "Already on the waitlist." }, { status: 409 });
  }

  await redis.hset(`waitlist:meta:${email}`, { timestamp: new Date().toISOString() });

  return NextResponse.json({ success: true });
}
