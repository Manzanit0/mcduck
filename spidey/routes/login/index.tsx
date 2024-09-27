import { Handlers, PageProps } from "$fresh/server.ts";
import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { AuthService } from "../../gen/auth.v1/auth_connect.ts";
import { LoginResponse } from "../../gen/auth.v1/auth_pb.ts";

const url = Deno.env.get("API_HOST")!;

export const handler: Handlers = {
  async GET(_, ctx) {
    console.log(url);
    return await ctx.render();
  },
  async POST(req, _ctx) {
    const form = await req.formData();
    const email = form.get("email")?.toString();
    const password = form.get("password")?.toString();

    const transport = createConnectTransport({
      baseUrl: url!,
    });

    const client = createPromiseClient(AuthService, transport);

    let login: LoginResponse;
    try {
      login = await client.login({ email: email, password: password });
    } catch (err) {
      console.log(err);
      throw err;
    }

    const resp = new Response(null, { status: 303 });
    resp.headers.set("location", "/");
    // TODO: set Expires and Max-Age properties for cookie
    resp.headers.set("Set-Cookie", `_mcduck_fresh_key=${login.token}`);
    return resp;
  },
};

export default function Login(_: PageProps) {
  return (
    <div class="flex min-h-full flex-col justify-center px-6 py-12 lg:px-8">
      <div class="sm:mx-auto sm:w-full sm:max-w-sm">
        <img
          class="mx-auto h-10 w-auto"
          src="/logo.svg"
          alt="the Fresh logo: a sliced lemon dripping with juice"
        />
        <h2 class="mt-10 text-center text-2xl font-bold leading-9 tracking-tight text-gray-900">
          Sign in to your account
        </h2>
      </div>

      <div class="mt-10 sm:mx-auto sm:w-full sm:max-w-sm">
        <form class="space-y-6" action="#" method="POST">
          <div>
            <label
              for="email"
              class="block text-sm font-medium leading-6 text-gray-900"
            >
              Email address
            </label>
            <div class="mt-2">
              <input
                id="email"
                name="email"
                type="email"
                autocomplete="email"
                required
                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-gray-600 sm:text-sm sm:leading-6"
              />
            </div>
          </div>

          <div>
            <div class="flex items-center justify-between">
              <label
                for="password"
                class="block text-sm font-medium leading-6 text-gray-900"
              >
                Password
              </label>
              <div class="text-sm">
                <a
                  href="#"
                  class="font-semibold text-gray-900 hover:text-gray-500"
                >
                  Forgot password?
                </a>
              </div>
            </div>
            <div class="mt-2">
              <input
                id="password"
                name="password"
                type="password"
                autocomplete="current-password"
                required
                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-gray-600 sm:text-sm sm:leading-6"
              />
            </div>
          </div>

          <div>
            <button
              type="submit"
              class="flex w-full justify-center rounded-md bg-gray-800 px-3 py-1.5 text-sm font-semibold leading-6 text-white shadow-sm hover:bg-gray-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
            >
              Sign in
            </button>
          </div>
        </form>

        <p class="mt-10 text-center text-sm text-gray-500">
          Not a member?
          <a
            href="/register"
            class="pl-1 font-semibold leading-6 text-gray-900 hover:text-gray-500"
          >
            Register
          </a>
        </p>
      </div>
    </div>
  );
}
