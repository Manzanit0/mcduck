import MobileHeader from "../islands/MobileHeader.tsx";
import { State } from "../routes/_middleware.ts";

export type NavbarProps = {
  state: State;
  currentRoute: string;
};

export default function Navbar(props: NavbarProps) {
  if (props.state && props.state.loggedIn) {
    return (
      <nav class="bg-gray-800">
        <MobileHeader {...props} />
        <div class="max-sm:hidden mx-auto max-w-7xl px-2 sm:px-6 lg:px-8">
          <div class="relative flex h-16 items-center justify-between">
            <div class="flex flex-1 items-center justify-center sm:items-stretch sm:justify-start">
              <div class="flex flex-shrink-0 items-center">
                <img
                  class="h-8 w-auto"
                  src="/logo.svg"
                  alt="the Fresh logo: a sliced lemon dripping with juice"
                />
              </div>
              <div class="hidden sm:ml-6 sm:block">
                <div class="flex space-x-4">
                  {navLink("Dashboard", "/", props.currentRoute)}
                  {navLink("Expenses", "/greet/javier", props.currentRoute)}
                  {navLink("Receips", "/greet/javier", props.currentRoute)}
                </div>
              </div>
            </div>
            <div class="absolute inset-y-0 right-0 flex items-center pr-2 sm:static sm:inset-auto sm:ml-6 sm:pr-0">
              <div class="hidden sm:ml-6 sm:block">
                <div class="flex space-x-4">
                  {navLink("Signout", "/", props.currentRoute)}
                  <div class="rounded-md px-3 py-2 text-sm font-medium text-white ">
                    Hello {props.state.userEmail}!
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </nav>
    );
  } else {
    return (
      <nav class="bg-gray-800">
        <MobileHeader {...props} />
        <div class="max-sm:hidden mx-auto max-w-7xl px-2 sm:px-6 lg:px-8">
          <div class="relative flex h-16 items-center justify-between">
            <div class="flex flex-1 items-center justify-center sm:items-stretch sm:justify-start">
              <div class="flex flex-shrink-0 items-center">
                <img
                  class="h-8 w-auto"
                  src="/logo.svg"
                  alt="the Fresh logo: a sliced lemon dripping with juice"
                />
              </div>
              <div class="hidden sm:ml-6 sm:block">
                <div class="flex space-x-4">
                  {navLink("Home", "/", props.currentRoute)}
                  {navLink("Live Demo", "/greet/javier", props.currentRoute)}
                </div>
              </div>
            </div>
            <div class="absolute inset-y-0 right-0 flex items-center pr-2 sm:static sm:inset-auto sm:ml-6 sm:pr-0">
              <div class="hidden sm:ml-6 sm:block">
                <div class="flex space-x-4">
                  {navLink("Register", "/register", props.currentRoute)}
                  {navLink("Login", "/login", props.currentRoute)}
                </div>
              </div>
            </div>
          </div>
        </div>
      </nav>
    );
  }
}

function navLink(text: string, route: string, currentRoute: string) {
  const selectedPageStyles =
    "rounded-md bg-gray-900 px-3 py-2 text-sm font-medium text-white";

  const otherPageStyles =
    "rounded-md px-3 py-2 text-sm font-medium text-gray-300 hover:bg-gray-700 hover:text-white";

  const styles = currentRoute === route ? selectedPageStyles : otherPageStyles;

  return (
    <a href={route} class={styles} aria-current="page">
      {text}
    </a>
  );
}
