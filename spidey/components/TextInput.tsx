import { JSX } from "preact";

interface TextInputProps {
  placeholder?: string
  value: string;
  onfocusout: (e: JSX.TargetedEvent<HTMLInputElement>) => Promise<void>;
}

export default function TextInput(props: TextInputProps) {
  return (
    <div>
      <div class="relative mt-2 rounded-md shadow-sm">
        <input
          type="text"
          class="block w-full rounded-md border-0 text-gray-900 ring-1 ring-inset ring-gray-300 sm:text-sm sm:leading-6 focus:outline-none focus:ring-2 focus:ring-gray-500"
          placeholder={props.placeholder}
          value={props.value}
          onfocusout={props.onfocusout}
        />
      </div>
    </div>
  );
}

