import { type PageProps } from "$fresh/server.ts";
import Navbar from "../components/Navbar.tsx";
import { State } from "./_middleware.ts";

export default function App({
  Component,
  state,
  url,
}: PageProps<unknown, State>) {
  return (
    <html>
      <head>
        <meta charset="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>spidey</title>
        <link rel="stylesheet" href="/styles.css" />
      </head>
      <body>
        <div class="navbar">
          <Navbar state={state} currentRoute={url.pathname} />
        </div>
        <Component />
      </body>
    </html>
  );
}
