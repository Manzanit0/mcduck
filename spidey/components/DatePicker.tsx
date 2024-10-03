import { JSX } from "preact";

interface DatepickerProps {
  value: string;
  onChange: (e: JSX.TargetedEvent<HTMLInputElement>) => Promise<void>;
}

export default function DatePicker(props: DatepickerProps) {
  return (
    <div>
      <div class="relative mt-2 rounded-md shadow-sm">
        <input
          type="date"
          class="block w-full rounded-md border-0 text-gray-900 ring-1 ring-inset ring-gray-300 sm:text-sm sm:leading-6 focus:outline-none focus:ring-2 focus:ring-gray-500"
          placeholder="0.00"
          value={props.value}
          onChange={props.onChange}
        />
      </div>
    </div>
  );
}

