import { FreshContext } from "$fresh/server.ts";
import { getCookies } from "$std/http/cookie.ts";
import { decodeJwt } from "$jose/index.ts";

export interface State {
  loggedIn: boolean;
  authToken: string;
  userEmail: string;
}

const authCookieName = "_mcduck_fresh_key";
export async function handler(req: Request, ctx: FreshContext<State>) {
  // NOTE: We don't want this middleware to run N times for every resource the
  // page requests.
  if (ctx.destination !== "route") {
    const resp = await ctx.next();
    return resp;
  }

  console.log("authentication middleware", ctx.destination, ctx.url.pathname);

  const cookies = getCookies(req.headers);
  if (cookies[authCookieName] !== undefined && cookies[authCookieName] !== "") {
    ctx.state.loggedIn = true;
    ctx.state.authToken = cookies[authCookieName];

    const jwt = decodeJwt(cookies[authCookieName]);
    ctx.state.userEmail = jwt.sub!;
  }

  const resp = await ctx.next();
  return resp;
}

