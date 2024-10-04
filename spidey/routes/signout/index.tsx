import { FreshContext } from "$fresh/server.ts";

export const handler = (_req: Request, _ctx: FreshContext): Response => {
  console.log("signout");

  const resp = new Response(null, { status: 303 });
  resp.headers.set("location", "/");
  resp.headers.set(
    "Set-Cookie",
    `_mcduck_fresh_key=""; expires=${new Date(0).toUTCString()};`,
  );
  return resp;
};
