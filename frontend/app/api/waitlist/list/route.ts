import { NextRequest, NextResponse } from "next/server";
import { Redis } from "@upstash/redis";

const redis = Redis.fromEnv();

export async function GET(req: NextRequest) {
  const key = req.headers.get("x-admin-key");
  if (!process.env.ADMIN_KEY || key !== process.env.ADMIN_KEY) {
    return new NextResponse("Forbidden", { status: 403 });
  }

  const emails = await redis.smembers<string[]>("waitlist");
  const entries = await Promise.all(
    emails.map(async (email) => {
      const meta = await redis.hgetall<{ timestamp: string }>(`waitlist:meta:${email}`);
      return { email, timestamp: meta?.timestamp ?? null };
    })
  );

  entries.sort((a, b) => (a.timestamp ?? "").localeCompare(b.timestamp ?? ""));
  return NextResponse.json(entries);
}
