import { FreshContext } from "$fresh/server.ts";
import { getAuthTokenFromRequest, AuthState } from "../lib/auth.ts";

export async function handler(req: Request, ctx: FreshContext<AuthState>) {
  // NOTE: We don't want this middleware to run N times for every resource the
  // page requests.
  if (ctx.destination !== "route") {
    const resp = await ctx.next();
    return resp;
  }

  ctx.state = getAuthTokenFromRequest(req)

  console.log("authentication middleware", ctx.destination, ctx.url.pathname);

  const resp = await ctx.next();
  return resp;
}
