import { useSignal } from "@preact/signals";

export type MobileHeaderProps = {
  currentRoute: string;
};

export default function MobileHeader(props: MobileHeaderProps) {
  const checked = useSignal(false);

  return (
    <div class="sm:hidden mx-auto max-w-7xl px-2 sm:px-6 lg:px-8">
      <div class="relative flex h-16 items-center justify-between">
        <div class="absolute inset-y-0 left-0 flex items-center">
          <button
            type="button"
            class="relative inline-flex items-center justify-center rounded-md p-2 text-gray-400 hover:bg-gray-700 hover:text-white focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white"
            aria-controls="mobile-menu"
            aria-expanded="false"
            onClick={() => (checked.value = !checked.value)}
          >
            <span class="absolute -inset-0.5"></span>
            <span class="sr-only">Open main menu</span>
            <svg
              class={`h-6 w-6 ${checked.value ? "hidden" : "block"}`}
              fill="none"
              viewBox="0 0 24 24"
              stroke-width="1.5"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5"
              />
            </svg>
            <svg
              class={`h-6 w-6 ${checked.value ? "block" : "hidden"}`}
              fill="none"
              viewBox="0 0 24 24"
              stroke-width="1.5"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>
        <div class="flex flex-1 items-center justify-center sm:items-stretch sm:justify-start">
          <div class="flex flex-shrink-0 items-center">
            <img
              class="h-8 w-auto"
              src="/logo.svg"
              alt="the Fresh logo: a sliced lemon dripping with juice"
            />
          </div>
        </div>
      </div>
      <div
        class={`overflow-hidden transition-all duration-300 ${
          checked.value ? "max-h-64" : "max-h-0"
        }`}
        id="mobile-menu"
      >
        <div class="space-y-1 px-2 pb-3 pt-2">
          {mobileNavLink("Home", "/", props.currentRoute)}
          {mobileNavLink("Live Demo", "/greet/javier", props.currentRoute)}
          {mobileNavLink("Register", "/register", props.currentRoute)}
          {mobileNavLink("Login", "/login", props.currentRoute)}
        </div>
      </div>
    </div>
  );
}

function mobileNavLink(text: string, route: string, currentRoute: string) {
  const commonStyles = "block rounded-md px-3 py-2 text-base font-medium ";

  const selectedPageStyles = "text-white bg-gray-900 ";

  const otherPageStyles = "text-gray-300 hover:bg-gray-700 hover:text-white";

  const styles = currentRoute == route ? selectedPageStyles : otherPageStyles;

  return (
    <a href={route} class={`${commonStyles} ${styles}`} aria-current="page">
      {text}
    </a>
  );
}
