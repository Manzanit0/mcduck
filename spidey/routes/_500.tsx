import { PageProps } from "$fresh/server.ts";
import { Head } from "$fresh/runtime.ts";

export default function Error500Page({ error }: PageProps) {
  return (
    <>
      <Head>
        <title>Internal Error</title>
      </Head>
      <div class="px-4 py-8 mx-auto bg-[#86efac]">
        <div class="max-w-screen-md mx-auto flex flex-col items-center justify-center">
          <img
            class="my-6"
            src="/logo.svg"
            width="128"
            height="128"
            alt="the Fresh logo: a sliced lemon dripping with juice"
          />
          <h1 class="text-4xl font-bold">Something has broken!</h1>
          <p class="my-4">{(error as Error).message}</p>
          <a href="/" class="underline">
            Go back home
          </a>
        </div>
      </div>
    </>
  );
}
