import { JSX } from "preact";

export default function Label(props: JSX.HTMLAttributes<HTMLLabelElement>) {
  return (
    <label
      {...props}
      class="block text-sm font-medium leading-6 text-gray-900"
    />
  );
}
